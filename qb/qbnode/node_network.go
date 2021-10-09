package qbnode

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pbft"
	"qbtx"
	"utils"
)

// 设置路由
func (node *Node) setRoute() {
	http.HandleFunc("/transaction", node.getTranscation)
	http.HandleFunc("/reply", node.getReply)
	http.HandleFunc("/txreply", node.getTXReply)
}

// getTranscation，解析交易消息
func (node *Node) getTranscation(writer http.ResponseWriter, request *http.Request) {
	var msg qbtx.Transaction
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgBlock <- &msg // 将解码后的交易消息放入打包通道
	file, _ := utils.Init_log(NODE_LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix("[listen transcation]")
	log.Println("receive a transcation message")
	defer file.Close()
}

// getTranscation，解析交易消息
func (node *Node) getReply(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.ReplyMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgBroadcast <- &msg // 将解码后的交易消息放入通道
	file, _ := utils.Init_log(NODE_LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix("[listen reply]")
	log.Println("receive a transcation message")
	defer file.Close()
}

func (node *Node) getTXReply(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.ReplyMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgDelivery <- &msg
	file, _ := utils.Init_log(CLIENT_LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix("[listen reply]")
	log.Println("receive the result of tx")
	defer file.Close()

}

// node.httplisten，开启Http服务器
// 参数：无
// 返回值：无
func (node *Node) Httplisten() {
	url := node.Node_table[node.Node_name]
	fmt.Printf("Node will be started at %s...\n", url)
	if err := http.ListenAndServe(url, nil); err != nil {
		fmt.Println(err)
		return
	}
}
