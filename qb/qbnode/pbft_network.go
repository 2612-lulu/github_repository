package qbnode

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"qb/pbft"
	"qb/qbtools"
)

// node.setRoute,设置路由规则，在启动http服务之前设置
// 参数：无
// 返回值：无
func (node *NodeConsensus) setRoute() {
	http.HandleFunc("/preprepare", node.getPrePrepare)
	http.HandleFunc("/prepare", node.getPrepare)
	http.HandleFunc("/commit", node.getCommit)
}

// node.getPrePrepare,pre-prepare消息解码
// 参数：
// 返回值：无
func (node *NodeConsensus) getPrePrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的消息放入通道MsgEntrance

	qbtools.Init_log(PBFT_LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-listen pre-prepare")
	log.Println("receive a preprepare message")
}

// node.getPrepare,prepare消息解码
// 参数：
// 返回值：无
func (node *NodeConsensus) getPrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的prepare消息放入通道MsgEntrance

	qbtools.Init_log(PBFT_LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-listen prepare")
	log.Println("receive a prepare message")
}

// node.getCommit,commit消息解码
// 参数：
// 返回值：无
func (node *NodeConsensus) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.CommitMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.MsgEntrance <- &msg // 将解码后的commit消息放入通道MsgEntrance
	qbtools.Init_log(PBFT_LOG_PATH + "listenHttp_" + node.Node_name + ".log")
	log.SetPrefix(string(node.Node_name[:]) + "-listen commit")
	log.Println("receive a commit message")
}
