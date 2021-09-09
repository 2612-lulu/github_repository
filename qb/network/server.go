package network

import (
	"fmt"
	"os"
)

var client *Client
var node *Node

func CreateServer(Name string) {
	if Name[:1] == "P" { // 如果是联盟节点
		node = NewNode(Name)
	} else if Name[:1] == "C" { // 如果是客户端调用此函数
		client = NewClient(Name) // 启动client节点
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
