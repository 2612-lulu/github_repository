package qbnode

import (
	"encoding/json"
	"os"
	"pbft"
	"qbtx"
	"time"
	"utils"
)

// 打包时间间隔
const BlockTimeDuration = time.Millisecond * 3000 // 3 second.

// log存放路径
const NODE_LOG_PATH = "../qb/qbnode/nodelog/"

// 节点
type Node struct {
	Node_name            string            // 联盟节点名称
	Node_ID              [16]byte          // 联盟节点ID，16字节QKD设备号
	Node_table           map[string]string // 节点索引表，key=Node_name, value=url
	Node_consensus_table map[string]string
	Addr_table           map[string]string

	TranscationMsgs []qbtx.Transaction

	PBFT_url     string
	Primary      string
	CurrentState Stage // 表明客户端状态

	MsgBroadcast chan interface{} // 广播通道
	MsgEntrance  chan interface{} // 无缓冲的信息接收通道
	MsgDelivery  chan interface{} // 无缓冲的信息发送通道
	MsgBlock     chan interface{} // 打包通道
	Block_clock  chan bool        // 打包计时通道
}
type Stage int

// 状态标识
const (
	Idle Stage = iota // Idle=0，节点已成功创建。
	TX                // TX=1
)

// NewNode，节点初始化
// 参数：节点名称string
// 返回值：经初始化的节点*Node
func NewNode(node_name string) *Node {
	// 初始化节点
	node := &Node{
		Node_name:            node_name,                                                // 联盟节点或客户段名称，形式为P1、P2...
		Node_ID:              utils.GetNodeID(node_name),                               // 客户端ID，16字节QKD设备号
		Node_table:           utils.InitConfig(utils.INIT_PATH + "node_localhost.txt"), // 联盟节点节点索引表，key=Node_name, value=url
		Node_consensus_table: utils.InitConfig(utils.INIT_PATH + "pbft_localhost.txt"),
		Addr_table:           make(map[string]string),
		Primary:              "",
		CurrentState:         Idle,

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgDelivery:  make(chan interface{}),
		MsgBlock:     make(chan interface{}), // 交易信息打包通道
		Block_clock:  make(chan bool),
	}
	file, _ := os.Open("../config/view.json") // 打开文件
	defer file.Close()                        // 关闭文件
	decoder := json.NewDecoder(file)          // NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	var view pbft.View
	err := decoder.Decode(&view) //Decode从输入流读取下一个json编码值并保存在v指向的值里
	if err != nil {
		panic(err)
	}
	node.Primary = view.Primary
	node.PBFT_url = node.Node_consensus_table[node_name]
	qbtx.N = 3*uint32(view.F) + 1
	node.setRoute()
	// 开启线程goroutine
	go node.blockMsg() // 打包通道
	go node.clockToBlock()
	go node.resolveMsg()
	go node.broadcastMsg()

	//node.httplisten() // 开启http
	return node
}
