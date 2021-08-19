package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"qb/pbft"
	"qb/qkdserv"
	"strconv"
	"time"
)

type Node struct {
	Node_name     [2]byte            // 联盟节点/客户端名称
	NodeID        [16]byte           // 联盟节点/客户端ID，16字节QKD设备号
	NodeTable     map[[2]byte]string // 节点索引表，key=Node_name, value=url
	ClientTable   map[[2]byte]string // 客户端索引表，key=Node_name, value=url
	View          *View              // 视图号
	CurrentState  *pbft.State        // 节点状态，默认为nil
	CommittedMsgs []*pbft.CommitMsg  // 被提交的信息????
	MsgBuffer     *MsgBuffer         // 五种消息类型缓冲列表
	MsgEntrance   chan interface{}   // 无缓冲的信息接收通道
	MsgDelivery   chan interface{}   // 无缓冲的信息发送通道
	Alarm         chan bool          // 警告通道
}

type MsgBuffer struct {
	ReqMsgs        []*pbft.RequestMsg
	ReplyMsgs      []*pbft.ReplyMsg
	PrePrepareMsgs []*pbft.PrePrepareMsg
	PrepareMsgs    []*pbft.PrepareMsg
	CommitMsgs     []*pbft.CommitMsg
}

type View struct {
	ID      int64   // 视图号
	Primary [2]byte // 主节点
}

const ResolvingTimeDuration = time.Millisecond * 1000 // 1 second.

// 节点初始化
func NewNode(node_name string) *Node {
	const view = 1 // 暂设视图号为1

	// 处理输入参数node_name
	var NodeName [2]byte
	for i := 0; i < 2; i++ {
		NodeName[i] = []byte(node_name)[i]
	}

	qkdserv.Node_name = NodeName // 调用此程序的当前节点或客户端名称

	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)

	// 初始化节点
	node := &Node{
		Node_name: NodeName, // 联盟节点或客户段名称，形式为P1、P2...或C1、C2...
		NodeTable: map[[2]byte]string{ // 节点索引表，key=Node_name, value=url
			{'P', '1'}: "localhost:1111",
			{'P', '2'}: "localhost:1112",
			{'P', '3'}: "localhost:1113",
			{'P', '4'}: "localhost:1114",
		},
		ClientTable: map[[2]byte]string{ // 客户端索引表，key=Node_name, value=url
			{'C', '1'}: "localhost:2111",
			{'C', '2'}: "localhost:2112",
			{'C', '3'}: "localhost:2113",
			{'C', '4'}: "localhost:2114",
		},
		View: &View{ // 视图号信息，视图号=主节点下标
			ID:      view,              // 视图号
			Primary: [2]byte{'P', '1'}, // 主节点,暂设为P1
		},

		// Consensus-related struct
		CurrentState:  nil,                        // 节点当前状态，默认是nil
		CommittedMsgs: make([]*pbft.CommitMsg, 0), // 被提交的信息
		MsgBuffer: &MsgBuffer{ // 初始化
			ReqMsgs:        make([]*pbft.RequestMsg, 0),
			ReplyMsgs:      make([]*pbft.ReplyMsg, 0),
			PrePrepareMsgs: make([]*pbft.PrePrepareMsg, 0),
			PrepareMsgs:    make([]*pbft.PrepareMsg, 0),
			CommitMsgs:     make([]*pbft.CommitMsg, 0),
		},

		// Channels
		MsgEntrance: make(chan interface{}), // 无缓冲的信息接收通道
		MsgDelivery: make(chan interface{}), // 无缓冲的信息发送通道
		Alarm:       make(chan bool),        // 警告通道
	}

	// 启动消息调度器
	go node.dispatchMsg()

	// Start alarm trigger
	go node.alarmToDispatcher()

	// 开始信息表决
	go node.resolveMsg()

	return node
}

// Broadcast，节点广播函数
func (node *Node) Broadcast(msg interface{}, path string) map[[2]byte]error {
	errorMap := make(map[[2]byte]error) // 存放广播结果

	// 将消息广播给其他联盟节点
	for nodeID, url := range node.NodeTable {
		// 因为不需要向自己进行广播了，所以就直接跳过
		if nodeID == node.Node_name {
			continue
		}
		// 将msg信息编码成json格式
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			errorMap[nodeID] = err
			continue
		}
		// 将json格式传送给其他的联盟节点
		send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
	}

	if len(errorMap) == 0 { // 如果转发消息均成功
		return nil
	} else { // 如果有转发失败的情况
		return errorMap
	}
}

