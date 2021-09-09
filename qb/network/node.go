package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"qb/block"
	"qb/mylog"
	"qb/pbft"
	"qb/qbtools"
	"qb/qkdserv"
)

type Node struct {
	Node_name    string            // 联盟节点名称
	Node_ID      [16]byte          // 联盟节点ID，16字节QKD设备号
	Node_table   map[string]string // 节点索引表，key=Node_name, value=url
	Client_table map[string]string // 客户端索引表，key=Client_name, value=url

	View      *View // 视图号
	PBFT      *Consensus
	Committed []*pbft.CommitMsg

	MsgBroadcast chan interface{} // 广播通道
	MsgBlock     chan interface{} // 打包通道
	MsgEntrance  chan interface{} // 无缓冲的信息接收通道
	MsgDelivery  chan interface{} // 无缓冲的信息发送通道
	Alarm        chan bool        // 警告通道
}

type View struct {
	ID      int64  // 视图号
	Primary string // 主节点
}
type Consensus struct {
	CurrentState *pbft.State // 节点状态，默认为nil
	MsgBuffer    *MsgBuffer  // 五种消息类型缓冲列表
}

type MsgBuffer struct {
	TranscationMsgs []*block.Transaction
	ReqMsgs         []*block.Block
	ReplyMsgs       []*pbft.ReplyMsg
	PrePrepareMsgs  []*pbft.PrePrepareMsg
	PrepareMsgs     []*pbft.PrepareMsg
	CommitMsgs      []*pbft.CommitMsg
}

const ResolvingTimeDuration = time.Millisecond * 1000 // 1 second.

// 节点初始化
func NewNode(node_name string) *Node {
	const view = 1                // 暂设视图号为1
	qkdserv.Node_name = node_name // 调用此程序的当前节点或客户端名称
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)

	// 初始化节点
	node := &Node{
		Node_name:    node_name,                                                   // 联盟节点或客户段名称，形式为P1、P2...
		Node_ID:      qbtools.GetNodeIDTable(node_name),                           // 客户端ID，16字节QKD设备号
		Node_table:   qbtools.InitConfig("./qbtools/config/node_localhost.txt"),   // 联盟节点节点索引表，key=Node_name, value=url
		Client_table: qbtools.InitConfig("./qbtools/config/client_localhost.txt"), // 客户端索引表，key=Client_name, value=url

		View: &View{ // 视图号信息，视图号=主节点下标
			ID:      view, // 视图号
			Primary: "P1", // 主节点,暂设为P1
		},

		PBFT: &Consensus{
			CurrentState: nil,
			MsgBuffer: &MsgBuffer{ // 初始化
				ReqMsgs:        make([]*block.Block, 0),
				ReplyMsgs:      make([]*pbft.ReplyMsg, 0),
				PrePrepareMsgs: make([]*pbft.PrePrepareMsg, 0),
				PrepareMsgs:    make([]*pbft.PrepareMsg, 0),
				CommitMsgs:     make([]*pbft.CommitMsg, 0),
			},
		},
		Committed: make([]*pbft.CommitMsg, 0),

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgBlock:     make(chan interface{}), // 交易信息打包通道
		MsgEntrance:  make(chan interface{}), // 无缓冲的信息接收通道
		MsgDelivery:  make(chan interface{}), // 无缓冲的信息发送通道
		Alarm:        make(chan bool),        // 警告通道
	}

	node.setRoute()
	// 开启线程goroutine
	go node.broadcastMsg()      // 广播信息
	go node.dispatchMsg()       // 启动消息调度器
	go node.alarmToDispatcher() // Start alarm trigger
	go node.blockMsg()
	go node.resolveMsg() // 开始信息表决

	node.httplisten() // 开启http
	return node
}

func (node *Node) NewConsensus() *Consensus {
	pbft_consensus := &Consensus{
		CurrentState: nil, // 节点当前状态，默认是nil
		MsgBuffer: &MsgBuffer{ // 初始化
			ReqMsgs:        make([]*block.Block, 0),
			ReplyMsgs:      make([]*pbft.ReplyMsg, 0),
			PrePrepareMsgs: make([]*pbft.PrePrepareMsg, 0),
			PrepareMsgs:    make([]*pbft.PrepareMsg, 0),
			CommitMsgs:     make([]*pbft.CommitMsg, 0),
		},
	}
	return pbft_consensus
}

