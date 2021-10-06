package qbcli

import (
	"qbtx"
	"utils"
)

const CLIENT_LOG_PATH = "../qb/qbcli/cli_log/"

// 客户端
type Client struct {
	Client_name  string            // 客户端名称
	Client_ID    [16]byte          // 客户端ID，16字节QKD设备号
	Client_table map[string]string // 客户端索引表，key=Client_name, value=url
	Node_table   map[string]string // 节点索引表，key=Node_name, value=url

	View         *View             // 视图号
	Transaction  *qbtx.Transaction // 待进行的交易
	CurrentState Stage

	MsgBroadcast chan interface{} // 信息发送通道
	MsgDelivery  chan interface{}
}

// 视图号
type View struct {
	ID      int64  // 视图号
	Primary string // 主节点
}
type Stage int

// 状态标识
const (
	Idle        Stage = iota // Idle=0，节点已成功创建，但共识过程尚未启动。
	PrePrepared              // PrePrepared=1，RequestMsg已成功处理。节点已准备好进入prepare阶段。
)

// NewClient，客户端初始化
// 参数：客户端名称string
// 返回值：初始化的客户端*Client
func NewClient(client_name string) *Client {
	const view = 1 // 暂设视图号为1

	// 初始化节点
	client := &Client{
		Client_name:  client_name,                                                // 客户端名称，形式为C1、C2...
		Client_ID:    utils.GetNodeIDTable(client_name),                          // 客户端ID，16字节QKD设备号
		Client_table: utils.InitConfig(utils.INIT_PATH + "client_localhost.txt"), // 客户端索引表，key=Node_name, value=url
		Node_table:   utils.InitConfig(utils.INIT_PATH + "node_localhost.txt"),   // 联盟节点节点索引表，key=Node_name, value=url

		View: &View{ // 视图号信息，视图号=主节点下标
			ID:      view, // 视图号
			Primary: "P1", // 主节点,暂设为P1
		},
		Transaction:  new(qbtx.Transaction),
		CurrentState: Idle,

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgDelivery:  make(chan interface{}), // 无缓冲的信息发送通道
	}

	client.setRoute() // 设置路由

	// 开启线程goroutine
	go client.broadcastMsg()
	go client.resolveMsg()

	return client
}
