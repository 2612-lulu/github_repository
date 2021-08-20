package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"qb/pbft"
)

type Server struct {
	url  string
	node *Node
}

// server.NewServer，server服务初始化
func NewServer(node_name string) *Server {
	// 根据传进来的Node_name新建了一个节点
	// 节点的默认视图是1，并且该节点启动了三个协程：dispatchMsg、alarmToDispatcher、resolveMsg
	node := NewNode(node_name)

	// 为该节点创建一个server服务
	var NodeName [2]byte // 处理输入参数node_name
	for i := 0; i < 2; i++ {
		NodeName[i] = []byte(node_name)[i]
	}
	var server *Server
	if bytes.Equal(NodeName[:1], []byte("P")) { // 如果是联盟节点开启此服务，则进入监听状态
		server = &Server{node.NodeTable[NodeName], node}
		//fmt.Println("The server is used by node")
		// 设置路由
		server.setRoute()
	} else if bytes.Equal(NodeName[:1], []byte("C")) { // 如果是客户端调用此服务，则输入操作指令即可发送req请求共识
		server = &Server{node.ClientTable[NodeName], node}
		fmt.Println("The server is used by client")
		fmt.Println("Now you can input message to request pbft")
		var operation string
		fmt.Scanln(&operation)
		node.Request(operation, NodeName)
		// 设置路由
		server.setRoute()
	} else {
		fmt.Println("The input is wrong,please input again")
		fmt.Scanln(&node_name)
		server = NewServer(node_name)
	}
	return server
}

// server.Start，开启Http服务器
func (server *Server) Start() {
	fmt.Printf("Server will be started at %s...\n", server.url)
	if err := http.ListenAndServe(server.url, nil); err != nil {
		fmt.Println(err)
		return
	}
}

// server.setRoute,设置路由规则，在启动http服务之前设置
func (server *Server) setRoute() {
	http.HandleFunc("/request", server.getReq)
	http.HandleFunc("/preprepare", server.getPrePrepare)
	http.HandleFunc("/prepare", server.getPrepare)
	http.HandleFunc("/commit", server.getCommit)
	http.HandleFunc("/reply", server.getReply)
}

// server.getReq,request消息解码
func (server *Server) getReq(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.RequestMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	server.node.MsgEntrance <- &msg // 将解码后的request消息放入消息接收通道
}

// server.getPrePrepare,pre-prepare消息解码
func (server *Server) getPrePrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	server.node.MsgEntrance <- &msg // 将解码后的消息放入通道MsgEntrance
}

// server.getPrepare,prepare消息解码
func (server *Server) getPrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg // 将解码后的prepare消息放入通道MsgEntrance
}

// server.getCommit,commit消息解码
func (server *Server) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.CommitMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg // 将解码后的commit消息放入通道MsgEntrance
}

// server.getReply,reply消息解码
func (server *Server) getReply(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.ReplyMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.GetReply(&msg) // 将解码后的reply消息放入通道MsgEntrance
}

// send，通信函数，实现点对点通信
func send(url string, msg []byte) {
	buff := bytes.NewBuffer(msg)
	http.Post("http://"+url, "application/json", buff)
}