// 协程1：dispatchMsg
func (node *Node) dispatchMsg() {
	for {
		select {
		case msg := <-node.MsgEntrance: // 信息接收通道：如果MsgEntrance通道有消息传送过来，拿到msg
			err := node.routeMsg(msg) // 进行routeMsg
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		case <-node.Alarm:
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		}
	}
}

func (node *Node) routeMsg(msg interface{}) []error {
	switch msg := msg.(type) {
	// 处理RequestMsg信息
	case *pbft.RequestMsg:
		if node.CurrentState == nil { // 当CurrentState为nil时
			msgs := make([]*pbft.RequestMsg, len(node.MsgBuffer.ReqMsgs))
			copy(msgs, node.MsgBuffer.ReqMsgs) // 复制缓冲数据

			msgs = append(msgs, msg) // 附加新到达的消息

			// 清空重置
			node.MsgBuffer.ReqMsgs = make([]*pbft.RequestMsg, 0)

			node.MsgDelivery <- msgs // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		} else { // 当CurrentState不为nil时，直接往MsgBuffer缓冲通道中进行添加，会涉及到数组扩容问题
			node.MsgBuffer.ReqMsgs = append(node.MsgBuffer.ReqMsgs, msg)
		}
	// 处理PrePrepare信息
	case *pbft.PrePrepareMsg:
		if node.CurrentState == nil { // 当CurrentState为nil时
			msgs := make([]*pbft.PrePrepareMsg, len(node.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.MsgBuffer.PrePrepareMsgs)                      // 复制缓冲数据
			msgs = append(msgs, msg)                                       // 附加新到达的消息
			node.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0) // 清空重置

			node.MsgDelivery <- msgs // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		} else { // 当CurrentState不为nil时，直接往MsgBuffer缓冲通道中进行添加
			node.MsgBuffer.PrePrepareMsgs = append(node.MsgBuffer.PrePrepareMsgs, msg)
		}
	// 处理Prepare信息
	case *pbft.PrepareMsg:
		if node.CurrentState == nil || node.CurrentState.CurrentStage != pbft.PrePrepared {
			node.MsgBuffer.PrepareMsgs = append(node.MsgBuffer.PrepareMsgs, msg)
		} else {
			msgs := make([]*pbft.PrepareMsg, len(node.MsgBuffer.PrepareMsgs))
			copy(msgs, node.MsgBuffer.PrepareMsgs)                   // 复制缓冲数据
			msgs = append(msgs, msg)                                 // 附加新到达的消息
			node.MsgBuffer.PrepareMsgs = make([]*pbft.PrepareMsg, 0) // 清空重置

			node.MsgDelivery <- msgs // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		}
	// 处理CommitMsg信息
	case *pbft.CommitMsg:
		if node.CurrentState == nil || node.CurrentState.CurrentStage != pbft.Prepared {
			node.MsgBuffer.CommitMsgs = append(node.MsgBuffer.CommitMsgs, msg)
		} else {
			msgs := make([]*pbft.CommitMsg, len(node.MsgBuffer.CommitMsgs))
			copy(msgs, node.MsgBuffer.CommitMsgs)                  // 复制缓冲数据
			msgs = append(msgs, msg)                               // 附加新到达的消息
			node.MsgBuffer.CommitMsgs = make([]*pbft.CommitMsg, 0) // 清空重置

			node.MsgDelivery <- msgs // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		}

	// 处理ReplyMsg信息
	case *pbft.ReplyMsg:
		if node.CurrentState == nil { // 当CurrentState为nil时
			msgs := make([]*pbft.ReplyMsg, len(node.MsgBuffer.ReplyMsgs))
			copy(msgs, node.MsgBuffer.ReplyMsgs)                 // 复制缓冲数据
			msgs = append(msgs, msg)                             // 附加新到达的消息
			node.MsgBuffer.ReplyMsgs = make([]*pbft.ReplyMsg, 0) // 清空重置

			node.MsgDelivery <- msgs // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		} else { // 当CurrentState不为nil时，直接往MsgBuffer缓冲通道中进行添加
			node.MsgBuffer.ReplyMsgs = append(node.MsgBuffer.ReplyMsgs, msg)
		}
	}
	return nil
}

func (node *Node) routeMsgWhenAlarmed() []error {
	if node.CurrentState == nil {
		// 检查ReqMsgs, 并发送到MsgDelivery.
		if len(node.MsgBuffer.ReqMsgs) != 0 {
			msgs := make([]*pbft.RequestMsg, len(node.MsgBuffer.ReqMsgs))
			copy(msgs, node.MsgBuffer.ReqMsgs)

			node.MsgDelivery <- msgs
		}

		// 检查PrePrepareMsgs, 并发送到MsgDelivery.
		if len(node.MsgBuffer.PrePrepareMsgs) != 0 {
			msgs := make([]*pbft.PrePrepareMsg, len(node.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.MsgBuffer.PrePrepareMsgs)

			node.MsgDelivery <- msgs
		}
	} else {
		switch node.CurrentState.CurrentStage {
		case pbft.PrePrepared:
			// 检查PrepareMsgs,并发送到MsgDelivery.
			if len(node.MsgBuffer.PrepareMsgs) != 0 {
				msgs := make([]*pbft.PrepareMsg, len(node.MsgBuffer.PrepareMsgs))
				copy(msgs, node.MsgBuffer.PrepareMsgs)

				node.MsgDelivery <- msgs
			}
		case pbft.Prepared:
			// 检查CommitMsgs,并发送到MsgDelivery.
			if len(node.MsgBuffer.CommitMsgs) != 0 {
				msgs := make([]*pbft.CommitMsg, len(node.MsgBuffer.CommitMsgs))
				copy(msgs, node.MsgBuffer.CommitMsgs)

				node.MsgDelivery <- msgs
			}
		}
	}

	return nil
}

// 协程2：resolveMsg
func (node *Node) resolveMsg() {
	for {
		msgs := <-node.MsgDelivery // 从调度器中获取缓存信息
		switch msgs := msgs.(type) {
		// 节点表决决策信息
		case []*pbft.RequestMsg:
			errs := node.resolveRequestMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		case []*pbft.PrePrepareMsg:
			errs := node.resolvePrePrepareMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		case []*pbft.PrepareMsg:
			errs := node.resolvePrepareMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}

			}
		case []*pbft.CommitMsg:
			errs := node.resolveCommitMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
			/*case []*pbft.ReplyMsg:
			errs := node.resolveReplyMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}*/
		}
	}
}

// 协程3，alarmToDispatcher，警告信息
func (node *Node) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}

