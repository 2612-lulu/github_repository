package qbcli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qb/pbft"
	"qb/qbtools"
	"qb/qbtx"
)

// client.setRoute,设置路由规则，在启动http服务之前设置
// 参数：
// 返回值：无
func (client *Client) setRoute() {
	http.HandleFunc("/txmessage", client.getTXmessage)
	http.HandleFunc("/reply", client.getReply)
}

// server.getTXmessage,reply消息解码
// 参数：
// 返回值：无
func (client *Client) getTXmessage(writer http.ResponseWriter, request *http.Request) {
	var msg qbtx.Transaction
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	client.MsgDelivery <- &msg

	qbtools.Init_log(CLIENT_LOG_PATH + "listenHttp_" + client.Client_name + ".log")
	log.SetPrefix("【receive txmessage】")
	log.Println("receive a transcation message")
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
	qbtools.Init_log(CLIENT_LOG_PATH + "listenHttp_" + client.Client_name + ".log")
	log.SetPrefix("【receive reply】")
	log.Println("receive the result of tx")
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
