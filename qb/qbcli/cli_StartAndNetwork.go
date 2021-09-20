package qbcli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qb/pbft"
	"qb/qbtools"
	"qb/qbtx"
	"qb/qkdserv"
	"qb/quantumbc"
)

const LOG_PATH = "/root/study/github_repository/qb/qbcli/log/"

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
	const view = 1                  // 暂设视图号为1
	qkdserv.Node_name = client_name // 调用此程序的当前节点或客户端名称
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)

	//w := qbwallet.NewWallet()
	bc := quantumbc.CreateBlockchain(client_name, client_name)
	defer bc.DB.Close()
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

// client.setRoute,设置路由规则，在启动http服务之前设置
// 参数：
// 返回值：无
func (client *Client) setRoute() {
	http.HandleFunc("/reply", client.getReply)
}

// server.getReply,reply消息解码
// 参数：
// 返回值：无
func (client *Client) getReply(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.ReplyMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	client.MsgEntrance <- &msg
}

// client.httplisten，开启Http服务器
// 参数：
// 返回值：无
func (client *Client) Httplisten() {
	url := client.Client_table[client.Client_name]
	fmt.Printf("Server will be started at %s...\n", url)
	if err := http.ListenAndServe(url, nil); err != nil {
		fmt.Println(err)
		return
	}
}

// 线程1：broasdcastMsg，用于广播交易信息
func (client *Client) broadcastMsg() {
	for {
		msg := <-client.MsgBroadcast
		switch msg := msg.(type) {
		case *qbtx.Transaction:
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				fmt.Println(err)
			}
			qbtools.Send(client.Node_table[client.View.Primary]+"/transcation", jsonMsg)

			qbtools.Init_log(LOG_PATH + client.Client_name + ".log")
			log.SetPrefix("[send transcation]")
			log.Println("send a transcation to the Primary node")
		}
	}
}

// 线程2：dispatchMsg，用于处理收到的消息，一般用于暂时存储消息
func (client *Client) dispatchMsg() {
	for {
		msg := <-client.MsgEntrance
		switch msg := msg.(type) {
		case *pbft.ReplyMsg:
			if len(client.ReplyMsgs) >= 2*pbft.F { // 收到符合要求的reply消息
				msgs := make([]*pbft.ReplyMsg, len(client.ReplyMsgs))
				copy(msgs, client.ReplyMsgs) // 复制缓冲数据
				msgs = append(msgs, msg)     // 附加新到达的消息
				client.MsgDelivery <- msgs
				client.ReplyMsgs = make([]*pbft.ReplyMsg, 0) // 清空
			} else {
				client.ReplyMsgs = append(client.ReplyMsgs, msg)
			}
		}
	}
}

// client.genTranscationMsg,将用户输入的内容处理为交易信息结构
// 参数：用户输入的内容string
// 返回值：交易信息*block.Transaction
func (client *Client) genTranscationMsg(message string) *qbtx.Transaction {
	var tx qbtx.Transaction
	return &tx
}

// 线程3：resolveMsg，用于对收到的信息作具体处理
func (client *Client) resolveMsg() {
	for {
		msgs := <-client.MsgDelivery // 从调度器通道中获取缓存信息
		switch msgs := msgs.(type) {
		case []*pbft.ReplyMsg:
			qbtools.Init_log(LOG_PATH + "result.log")
			fmt.Println(msgs)
			log.SetPrefix(client.Client_name + "-tx success-")
			log.Println("transcation success")
		}
	}
}
