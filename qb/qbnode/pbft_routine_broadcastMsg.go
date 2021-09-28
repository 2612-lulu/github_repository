package qbnode

import (
	"encoding/json"
	"fmt"
	"log"
	"qb/pbft"
	"qb/qbtools"
)

// 进程1：broadcastMsg,用于广播消息
func (node *NodeConsensus) broadcastMsg() {
	for {
		msg := <-node.MsgBroadcast
		switch msg := msg.(type) {
		case *pbft.PrePrepareMsg:
			fmt.Println("====================[START NEW PBFT]==============================")
			qbtools.LogStage("Request", false)
			qbtools.LogStage("Request", true)
			qbtools.LogStage("Pre-Prepare", false)
			node.broadcast(msg, "/preprepare") // 发送preprepare信息给其他节点

			qbtools.Init_log(PBFT_LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast preprepare]")
			log.Println("broadcast preprepare message")
		case *pbft.PrepareMsg:
			fmt.Println("====================[START NEW PBFT]==============================")
			qbtools.LogStage("Pre-prepare", true)
			qbtools.LogStage("Prepare", false)
			node.broadcast(msg, "/prepare") // 发送prepare信息给其他节点

			qbtools.Init_log(PBFT_LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast prepare]")
			log.Println("broadcast prepare message")
		case *pbft.CommitMsg:
			qbtools.LogStage("Prepare", true)
			qbtools.LogStage("Commit", false)
			node.broadcast(msg, "/commit") // 发送commit信息给其他节点

			qbtools.Init_log(PBFT_LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast commit]")
			log.Println("broadcast commit message")
		case []*pbft.ReplyMsg:
			qbtools.LogStage("Commit", true)
			qbtools.LogStage("Reply", false)
			node.broadcastReply(msg, "/reply")
			qbtools.LogStage("Reply", true)

			node.PBFT.CurrentState = nil

			qbtools.Init_log(PBFT_LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[broadcast reply]")
			log.Println("broadcast reply message")
		}
	}
}

// node.broadcast，节点广播函数
// 参数：待广播消息，
// 返回值：广播错误map[string]error，广播无误len(errorMap) == 0
func (node *NodeConsensus) broadcast(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果
	// 将消息广播给其他联盟节点
	for nodeID, url := range node.Node_consensus_table {
		if nodeID != node.Node_name { // 不需要向自己进行广播
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				errorMap[nodeID] = err
				continue
			}
			// 将json格式传送给其他的联盟节点
			qbtools.Send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
		} else {
			continue
		}

	}

	if len(errorMap) == 0 { // 如果转发消息均成功
		return nil
	} else { // 如果有转发失败的情况
		return errorMap
	}
}

// node.broadcastReply，节点广播函数,用于广播应答消息
// 参数：待广播消息，
// 返回值：广播错误map[string]error，广播无误len(errorMap) == 0
func (node *NodeConsensus) broadcastReply(msg []*pbft.ReplyMsg, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果

	// 将Reply消息广播给相应的客户端
	for _, reply := range msg {
		jsonMsg, err := json.Marshal(reply) // 将msg信息编码成json格式
		if err != nil {
			errorMap[reply.Client_name] = err
			continue
		}
		url := node.Client_table[reply.Client_name]
		// 将json格式发送到相应客户端
		qbtools.Send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
	}

	if len(errorMap) == 0 { // 如果转发消息均成功
		return nil
	} else { // 如果有转发失败的情况
		return errorMap
	}
}
