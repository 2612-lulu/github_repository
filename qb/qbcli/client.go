package qbcli

import (
	"qb/pbft"
	"qb/qbtools"
)

const LOG_PATH = "/root/study/github_repository/qb/qbcli/cli_log/"

// 客户端
type Client struct {
	Client_name  string            // 客户端名称
	Client_ID    [16]byte          // 客户端ID，16字节QKD设备号
	Client_table map[string]string // 客户端索引表，key=Client_name, value=url
	Node_table   map[string]string // 节点索引表，key=Node_name, value=url
	View         *pbft.View        // 视图号

	ReplyMsgs    []*pbft.ReplyMsg // 接收的reply消息缓冲列表
	CurrentState *pbft.State

	MsgBroadcast chan interface{} // 信息发送通道
	MsgEntrance  chan interface{}
	MsgDelivery  chan interface{}
}

// NewClient，客户端初始化
// 参数：客户端名称string
// 返回值：初始化的客户端*Client
func NewClient(client_name string) *Client {
	const view = 1 // 暂设视图号为1

	// 初始化节点
	client := &Client{
		Client_name:  client_name,                                                    // 客户端名称，形式为C1、C2...
		Client_ID:    qbtools.GetNodeIDTable(client_name),                            // 客户端ID，16字节QKD设备号
		Client_table: qbtools.InitConfig(qbtools.INIT_PATH + "client_localhost.txt"), // 客户端索引表，key=Node_name, value=url
		Node_table:   qbtools.InitConfig(qbtools.INIT_PATH + "node_localhost.txt"),   // 联盟节点节点索引表，key=Node_name, value=url
		View: &pbft.View{ // 视图号信息，视图号=主节点下标
			ID:      view, // 视图号
			Primary: "P1", // 主节点,暂设为P1
		},

		CurrentState: nil,
		ReplyMsgs:    make([]*pbft.ReplyMsg, 0),

		// 初始化通道Channels
		MsgBroadcast: make(chan interface{}), // 信息发送通道
		MsgEntrance:  make(chan interface{}),
		MsgDelivery:  make(chan interface{}), // 无缓冲的信息发送通道
	}

	client.setRoute() // 设置路由

	// 开启线程goroutine
	go client.broadcastMsg()
	go client.dispatchMsg()
	go client.resolveMsg()

	return client
}
