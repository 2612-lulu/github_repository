package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qb/block"
	"qb/qbtools"
	"time"
)

// 打包时间间隔
const BlockTimeDuration = time.Millisecond * 2000 // 1 second.
// 节点
type Node struct {
	Node_name            string            // 联盟节点名称
	Node_ID              [16]byte          // 联盟节点ID，16字节QKD设备号
	Node_table           map[string]string // 节点索引表，key=Node_name, value=url
	Client_table         map[string]string // 客户端索引表，key=Client_name, value=url
	Node_consensus_table map[string]string // 节点共识端口索引表，key=Node_name, value=url

	PBFT_consensus *NodeConsensus

	TranscationMsgs []*block.Transaction

	MsgBroadcast chan interface{} // 广播通道
	MsgBlock     chan interface{} // 打包通道
	MsgEntrance  chan interface{} // 无缓冲的信息接收通道
	MsgDelivery  chan interface{} // 无缓冲的信息发送通道
	Block_clock  chan bool        // 打包计时通道
}

// NewNode，节点初始化
// 参数：节点名称string
// 返回值：经初始化的节点*Node
func NewNode(node_name string) *Node {
	//const view = 1                // 暂设视图号为1
	//qkdserv.Node_name = node_name // 调用此程序的当前节点或客户端名称
	// 初始化签名密钥池
	//qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)

	// 初始化节点
	node := &Node{
		Node_name:            node_name,                                                           // 联盟节点或客户段名称，形式为P1、P2...
		Node_ID:              qbtools.GetNodeIDTable(node_name),                                   // 客户端ID，16字节QKD设备号
		Node_table:           qbtools.InitConfig("./qbtools/config/node_localhost.txt"),           // 联盟节点节点索引表，key=Node_name, value=url
		Node_consensus_table: qbtools.InitConfig("./qbtools/config/node_consensus_localhost.txt"), // 联盟节点节点索引表，key=Node_name, value=url
		Client_table:         qbtools.InitConfig("./qbtools/config/client_localhost.txt"),         // 客户端索引表，key=Client_name, value=url

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgBlock:     make(chan interface{}), // 交易信息打包通道
		MsgEntrance:  make(chan interface{}), // 无缓冲的信息接收通道
		MsgDelivery:  make(chan interface{}), // 无缓冲的信息发送通道
		Block_clock:  make(chan bool),
	}

	node.setRoute()
	// 开启线程goroutine
	go node.goConsensus()
	go node.broadcastMsg() // 广播信息
	//go node.resolveMsg()   // 启动消息调度器
	go node.blockMsg() // 打包通道
	go node.clockToBlock()

	node.httplisten() // 开启http
	return node
}
func (node *Node) setRoute() {
	http.HandleFunc("/transcation", node.getTranscation)
}
func (node *Node) getTranscation(writer http.ResponseWriter, request *http.Request) {
	var msg block.Transaction
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgBlock <- &msg // 将解码后的交易消息放入打包通道

	qbtools.Init_log("./network/consensuslog/listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-transcation")
	log.Println("receive a transcation message")
}

// node.httplisten，开启Http服务器
// 参数：无
// 返回值：无
func (node *Node) httplisten() {
	url := node.Node_table[node.Node_name]
	fmt.Printf("Node will be started at %s...\n", url)
	if err := http.ListenAndServe(url, nil); err != nil {
		fmt.Println(err)
		return
	}
}

func (node *Node) goConsensus() {

	node.PBFT_consensus = NewNodeConsensus(node.Node_name)
}

// 进程2：broadcastMsg,用于广播消息
func (node *Node) broadcastMsg() {
	for {
		msg := <-node.MsgBroadcast
		switch msg := msg.(type) {
		case *block.Block:
			url := node.Node_consensus_table[node.Node_name]
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				fmt.Println(err)
			}
			// 将json格式传送给其他的联盟节点
			send(url+"/request", jsonMsg) // url：localhost:1111  path：/prepare等等
			qbtools.Init_log("./network/consensuslog/broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast request]")
			log.Println("broadcast request message to pbft")
		}
	}
}

// 线程3：clockToBlock，提醒打包时间片（1s）到
func (node *Node) clockToBlock() {
	for {
		time.Sleep(BlockTimeDuration)
		node.Block_clock <- true
	}
}

// 线程5：blockMsg，用于打包收到的交易信息
func (node *Node) blockMsg() {
	for {
		select {
		case msg := <-node.MsgBlock: // 从调度器通道中获取缓存信息
			err := node.startTopbft(msg)
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		case <-node.Block_clock:
			err := node.blockWhenClock()
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		}
	}
}

// node.startTopbft,将收到的交易信息存储到数组
// 参数：收到的消息
// 返回值：处理错误error，默认为nil
func (node *Node) startTopbft(msg interface{}) error {
	switch msg := msg.(type) {
	case *block.Transaction:
		node.TranscationMsgs = append(node.TranscationMsgs, msg)

		qbtools.Init_log("./network/consensuslog/block_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[block.Transaction]")
		log.Println("get a transcation message")
	}
	return nil
}

// node.blockWhenClock,当时间片到时，将收到的交易信息打包
// 参数：无
// 返回值：处理错误error，默认为nil
func (node *Node) blockWhenClock() error {
	if len(node.TranscationMsgs) >= block.BLOCK_LENGTH {
		qbtools.Init_log("./network/consensuslog/block_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[block.Block]")
		log.Println("create a new block, and put it into MsgEntrance channel")
		//fmt.Println("打包了一个区块，长度=", len(node.PBFT.MsgBuffer.TranscationMsgs))

		msgs := make([]*block.Transaction, len(node.TranscationMsgs))
		copy(msgs, node.TranscationMsgs) // 复制缓冲数据
		request := block.CreateBlock(msgs)
		node.TranscationMsgs = make([]*block.Transaction, 0) // 清空重置

		node.MsgBroadcast <- request
	}
	return nil
}