// createStateForNewConsensus，创建新的共识
func (node *Node) createStateForNewConsensus() error {
	if node.CurrentState != nil { // 判断当前节点是不是处于其他阶段（预准备阶段或者准备阶段等等）
		return errors.New("another pbft consensus is ongoing") // 如果有，则输出提示
	}

	var lastSequenceID int64 // 获取上一个序列号
	// 判断当前阶段是否已经发送过消息
	if len(node.CommittedMsgs) == 0 { // 如果是首次进行共识，则上一个序列号lastSequenceID设置为-1
		lastSequenceID = -1
	} else { // 否则取出上一个序列号
		lastSequenceID = node.CommittedMsgs[len(node.CommittedMsgs)-1].Sequence_number
	}
	// 创建新的节点状态，即进行节点状态的初始化
	node.CurrentState = pbft.CreateState(node.View.ID, lastSequenceID)
	//LogStage("Create the replica status", true)
	return nil
}

// 只有客户端可调用此函数，用于生成request消息并将该消息发送至主节点以请求共识
func (node *Node) Request(operation string, node_name [2]byte) error {
	err := node.createStateForNewConsensus() // 创建新的共识
	if err != nil {                          // 如果节点未处于共识状态，输出错误
		return err
	}
	request, _ := node.CurrentState.GenReqMsg(operation, node_name)
	jsonMsg, err := json.Marshal(request)
	if err != nil {
		return err
	}
	send(node.NodeTable[node.View.Primary]+"/request", jsonMsg)
	fmt.Println(" The request have send to primary node")
	return nil

}

