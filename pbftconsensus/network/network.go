package network

import (
	"pbft"
	"qblock"
	"qbtx"
	"time"
	"utils"
)

// 数据处理时间限制
const ResolvingTimeDuration = time.Millisecond * 200 // 0.2 second.

const PBFT_LOG_PATH = "../pbftconsensus/network/network_log/"

var pbft_bc map[string]string

// 节点共识
type NodeConsensus struct {
	Node_name            string            // 联盟节点名称
	Node_ID              [16]byte          // 联盟节点ID，16字节QKD设备号
	Node_consensus_table map[string]string // 共识索引表，key=Node_name, value=url
	Node_table           map[string]string // 节点索引表，key=Node_name, value=url
	BC_url               string

	View      *pbft.View // 视图号
	PBFT      *Consensus
	Committed []*pbft.CommitMsg
	Req       *qblock.Block

	MsgBroadcast chan interface{} // 广播通道
	MsgEntrance  chan interface{} // 无缓冲的信息接收通道
	MsgDelivery  chan interface{} // 无缓冲的信息发送通道
	Alarm        chan bool        // 警告通道
	Result       chan interface{} // pbft结果
}

// 共识
type Consensus struct {
	CurrentState *pbft.State // 节点状态，默认为nil
	MsgBuffer    *MsgBuffer  // 五种消息类型缓冲列表
}

// 数据缓存区
type MsgBuffer struct {
	TranscationMsgs []*qbtx.Transaction
	ReqMsgs         []*qblock.Block
	PrePrepareMsgs  []*pbft.PrePrepareMsg
	PrepareMsgs     []*pbft.PrepareMsg
	CommitMsgs      []*pbft.CommitMsg
}

// NewNodeConsensus，节点共识初始化
// 参数：节点名称string
// 返回值：经初始化的节点*Node
func NewNodeConsensus(node_name string) *NodeConsensus {
	const view = 1 // 暂设视图号为1

	// 初始化节点
	node_consensus := &NodeConsensus{
		Node_name:            node_name,                                                // 联盟节点或客户段名称，形式为P1、P2...
		Node_ID:              utils.GetNodeIDTable(node_name),                          // 客户端ID，16字节QKD设备号
		Node_consensus_table: utils.InitConfig(utils.INIT_PATH + "pbft_localhost.txt"), // 联盟节点节点索引表，key=Node_name, value=url
		Node_table:           utils.InitConfig(utils.INIT_PATH + "node_localhost.txt"), // 联盟节点节点索引表，key=Node_name, value=url

		View: &pbft.View{ // 视图号信息，视图号=主节点下标
			ID:      view, // 视图号
			Primary: "P1", // 主节点,暂设为P1
		},

		PBFT: &Consensus{
			CurrentState: nil,
			MsgBuffer: &MsgBuffer{ // 初始化
				ReqMsgs:        make([]*qblock.Block, 0),
				PrePrepareMsgs: make([]*pbft.PrePrepareMsg, 0),
				PrepareMsgs:    make([]*pbft.PrepareMsg, 0),
				CommitMsgs:     make([]*pbft.CommitMsg, 0),
			},
		},
		Committed: make([]*pbft.CommitMsg, 0),
		Req:       new(qblock.Block),

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgEntrance:  make(chan interface{}), // 无缓冲的信息接收通道
		MsgDelivery:  make(chan interface{}), // 无缓冲的信息发送通道
		Alarm:        make(chan bool),        // 警告通道
		Result:       make(chan interface{}),
	}
	pbft_bc = utils.InitConfig(utils.INIT_PATH + "pbft_bc.txt")
	pbft_url := node_consensus.Node_consensus_table[node_consensus.Node_name]
	url, ok := pbft_bc[pbft_url]
	if ok {
		node_consensus.BC_url = url
	}

	node_consensus.setRoute() // 设置路由

	// 开启线程goroutine
	go node_consensus.broadcastMsg()      // 广播信息
	go node_consensus.dispatchMsg()       // 启动消息调度器
	go node_consensus.alarmToDispatcher() // Start alarm trigger
	go node_consensus.resolveMsg()        // 开始信息表决

	node_consensus.Httplisten() // 开启http
	return node_consensus
}
