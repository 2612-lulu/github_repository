package qbnode

import (
	"qb/pbft"
	"qb/qbtools"
	"qb/qbtx"
	"time"
)

// 打包时间间隔
const BlockTimeDuration = time.Millisecond * 2000 // 1 second.

// log存放路径
const LOG_PATH = "/root/study/github_repository/qb/qbnode/node_log/"

// 节点
type Node struct {
	Node_name    string            // 联盟节点名称
	Node_ID      [16]byte          // 联盟节点ID，16字节QKD设备号
	Node_table   map[string]string // 节点索引表，key=Node_name, value=url
	Client_table map[string]string // 客户端索引表，key=Client_name, value=url
	Addr_table   map[string]string

	PBFT_consensus *pbft.NodeConsensus

	TranscationMsgs []*qbtx.Transaction

	MsgBroadcast chan interface{} // 广播通道
	MsgEntrance  chan interface{} // 无缓冲的信息接收通道
	MsgDelivery  chan interface{} // 无缓冲的信息发送通道
	MsgBlock     chan interface{} // 打包通道
	Block_clock  chan bool        // 打包计时通道
}

// NewNode，节点初始化
// 参数：节点名称string
// 返回值：经初始化的节点*Node
func NewNode(node_name string) *Node {
	// 初始化节点
	node := &Node{
		Node_name:    node_name,                                                   // 联盟节点或客户段名称，形式为P1、P2...
		Node_ID:      qbtools.GetNodeIDTable(node_name),                           // 客户端ID，16字节QKD设备号
		Node_table:   qbtools.InitConfig("./qbtools/config/node_localhost.txt"),   // 联盟节点节点索引表，key=Node_name, value=url
		Client_table: qbtools.InitConfig("./qbtools/config/client_localhost.txt"), // 客户端索引表，key=Client_name, value=url
		Addr_table:   make(map[string]string),

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgEntrance:  make(chan interface{}), // 无缓冲的信息接收通道
		MsgDelivery:  make(chan interface{}), // 无缓冲的信息发送通道
		MsgBlock:     make(chan interface{}), // 交易信息打包通道
		Block_clock:  make(chan bool),
	}

	node.setRoute()
	// 开启线程goroutine
	go node.blockMsg() // 打包通道
	go node.clockToBlock()
	go node.resolveMsg()
	go node.broadcastMsg()
	go node.pbftResolveResult()

	node.PBFT_consensus = pbft.NewNodeConsensus(node_name) // 开启共识

	//node.httplisten() // 开启http
	return node
}