func init_log(path string) error {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644) //【如果已经存在，则在尾部添加写】
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return err
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	return nil
}

// setRoute,设置路由规则，在启动http服务之前设置
func (node *Node) setRoute() {
	http.HandleFunc("/transcation", node.getTranscation)
	http.HandleFunc("/preprepare", node.getPrePrepare)
	http.HandleFunc("/prepare", node.getPrepare)
	http.HandleFunc("/commit", node.getCommit)
}

// node.getTranscation,交易信息解码
func (node *Node) getTranscation(writer http.ResponseWriter, request *http.Request) {
	var msg block.Transaction
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的交易消息放入打包通道
	init_log("./network/nodelog/listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-transcation")
	log.Println("receive a transcation message")
}

// node.getPrePrepare,pre-prepare消息解码
func (node *Node) getPrePrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的消息放入通道MsgEntrance
	init_log("./network/nodelog/listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-pre-prepare")
	log.Println("receive a preprepare message")
}

// node.getPrepare,prepare消息解码
func (node *Node) getPrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的prepare消息放入通道MsgEntrance
	init_log("./network/nodelog/listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-prepare")
	log.Println("receive a prepare message")
}

// node.getCommit,commit消息解码
func (node *Node) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.CommitMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的commit消息放入通道MsgEntrance
	init_log("./network/nodelog/listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-commit")
	log.Println("receive a commit message")
}

// node.httplisten，开启Http服务器
func (node *Node) httplisten() {
	url := node.Node_table[node.Node_name]
	fmt.Printf("Server will be started at %s...\n", url)
	if err := http.ListenAndServe(url, nil); err != nil {
		fmt.Println(err)
		return
	}
}

// 进程1：BroadcastMsg
func (node *Node) broadcastMsg() {
	for {
		msg := <-node.MsgBroadcast
		switch msg := msg.(type) {
		case *pbft.PrePrepareMsg:
			mylog.LogStage("Request", true)
			node.broadcast(msg, "/preprepare") // 发送preprepare信息给其他节点
			mylog.LogStage("Pre-Prepare", false)

			init_log("./network/nodelog/broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast preprepare]")
			log.Println("broadcast preprepare message")
		case *pbft.PrepareMsg:
			mylog.LogStage("Pre-prepare", true)
			node.broadcast(msg, "/prepare") // 发送prepare信息给其他节点
			mylog.LogStage("Prepare", false)

			init_log("./network/nodelog/broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast prepare]")
			log.Println("broadcast prepare message")
		case *pbft.CommitMsg:
			mylog.LogStage("Prepare", true)
			node.broadcast(msg, "/commit") // 发送commit信息给其他节点
			mylog.LogStage("Commit", false)

			init_log("./network/nodelog/broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast commit]")
			log.Println("broadcast commit message")
		case *pbft.ReplyMsg:
			mylog.LogStage("Commit", true)
			node.broadcastReply(msg, "/reply")
			mylog.LogStage("Reply", false)

			init_log("./network/nodelog/broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast reply]")
			log.Println("broadcast reply message")

			node.PBFT.CurrentState = nil
		}
	}
}

