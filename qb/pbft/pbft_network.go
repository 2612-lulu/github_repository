package pbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"qb/qblock"
	"qb/qbtools"
	"qb/qbtx"
	"time"
)

// 节点共识
type NodeConsensus struct {
	Node_name            string            // 联盟节点名称
	Node_ID              [16]byte          // 联盟节点ID，16字节QKD设备号
	Node_consensus_table map[string]string // 节点索引表，key=Node_name, value=url
	Client_table         map[string]string // 客户端索引表，key=Client_name, value=url

	View      *View // 视图号
	PBFT      *Consensus
	Committed []*CommitMsg

	MsgBroadcast chan interface{} // 广播通道
	MsgEntrance  chan interface{} // 无缓冲的信息接收通道
	MsgDelivery  chan interface{} // 无缓冲的信息发送通道
	Alarm        chan bool        // 警告通道
}

// 视图号
type View struct {
	ID      int64  // 视图号
	Primary string // 主节点
}

// 共识
type Consensus struct {
	CurrentState *State     // 节点状态，默认为nil
	MsgBuffer    *MsgBuffer // 五种消息类型缓冲列表
}

// 数据缓存区
type MsgBuffer struct {
	TranscationMsgs []*qbtx.Transaction
	ReqMsgs         []*qblock.Block
	PrePrepareMsgs  []*PrePrepareMsg
	PrepareMsgs     []*PrepareMsg
	CommitMsgs      []*CommitMsg
}

// 数据处理时间限制
const ResolvingTimeDuration = time.Millisecond * 200 // 0.2 second.

const LOG_PATH = "/root/study/github_repository/qb/pbft/log/pbft_network_log/"

// NewNodeConsensus，节点共识初始化
// 参数：节点名称string
// 返回值：经初始化的节点*Node
func NewNodeConsensus(node_name string) *NodeConsensus {
	const view = 1 // 暂设视图号为1

	// 初始化节点
	node_consensus := &NodeConsensus{
		Node_name:            node_name,                                                      // 联盟节点或客户段名称，形式为P1、P2...
		Node_ID:              qbtools.GetNodeIDTable(node_name),                              // 客户端ID，16字节QKD设备号
		Node_consensus_table: qbtools.InitConfig(qbtools.INIT_PATH + "node_localhost.txt"),   // 联盟节点节点索引表，key=Node_name, value=url
		Client_table:         qbtools.InitConfig(qbtools.INIT_PATH + "client_localhost.txt"), // 客户端索引表，key=Client_name, value=url

		View: &View{ // 视图号信息，视图号=主节点下标
			ID:      view, // 视图号
			Primary: "P1", // 主节点,暂设为P1
		},

		PBFT: &Consensus{
			CurrentState: nil,
			MsgBuffer: &MsgBuffer{ // 初始化
				ReqMsgs:        make([]*qblock.Block, 0),
				PrePrepareMsgs: make([]*PrePrepareMsg, 0),
				PrepareMsgs:    make([]*PrepareMsg, 0),
				CommitMsgs:     make([]*CommitMsg, 0),
			},
		},
		Committed: make([]*CommitMsg, 0),

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgEntrance:  make(chan interface{}), // 无缓冲的信息接收通道
		MsgDelivery:  make(chan interface{}), // 无缓冲的信息发送通道
		Alarm:        make(chan bool),        // 警告通道
	}

	node_consensus.setRoute()
	// 开启线程goroutine
	go node_consensus.broadcastMsg()      // 广播信息
	go node_consensus.dispatchMsg()       // 启动消息调度器
	go node_consensus.alarmToDispatcher() // Start alarm trigger
	go node_consensus.resolveMsg()        // 开始信息表决

	return node_consensus
}

// node.setRoute,设置路由规则，在启动http服务之前设置
// 参数：无
// 返回值：无
func (node *NodeConsensus) setRoute() {
	http.HandleFunc("/preprepare", node.getPrePrepare)
	http.HandleFunc("/prepare", node.getPrepare)
	http.HandleFunc("/commit", node.getCommit)
}

