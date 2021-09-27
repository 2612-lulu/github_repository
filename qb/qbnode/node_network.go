package qbnode

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qb/qbtools"
	"qb/qbtx"
)

// 设置路由
func (node *Node) setRoute() {
	http.HandleFunc("/transaction", node.getTranscation)
	http.HandleFunc("/togentx", node.getAndGenTx)
}

// getANdFenTx，解析交易消息
func (node *Node) getAndGenTx(writer http.ResponseWriter, request *http.Request) {
	var msg qbtx.ToGenTx
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgDelivery <- &msg // 将解码后的交易消息放入打包通道

	qbtools.Init_log(LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-tx_message")
	log.Println("receive a transcation esstencial message")
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

	qbtools.Init_log(LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-transcation")
	log.Println("receive a transcation message")
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
