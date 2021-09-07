package network

import (
	"bytes"
	"fmt"
	"os"
)

var client *Client
var node *Node

func CreateServer(Name string) {
	var NodeName [2]byte // 处理输入参数nodeID
	for i := 0; i < 2; i++ {
		NodeName[i] = []byte(Name)[i]
	}
	if bytes.Equal(NodeName[:1], []byte("P")) { // 如果是联盟节点
		node = NewNode(NodeName)
	} else if bytes.Equal(NodeName[:1], []byte("C")) { // 如果是客户端调用此函数
		client = NewClient(NodeName) // 启动client节点
		m := os.Args[2]
		client.MsgEntrance <- m
		client.httplisten() // 开启http
	} else { // 输入错误
		fmt.Println("	Input Arg[1] wrong,please input again")
		var nodeName string
		fmt.Scanln(&nodeName)
		CreateServer(nodeName)
	}
}