// node.getPrePrepare,pre-prepare消息解码
// 参数：
// 返回值：无
func (node *NodeConsensus) getPrePrepare(writer http.ResponseWriter, request *http.Request) {
	var msg PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的消息放入通道MsgEntrance

	qbtools.Init_log(LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-pre-prepare")
	log.Println("receive a preprepare message")
}

// node.getPrepare,prepare消息解码
// 参数：
// 返回值：无
func (node *NodeConsensus) getPrepare(writer http.ResponseWriter, request *http.Request) {
	var msg PrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的prepare消息放入通道MsgEntrance

	qbtools.Init_log(LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-prepare")
	log.Println("receive a prepare message")
}

// node.getCommit,commit消息解码
// 参数：
// 返回值：无
func (node *NodeConsensus) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg CommitMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的commit消息放入通道MsgEntrance
	qbtools.Init_log(LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-commit")
	log.Println("receive a commit message")
}

// 进程1：broadcastMsg,用于广播消息
func (node *NodeConsensus) broadcastMsg() {
	for {
		msg := <-node.MsgBroadcast
		switch msg := msg.(type) {
		case *PrePrepareMsg:
			fmt.Println("------------------[START NEW PBFT]-----------------")
			qbtools.LogStage("Request", false)
			qbtools.LogStage("Request", true)
			qbtools.LogStage("Pre-Prepare", false)
			node.broadcast(msg, "/preprepare") // 发送preprepare信息给其他节点

			qbtools.Init_log(LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast preprepare]")
			log.Println("broadcast preprepare message")
		case *PrepareMsg:
			fmt.Println("------------------[START NEW PBFT]-----------------")
			qbtools.LogStage("Pre-prepare", true)
			qbtools.LogStage("Prepare", false)
			node.broadcast(msg, "/prepare") // 发送prepare信息给其他节点

			qbtools.Init_log(LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast prepare]")
			log.Println("broadcast prepare message")
		case *CommitMsg:
			qbtools.LogStage("Prepare", true)
			qbtools.LogStage("Commit", false)
			node.broadcast(msg, "/commit") // 发送commit信息给其他节点

			qbtools.Init_log(LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast commit]")
			log.Println("broadcast commit message")
		case *ReplyMsg:
			qbtools.LogStage("Commit", true)
			qbtools.LogStage("Reply", false)
			node.broadcastReply(msg, "/reply")
			qbtools.LogStage("Reply", true)

			node.PBFT.CurrentState = nil

			qbtools.Init_log(LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast reply]")
			log.Println("broadcast reply message")
		}
	}
}

// node.broadcast，节点广播函数
// 参数：待广播消息，
// 返回值：广播错误map[string]error，广播无误len(errorMap) == 0
func (node *NodeConsensus) broadcast(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果
	// 将消息广播给其他联盟节点
	for nodeID, url := range node.Node_consensus_table {
		if nodeID != node.Node_name { // 不需要向自己进行广播
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				errorMap[nodeID] = err
				continue
			}
			// 将json格式传送给其他的联盟节点
			qbtools.Send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
		} else {
			continue
		}

	}

	if len(errorMap) == 0 { // 如果转发消息均成功
		return nil
	} else { // 如果有转发失败的情况
		return errorMap
	}
}

// node.broadcastReply，节点广播函数,用于广播应答消息
// 参数：待广播消息，
// 返回值：广播错误map[string]error，广播无误len(errorMap) == 0
func (node *NodeConsensus) broadcastReply(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果

	// 将Reply消息广播给相应的客户端
	for _, transcation := range node.PBFT.CurrentState.Msg_logs.ReqMsg.Transactions {
		for _, vin := range transcation.Vin {
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				errorMap[vin.From] = err
				continue
			}
			url := node.Client_table[vin.From]
			// 将json格式发送到相应客户端
			qbtools.Send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
		}
	}

	if len(errorMap) == 0 { // 如果转发消息均成功
		return nil
	} else { // 如果有转发失败的情况
		return errorMap
	}
}

// 线程2：alarmToDispatcher，提醒处理时间片（0.2s）到
func (node *NodeConsensus) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}