// broadcast，节点广播函数
func (node *Node) broadcast(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果
	// 将消息广播给其他联盟节点
	for nodeID, url := range node.Node_table {
		if nodeID != node.Node_name { // 不需要向自己进行广播
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				errorMap[nodeID] = err
				continue
			}
			// 将json格式传送给其他的联盟节点
			send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
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

// broadcastReply，节点广播函数
func (node *Node) broadcastReply(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果

	// 将Reply消息广播给相应的客户端
	for _, transcation := range node.PBFT.CurrentState.Msg_logs.ReqMsg.Transactions {
		jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
		if err != nil {
			errorMap[transcation.Name] = err
			continue
		}
		url := node.Client_table[transcation.Name]
		// 将json格式发送到相应客户端
		send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
	}

	if len(errorMap) == 0 { // 如果转发消息均成功
		return nil
	} else { // 如果有转发失败的情况
		return errorMap
	}
}

// 线程2：dispatchMsg
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
	init_log("./network/nodelog/dispatch_" + node.Node_name + ".log")
	switch msg := msg.(type) {
	case *block.Transaction:
		if len(node.PBFT.MsgBuffer.TranscationMsgs) >= block.Block_Length-1 { // 当区块长度满足要求时
			msgs := make([]*block.Transaction, len(node.PBFT.MsgBuffer.TranscationMsgs))
			copy(msgs, node.PBFT.MsgBuffer.TranscationMsgs) // 复制缓冲数据
			msgs = append(msgs, msg)                        // 附加新到达的消息

			request := block.CreateBlock(msgs)
			node.PBFT.MsgBuffer.TranscationMsgs = make([]*block.Transaction, 0) // 清空重置
			node.MsgBlock <- request
			log.SetPrefix(node.Node_name + "-[block.Transaction]")
			log.Println("create a new block, and put it into MsgEntrance channel")
		} else { // 数量不够打包时，直接往MsgBuffer缓冲通道中进行添加
			node.PBFT.MsgBuffer.TranscationMsgs = append(node.PBFT.MsgBuffer.TranscationMsgs, msg)
			log.SetPrefix(node.Node_name + "-[block.Transaction]")
			log.Println("get a transcation message")
		}
	// 处理PrePrepare信息
	case *pbft.PrePrepareMsg:
		if node.PBFT.CurrentState == nil { // 当CurrentState为nil时
			msgs := make([]*pbft.PrePrepareMsg, len(node.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrePrepareMsgs)                      // 复制缓冲数据
			msgs = append(msgs, msg)                                            // 附加新到达的消息
			node.PBFT.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0) // 清空重置
			node.MsgDelivery <- msgs
			// 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			log.SetPrefix(node.Node_name + "-[pbft.PrePrepareMsg]")
			log.Println("[put pre-prepare message into MsgDelivery channel]")
		} else { // 当CurrentState不为nil时，直接往MsgBuffer缓冲通道中进行添加
			node.PBFT.MsgBuffer.PrePrepareMsgs = append(node.PBFT.MsgBuffer.PrePrepareMsgs, msg)
			log.SetPrefix(node.Node_name + "-[pbft.PrePrepareMsg]")
			log.Println("[get a pre-prepare message, but don't put it into channel]")
		}
	// 处理Prepare信息
	case *pbft.PrepareMsg:
		if node.PBFT.CurrentState == nil || node.PBFT.CurrentState.Current_stage != pbft.PrePrepared {
			node.PBFT.MsgBuffer.PrepareMsgs = append(node.PBFT.MsgBuffer.PrepareMsgs, msg)
			log.SetPrefix(node.Node_name + "-[pbft.PrepareMsg]")
			log.Println("[get a prepare message,but don't put it into channel]")
		} else {
			msgs := make([]*pbft.PrepareMsg, len(node.PBFT.MsgBuffer.PrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrepareMsgs)                   // 复制缓冲数据
			msgs = append(msgs, msg)                                      // 附加新到达的消息
			node.PBFT.MsgBuffer.PrepareMsgs = make([]*pbft.PrepareMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                                      // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			log.SetPrefix(node.Node_name + "-[pbft.PrepareMsg]")
			log.Println("[put prepare message into MsgDelivery channel]")
		}
	// 处理CommitMsg信息
	case *pbft.CommitMsg:
		if node.PBFT.CurrentState == nil || node.PBFT.CurrentState.Current_stage != pbft.Prepared {
			node.PBFT.MsgBuffer.CommitMsgs = append(node.PBFT.MsgBuffer.CommitMsgs, msg)
			log.SetPrefix(node.Node_name + "-[pbft.CommitMsg]")
			log.Println("[get a commit message,but don't put it into channel]")
		} else {
			msgs := make([]*pbft.CommitMsg, len(node.PBFT.MsgBuffer.CommitMsgs))
			copy(msgs, node.PBFT.MsgBuffer.CommitMsgs)                  // 复制缓冲数据
			msgs = append(msgs, msg)                                    // 附加新到达的消息
			node.PBFT.MsgBuffer.CommitMsgs = make([]*pbft.CommitMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                                    // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			log.SetPrefix(node.Node_name + "-[pbft.CommitMsg]")
			log.Println("[put commit message into MsgDelivery channel]")
		}
	}
	return nil
}

func (node *Node) routeMsgWhenAlarmed() []error {
	if node.PBFT.CurrentState == nil {
		// 检查PrePrepareMsgs, 并发送到MsgDelivery.
		if len(node.PBFT.MsgBuffer.PrePrepareMsgs) != 0 {
			msgs := make([]*pbft.PrePrepareMsg, len(node.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrePrepareMsgs)
			node.MsgDelivery <- msgs
		}
	} else {
		switch node.PBFT.CurrentState.Current_stage {
		case pbft.PrePrepared:
			// 检查PrepareMsgs,并发送到MsgDelivery.
			if len(node.PBFT.MsgBuffer.PrepareMsgs) != 0 {
				msgs := make([]*pbft.PrepareMsg, len(node.PBFT.MsgBuffer.PrepareMsgs))
				copy(msgs, node.PBFT.MsgBuffer.PrepareMsgs)

				node.MsgDelivery <- msgs
			}
		case pbft.Prepared:
			// 检查CommitMsgs,并发送到MsgDelivery.
			if len(node.PBFT.MsgBuffer.CommitMsgs) != 0 {
				msgs := make([]*pbft.CommitMsg, len(node.PBFT.MsgBuffer.CommitMsgs))
				copy(msgs, node.PBFT.MsgBuffer.CommitMsgs)
				node.MsgDelivery <- msgs
			}
		case pbft.Committed:
			// 检查ReplyMsgs,并发送到MsgDelivery.
			if len(node.PBFT.MsgBuffer.ReplyMsgs) != 0 {
				msgs := make([]*pbft.ReplyMsg, len(node.PBFT.MsgBuffer.ReplyMsgs))
				copy(msgs, node.PBFT.MsgBuffer.ReplyMsgs)
				node.MsgDelivery <- msgs
			}

		}
	}
	return nil
}

// 线程3：alarmToDispatcher，警告信息
func (node *Node) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}

func (node *Node) blockMsg() {
	for {
		msg := <-node.MsgBlock // 从调度器通道中获取缓存信息
		switch msg := msg.(type) {
		case *block.Block:
			init_log("./network/nodelog/block_" + node.Node_name + ".log")
			if node.PBFT.CurrentState == nil { // 如果此时不存在共识
				node.MsgDelivery <- msg
				log.SetPrefix(node.Node_name + "-[block.Block]")
				log.Println("there no pbft on going, now create a new one and put the new block into MsgDelivery channel")
			} else {
				node.MsgEntrance <- msg
				log.SetPrefix(node.Node_name + "-[block.Block]")
				log.Println("exit another pbft,so put the block into MsgEntrance channel agains")
			}
		}
	}
}

// 线程4：resolveMsg
func (node *Node) resolveMsg() {
	for {
		msgs := <-node.MsgDelivery // 从调度器通道中获取缓存信息
		switch msgs := msgs.(type) {
		// 节点表决决策信息
		case *block.Block:
			init_log("./network/nodelog/resolve_" + node.Node_name + ".log")
			if node.PBFT.CurrentState != nil {
				log.SetPrefix(node.Node_name + "-[block.Block]")
				log.Println("get a block,but another pbft consensus is ongoing")
			} else {
				node.PBFT = node.NewConsensus() // 初始化一次共识
				err := node.resolveRequestMsg(msgs)
				if err != nil {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
				log.SetPrefix(node.Node_name + "-[receive block.Block]")
				log.Println("Msgdelivery channel get a new block,and create a new consensus")

			}
		case []*pbft.PrePrepareMsg:
			init_log("./network/nodelog/resolve_" + node.Node_name + ".log")
			if node.PBFT.CurrentState != nil {
				log.SetPrefix(node.Node_name + "-[receive PrePrepareMsg]")
				log.Println("get a preprepare,but another pbft consensus is ongoing")
			} else {
				node.PBFT = node.NewConsensus()          // 初始化一次共识
				err := node.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
				if err != nil {
					fmt.Println(err)
				}

				errs := node.resolvePrePrepareMsg(msgs)
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err) // TODO: send err to ErrorChannel
					}
				}
				log.SetPrefix(node.Node_name + "-[receive PrePrepareMsg]")
				log.Println("Msgdelivery channel get a preprepare, and handle it")
			}
		case []*pbft.PrepareMsg:
			init_log("./network/nodelog/resolve_" + node.Node_name + ".log")
			errs := node.resolvePrepareMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
			log.SetPrefix(node.Node_name + "-[receive PrepareMsg]")
			log.Println("Msgdelivery channel get a prepare message, and handle it")

		case []*pbft.CommitMsg:
			init_log("./network/nodelog/resolve_" + node.Node_name + ".log")
			errs := node.resolveCommitMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
			log.SetPrefix(node.Node_name + "-[receive CommitMsg]")
			log.Println("Msgdelivery channel get a commit message, and handle it")

		}
	}
}

// createStateForNewConsensus，创建新的共识
func (node *Node) createStateForNewConsensus() error {
	if node.PBFT.CurrentState != nil { // 判断当前节点是不是处于其他阶段（预准备阶段或者准备阶段等等）
		return errors.New("another pbft consensus is ongoing") // 如果有，则输出提示
	}
	var lastSequenceID int64 // 获取上一个序列号
	// 判断当前阶段是否已经发送过消息
	if len(node.Committed) == 0 { // 如果是首次进行共识，则上一个序列号lastSequenceID设置为-1
		lastSequenceID = -1
	} else { // 否则取出上一个序列号????
		lastSequenceID = node.Committed[len(node.Committed)-1].Sequence_number
	}
	// 创建新的节点状态，即进行节点状态的初始化
	node.PBFT.CurrentState = pbft.CreateState(node.View.ID, lastSequenceID)
	return nil
}

// node.resolveRequestMsg,[request]处理输入的req消息
func (node *Node) resolveRequestMsg(msgs *block.Block) error {
	err := node.createStateForNewConsensus() // 创建新的共识
	if err != nil {                          // 如果节点未处于共识状态，输出错误
		return err
	}

	prePrepareMsg, err := node.PBFT.CurrentState.PrePrePare(msgs) // 进入共识，获得preprepare消息
	if err != nil {
		return err
	} else {
		node.MsgBroadcast <- prePrepareMsg // 将待广播消息放入通道
		return nil
	}
}

// node.resolvePrePrepareMsg,[pre-prepare]由从节点处理PrePrepare消息
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

// node.resolvePrePrepare,[pre-prepare]
func (node *Node) resolvePrePrepare(prePrepareMsg *pbft.PrePrepareMsg) error {
	prePareMsg, err := node.PBFT.CurrentState.PrePare(prePrepareMsg) // 获得prepare信息
	if err != nil {
		return err
	}
	if prePareMsg != nil {
		init_log("./network/nodelog/resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[get prepareMsg]")
		log.Println("put prepare message into broadcast channel")
		node.MsgBroadcast <- prePareMsg // 将待广播消息放入通道
	}
	return nil
}

// node.resolvePrepareMsg,[prepare]所有联盟节点接收prepare消息，处理得到commit消息
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

// node.resolvePrepare,[prepare]
func (node *Node) resolvePrepare(prepareMsg *pbft.PrepareMsg) error {
	commitMsg, err := node.PBFT.CurrentState.Commit(prepareMsg)
	if err != nil {
		return err
	}
	if commitMsg != nil {
		init_log("./network/nodelog/resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[get commitMsg]")
		log.Println("put commit message into broadcast channel")

		node.MsgBroadcast <- commitMsg // 将待广播消息放入通道
	}
	return nil
}

// node.resolveCommitMsg,[commit]所有节点处理接收的commit消息得到reply消息
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

// node.resolveCommit,[commit]
func (node *Node) resolveCommit(commitMsg *pbft.CommitMsg) error {
	replyMsg, err := node.PBFT.CurrentState.Reply(commitMsg)
	if err != nil {
		return err
	}
	if replyMsg != nil {
		init_log("./network/nodelog/resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[get replyMsg]")
		log.Println("put reply message into broadcast channel")

		node.Committed = append(node.Committed, node.PBFT.CurrentState.CommittedMessage)
		node.MsgBroadcast <- replyMsg // 将待广播消息放入通道
	}
	return nil
}
