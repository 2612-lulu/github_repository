package network

import (
	"encoding/json"
	"fmt"
	"pbft"
	"utils"
)

// 进程1：broadcastMsg,用于广播消息
func (consensus *NodeConsensus) broadcastMsg() {
	for {
		msg := <-consensus.MsgBroadcast
		switch msg := msg.(type) {
		case *pbft.PrePrepareMsg:
			fmt.Println("====================[START NEW PBFT]==============================")
			utils.LogStage("Request", false)
			utils.LogStage("Request", true)
			utils.LogStage("Pre-Prepare", false)
			consensus.broadcast(msg, "/preprepare") // 发送preprepare信息给其他节点
		case *pbft.PrepareMsg:
			fmt.Println("====================[START NEW PBFT]==============================")
			utils.LogStage("Pre-prepare", true)
			utils.LogStage("Prepare", false)
			consensus.broadcast(msg, "/prepare") // 发送prepare信息给其他节点
		case *pbft.CommitMsg:
			utils.LogStage("Prepare", true)
			utils.LogStage("Commit", false)
			consensus.broadcast(msg, "/commit") // 发送commit信息给其他节点
		case *pbft.ReplyMsg:
			utils.LogStage("Commit", true)
			utils.LogStage("Reply", false)
			consensus.broadcastReply(msg, "/reply")
			utils.LogStage("Reply", true)

			consensus.PBFT.CurrentState = nil
		}
	}
}

// broadcast，节点广播函数
// 参数：待广播消息，
// 返回值：广播错误map[string]error，广播无误len(errorMap) == 0
func (consensus *NodeConsensus) broadcast(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果
	// 将消息广播给其他联盟节点
	for nodeID, url := range consensus.Node_consensus_table {
		if nodeID != consensus.Node_name { // 不需要向自己进行广播
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				errorMap[nodeID] = err
				continue
			}
			// 将json格式传送给其他的联盟节点
			utils.Send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
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

// broadcastReply，节点广播函数,用于广播应答消息
// 参数：待广播消息，
// 返回值：广播错误map[string]error，广播无误len(errorMap) == 0
func (consensus *NodeConsensus) broadcastReply(msg *pbft.ReplyMsg, path string) error {
	jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
	if err != nil {
		return err
	}
	url := consensus.BC_url
	// 将json格式传送给其他的联盟节点
	utils.Send(url+path, jsonMsg) // url：localhost:1111  path：/prepare等等
	return nil

}