// 线程3：dispatchMsg，用于处理收到的消息，一般是对信息进行暂存
func (node *NodeConsensus) dispatchMsg() {
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

// node.routeMsg，对收到的消息进行暂存处理，满足要求时发送到消息处理通道
// 参数：收到的消息
// 返回值：处理错误[]byte，默认为nil
func (node *NodeConsensus) routeMsg(msg interface{}) []error {
	switch msg := msg.(type) {
	case *qblock.Block:
		if node.PBFT.CurrentState == nil { // 如果此时不存在共识
			msgs := make([]*qblock.Block, len(node.PBFT.MsgBuffer.ReqMsgs))
			copy(msgs, node.PBFT.MsgBuffer.ReqMsgs)                // 复制缓冲数据
			msgs = append(msgs, msg)                               // 附加新到达的消息
			node.PBFT.MsgBuffer.ReqMsgs = make([]*qblock.Block, 0) // 清空重置
			node.MsgDelivery <- msgs                               // 信息发送通道：将msgs中的信息发送给MsgDelivery通道

			qbtools.Init_log(LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[block.Block]")
			log.Println("there no pbft on going, now create a new one and put the new block into MsgDelivery channel")
		} else {
			node.PBFT.MsgBuffer.ReqMsgs = append(node.PBFT.MsgBuffer.ReqMsgs, msg)

			qbtools.Init_log(LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[block.Block]")
			log.Println("exit another pbft")
		}
	// 处理PrePrepare信息
	case *PrePrepareMsg:
		if node.PBFT.CurrentState == nil { // 当CurrentState为nil时,此时不存在共识
			msgs := make([]*PrePrepareMsg, len(node.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrePrepareMsgs)                 // 复制缓冲数据
			msgs = append(msgs, msg)                                       // 附加新到达的消息
			node.PBFT.MsgBuffer.PrePrepareMsgs = make([]*PrePrepareMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                                       // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			qbtools.Init_log(LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[pbft.PrePrepareMsg]")
			log.Println("[put pre-prepare message into MsgDelivery channel]")
		} else { // 当CurrentState不为nil时，直接往MsgBuffer缓冲通道中进行添加
			node.PBFT.MsgBuffer.PrePrepareMsgs = append(node.PBFT.MsgBuffer.PrePrepareMsgs, msg)
			qbtools.Init_log(LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[pbft.PrePrepareMsg]")
			log.Println("[get a pre-prepare message, but don't put it into channel]")
		}
	// 处理Prepare信息
	case *PrepareMsg:
		if node.PBFT.CurrentState == nil || node.PBFT.CurrentState.Current_stage != PrePrepared {
			node.PBFT.MsgBuffer.PrepareMsgs = append(node.PBFT.MsgBuffer.PrepareMsgs, msg)
			qbtools.Init_log(LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[pbft.PrepareMsg]")
			log.Println("[get a prepare message,but don't put it into channel]")
		} else {
			msgs := make([]*PrepareMsg, len(node.PBFT.MsgBuffer.PrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrepareMsgs)              // 复制缓冲数据
			msgs = append(msgs, msg)                                 // 附加新到达的消息
			node.PBFT.MsgBuffer.PrepareMsgs = make([]*PrepareMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                                 // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			qbtools.Init_log("./pbft/pbft_network_log/dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[pbft.PrepareMsg]")
			log.Println("[put prepare message into MsgDelivery channel]")
		}
	// 处理CommitMsg信息
	case *CommitMsg:
		if node.PBFT.CurrentState == nil || node.PBFT.CurrentState.Current_stage != Prepared {
			node.PBFT.MsgBuffer.CommitMsgs = append(node.PBFT.MsgBuffer.CommitMsgs, msg)
			qbtools.Init_log(LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[pbft.CommitMsg]")
			log.Println("[get a commit message,but don't put it into channel]")
		} else {
			msgs := make([]*CommitMsg, len(node.PBFT.MsgBuffer.CommitMsgs))
			copy(msgs, node.PBFT.MsgBuffer.CommitMsgs)             // 复制缓冲数据
			msgs = append(msgs, msg)                               // 附加新到达的消息
			node.PBFT.MsgBuffer.CommitMsgs = make([]*CommitMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                               // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			qbtools.Init_log(LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[pbft.CommitMsg]")
			log.Println("[put commit message into MsgDelivery channel]")
		}
	}
	return nil
}

// node.routeMsgWhenAlarmed,当时间片到时，对收到的消息进行暂存处理，满足要求时发送到消息处理通道
func (node *NodeConsensus) routeMsgWhenAlarmed() []error {
	if node.PBFT.CurrentState == nil {
		// 检查PrePrepareMsgs, 并发送到MsgDelivery.
		if len(node.PBFT.MsgBuffer.PrePrepareMsgs) != 0 {
			msgs := make([]*PrePrepareMsg, len(node.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrePrepareMsgs)
			node.PBFT.MsgBuffer.PrePrepareMsgs = make([]*PrePrepareMsg, 0)
			node.MsgDelivery <- msgs
		}
	} else {
		switch node.PBFT.CurrentState.Current_stage {
		case PrePrepared:
			// 检查PrepareMsgs,并发送到MsgDelivery.
			if len(node.PBFT.MsgBuffer.PrepareMsgs) != 0 {
				msgs := make([]*PrepareMsg, len(node.PBFT.MsgBuffer.PrepareMsgs))
				copy(msgs, node.PBFT.MsgBuffer.PrepareMsgs)
				node.PBFT.MsgBuffer.PrepareMsgs = make([]*PrepareMsg, 0) // 清空重置
				node.MsgDelivery <- msgs
			}
		case Prepared:
			// 检查CommitMsgs,并发送到MsgDelivery.
			if len(node.PBFT.MsgBuffer.CommitMsgs) != 0 {
				msgs := make([]*CommitMsg, len(node.PBFT.MsgBuffer.CommitMsgs))
				copy(msgs, node.PBFT.MsgBuffer.CommitMsgs)
				node.PBFT.MsgBuffer.CommitMsgs = make([]*CommitMsg, 0) // 清空重置
				node.MsgDelivery <- msgs
			}
		}
	}
	return nil
}

// 线程4：resolveMsg，用于对收到的信息作具体处理
func (node *NodeConsensus) resolveMsg() {
	for {
		msgs := <-node.MsgDelivery // 从调度器通道中获取缓存信息
		switch msgs := msgs.(type) {
		// 节点表决决策信息
		case []*qblock.Block:
			if node.PBFT.CurrentState != nil {
				qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[block.Block]")
				log.Println("get a block,but another pbft consensus is ongoing")
			} else {
				qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[receive block.Block]")
				log.Println("Msgdelivery channel get a new block,and create a new consensus")

				err := node.resolveRequestMsg(msgs)
				if err != nil {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		case []*PrePrepareMsg:
			if node.PBFT.CurrentState != nil {
				qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[receive PrePrepareMsg]")
				log.Println("get a preprepare,but another pbft consensus is ongoing")
			} else {
				qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[receive PrePrepareMsg]")
				log.Println("Msgdelivery channel get a preprepare, and handle it")

				errs := node.resolvePrePrepareMsg(msgs)
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err) // TODO: send err to ErrorChannel
					}
				}
			}
		case []*PrepareMsg:
			qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[receive PrepareMsg]")
			log.Println("Msgdelivery channel get a prepare message, and handle it")

			errs := node.resolvePrepareMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}

		case []*CommitMsg:
			qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[receive CommitMsg]")
			log.Println("Msgdelivery channel get a commit message, and handle it")

			errs := node.resolveCommitMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		}
	}
}

// node.createStateForNewConsensus，创建新的共识
// 参数：
// 返回值：创建错误error，默认为nil
func (node *NodeConsensus) createStateForNewConsensus() error {
	if node.PBFT.CurrentState != nil { // 判断当前节点是不是处于其他阶段（预准备阶段或者准备阶段等等）
		return errors.New("another pbft consensus is ongoing") // 如果有，则输出提示
	}
	var lastSequenceID int64 // 获取上一个序列号
	// 判断当前阶段是否已经发送过消息
	if len(node.Committed) == 0 { // 如果是首次进行共识，则上一个序列号lastSequenceID设置为-1
		lastSequenceID = -1
	} else { // 否则取出上一个序列号
		lastSequenceID = node.Committed[len(node.Committed)-1].Sequence_number
	}
	// 创建新的节点状态，即进行节点状态的初始化
	node.PBFT.CurrentState = CreateState(node.View.ID, lastSequenceID)
	return nil
}

// node.resolveRequestMsg,处理收到的区块数组
// 参数：区块数组[]*block.Block
// 返回值：处理错误[]error
func (node *NodeConsensus) resolveRequestMsg(msgs []*qblock.Block) []error {
	errs := make([]error, 0)

	err := node.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
	if err != nil {
		fmt.Println(err)
	}
	// 批量处理req信息
	for _, req := range msgs {
		err := node.resolveRequest(req)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 { // 如果有处理错误，则输出错误
		return errs
	}
	return nil
}

// node.resolveRequestMsg,处理单条区块，即请求消息，生成预准备消息
// 参数：区块*block.Block
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolveRequest(msgs *qblock.Block) error {
	prePrepareMsg, err := node.PBFT.CurrentState.PrePrePare(msgs) // 进入共识，获得preprepare消息
	if err != nil {
		return err
	} else {
		qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[get pre-prepareMsg]")
		log.Println("put pre-prepare message into broadcast channel")

		node.MsgBroadcast <- prePrepareMsg // 将待广播消息放入通道
		return nil
	}
}

// node.resolvePrePrepareMsg，由从节点处理PrePrepare消息数组
// 参数：预准备消息数组[]*pbft.PrePrepareMsg
// 返回值：处理错误[]error
func (node *NodeConsensus) resolvePrePrepareMsg(msgs []*PrePrepareMsg) []error {
	errs := make([]error, 0)

	err := node.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
	if err != nil {
		fmt.Println(err)
	}
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

// node.resolvePrePrepare,处理单条预准备消息，生成准备消息
// 参数：预准备消息*pbft.PrePrepareMsg
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolvePrePrepare(prePrepareMsg *PrePrepareMsg) error {
	prePareMsg, err := node.PBFT.CurrentState.PrePare(prePrepareMsg) // 获得prepare信息
	if err != nil {
		return err
	}
	if prePareMsg != nil {
		qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[get prepareMsg]")
		log.Println("put prepare message into broadcast channel")

		node.MsgBroadcast <- prePareMsg // 将待广播消息放入通道
	}
	return nil
}

// node.resolvePrepareMsg,处理准备消息数组，得到提交消息
// 参数：准备消息数组[]*pbft.PrepareMsg
// 返回值：处理错误[]error
func (node *NodeConsensus) resolvePrepareMsg(msgs []*PrepareMsg) []error {
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

// node.resolvePrepare,处理单条准备消息，生成提交消息
// 参数：准备消息*pbft.PrepareMsg
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolvePrepare(prepareMsg *PrepareMsg) error {
	commitMsg, err := node.PBFT.CurrentState.Commit(prepareMsg)
	if err != nil {
		return err
	}
	if commitMsg != nil {
		qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[get commitMsg]")
		log.Println("put commit message into broadcast channel")

		node.MsgBroadcast <- commitMsg // 将待广播消息放入通道
	}
	return nil
}

// node.resolveCommitMsg,所有节点处理接收的commit消息数组得到reply消息
// 参数：提交消息数组[]*pbft.CommitMsg
// 返回值：处理错误[]error
func (node *NodeConsensus) resolveCommitMsg(msgs []*CommitMsg) []error {
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

// node.resolveCommit,处理单条提交消息，生成应答消息
// 参数：准备消息*pbft.CommitMsg
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolveCommit(commitMsg *CommitMsg) error {
	replyMsg, err := node.PBFT.CurrentState.Reply(commitMsg)
	if err != nil {
		return err
	}
	if replyMsg != nil {
		qbtools.Init_log(LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[get replyMsg]")
		log.Println("put reply message into broadcast channel")

		node.Committed = append(node.Committed, commitMsg)
		node.MsgBroadcast <- replyMsg // 将待广播消息放入通道
	}
	return nil
}