func (node *Node) resolveRequestMsg(msgs []*pbft.RequestMsg) []error {
	errs := make([]error, 0)
	// 批量处理request信息
	for _, reqMsg := range msgs {
		err := node.resolveReq(reqMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 { // 如果有处理错误，则输出错误
		return errs
	}

	return nil
}

// 进入共识,由主节点处理request消息
func (node *Node) resolveReq(reqMsg *pbft.RequestMsg) error {
	//LogMsg(reqMsg)
	err := node.createStateForNewConsensus() // 创建新的共识
	if err != nil {                          // 如果节点未处于共识状态，输出错误
		return err
	}

	// 进入共识，获得preprepare消息
	prePrepareMsg, ok := node.CurrentState.PrePrePare(reqMsg)

	//LogStage(fmt.Sprintf("Consensus Process (ViewID:%d)", node.CurrentState.ViewID), false)

	// 发送pre-prepare消息给其他联盟节点
	if ok {
		node.Broadcast(prePrepareMsg, "/preprepare")
		fmt.Println(" received request, and have sent preprepare message")
		//LogStage("Pre-prepare", true)
	}

	return nil
}

// 由从节点处理PrePrepare消息
func (node *Node) resolvePrePrepareMsg(msgs []*pbft.PrePrepareMsg) []error {
	errs := make([]error, 0)

	// 批量处理pre-prepare信息
	for _, prePrepareMsg := range msgs {
		err := node.resolvePrePrepare(prePrepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 { // 如果有处理错误，则输出错误
		return errs
	}

	return nil
}

func (node *Node) resolvePrePrepare(prePrepareMsg *pbft.PrePrepareMsg) error {
	// LogMsg(prePrepareMsg)
	err := node.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
	if err != nil {
		return err
	}

	// 处理预准备信息，获得prepare信息
	prePareMsg, ok := node.CurrentState.PrePare(prePrepareMsg)
	if ok {
		// 添加当前节点编号
		prePareMsg.Node_i, _ = strconv.ParseInt(string(node.Node_name[1]), 10, 64)
		//LogStage("Pre-prepare", true)
		node.Broadcast(prePareMsg, "/prepare") // 发送prepare信息给其他节点
		fmt.Println(" received pre-prepare message, and have sent prepare message")
		//LogStage("Prepare", false)
	}

	return nil
}

// 所有联盟节点接收prepare消息，处理得到commit消息
func (node *Node) resolvePrepareMsg(msgs []*pbft.PrepareMsg) []error {
	errs := make([]error, 0)

	// 批量处理准备消息
	for _, prepareMsg := range msgs {
		err := node.resolvePrepare(prepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolvePrepare(prepareMsg *pbft.PrepareMsg) error {
	//LogMsg(prepareMsg)

	commitMsg := pbft.CommitMsg{}
	if node.CurrentState.VerifyPrepareMsg(prepareMsg) {
		commitMsg = *node.CurrentState.GetCommitMsg(prepareMsg)
		commitMsg.Node_i, _ = strconv.ParseInt(string(node.Node_name[1]), 10, 64)
		//LogStage("Prepare", true)
		node.Broadcast(commitMsg, "/commit")
		fmt.Println(" received prepare message, and have sent commit message")
		//LogStage("Commit", false)
	}
	return nil

}

func (node *Node) resolveCommitMsg(msgs []*pbft.CommitMsg) []error {
	errs := make([]error, 0)

	// 批量处理commit信息
	for _, commitMsg := range msgs {
		err := node.resolveCommit(commitMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}
func (node *Node) resolveCommit(commitMsg *pbft.CommitMsg) error {
	//LogMsg(commitMsg)
	replyMsg := pbft.ReplyMsg{}
	if node.CurrentState.VerifyCommitMsg(commitMsg) {
		replyMsg = *node.CurrentState.GetReplyMsg(commitMsg)
		replyMsg.Node_i, _ = strconv.ParseInt(string(node.Node_name[1]), 10, 64)

		// Save the last version of committed messages to node.????
		node.CommittedMsgs = append(node.CommittedMsgs, commitMsg)

		//LogStage("Commit", true)
		node.Reply(&replyMsg)
		//LogStage("Reply", true)
	}

	return nil
}

// 各联盟节点发送reply消息给客户端
func (node *Node) Reply(msg *pbft.ReplyMsg) error {

	/*for _, value := range node.CommittedMsgs {
		fmt.Printf("Committed value: %s, %d, %s, %d", value.ClientID, value.Timestamp, value.Operation, value.SequenceID)
	}
	fmt.Print("\n")*/

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	send(node.ClientTable[node.CurrentState.Msg_logs.ReqMsg.Sign_client.Main_row_num.Sign_Node_Name]+"/reply", jsonMsg)

	return nil
}

func (node *Node) GetReply(msg *pbft.ReplyMsg) {
	fmt.Print("Result:", msg.Result)
	fmt.Print("by node ", msg.Result)
}
