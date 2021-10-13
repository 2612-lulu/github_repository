package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pbft"
	"qblock"
	"utils"
)

// setRoute,设置路由规则，在启动http服务之前设置
// 参数：无
// 返回值：无
func (consensus *NodeConsensus) setRoute() {
	http.HandleFunc("/request", consensus.getRequest)
	http.HandleFunc("/preprepare", consensus.getPrePrepare)
	http.HandleFunc("/prepare", consensus.getPrepare)
	http.HandleFunc("/commit", consensus.getCommit)
}

func (consensus *NodeConsensus) getRequest(writer http.ResponseWriter, request *http.Request) {
	var msg qblock.Block
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	consensus.PBFT.CurrentState = nil
	consensus.MsgEntrance <- &msg // 将解码后的消息放入通道MsgEntrance

	file, _ := utils.Init_log(PBFT_LOG_PATH + "listenHttp_" + consensus.Node_name + ".log")
	log.SetPrefix("[listen request]")
	log.Println("receive a request message")
	defer file.Close()
}

// getPrePrepare,pre-prepare消息解码
// 参数：
// 返回值：无
func (consensus *NodeConsensus) getPrePrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	consensus.PBFT.CurrentState = nil
	consensus.MsgEntrance <- &msg // 将解码后的消息放入通道MsgEntrance

	file, _ := utils.Init_log(PBFT_LOG_PATH + "listenHttp_" + consensus.Node_name + ".log")
	log.SetPrefix("[listen pre-prepare]")
	log.Println("receive a preprepare message")
	defer file.Close()
}

// getPrepare,prepare消息解码
// 参数：
// 返回值：无
func (consensus *NodeConsensus) getPrepare(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.PrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	consensus.MsgEntrance <- &msg // 将解码后的prepare消息放入通道MsgEntrance

	file, _ := utils.Init_log(PBFT_LOG_PATH + "listenHttp_" + consensus.Node_name + ".log")
	log.SetPrefix("[listen prepare]")
	log.Println("receive a prepare message")
	defer file.Close()
}

// getCommit,commit消息解码
// 参数：
// 返回值：无
func (consensus *NodeConsensus) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg pbft.CommitMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	consensus.MsgEntrance <- &msg // 将解码后的commit消息放入通道MsgEntrance
	file, _ := utils.Init_log(PBFT_LOG_PATH + "listenHttp_" + consensus.Node_name + ".log")
	log.SetPrefix("[listen commit]")
	log.Println("receive a commit message")
	defer file.Close()
}

// node.httplisten，开启Http服务器
// 参数：无
// 返回值：无
func (nodeconsensus *NodeConsensus) Httplisten() {
	url := nodeconsensus.Node_consensus_table[nodeconsensus.Node_name]
	fmt.Printf("Node will be started at %s...\n", url)
	if err := http.ListenAndServe(url, nil); err != nil {
		fmt.Println(err)
		return
	}
}
