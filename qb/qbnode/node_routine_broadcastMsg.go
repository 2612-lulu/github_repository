package qbnode

import (
	"encoding/json"
	"fmt"
	"log"
	"pbft"
	"qb/qbutxo"
	"qb/quantumbc"
	"qblock"
	"qbtx"
	"utils"
)

// 线程1：broasdcastMsg，用于广播交易信息
func (node *Node) broadcastMsg() {
	for {
		msg := <-node.MsgBroadcast
		switch msg := msg.(type) {
		case *qbtx.Transaction: // 客户端发送交易
			if node.CurrentState == Idle { // 如果未共识
				jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
				if err != nil {
					fmt.Println(err)
				}
				utils.Send(node.Node_table[node.Primary]+"/transaction", jsonMsg)
				node.CurrentState = TX // 更改状态
			} else {
				fmt.Println("The last transaction didn't finish,please wait")
			}
		case *qblock.Block:
			node.broadcastBlock(msg)
		case *pbft.ReplyMsg:
			// TODO1:区块上链
			// TODO2：主节点发送交易结果
			node.addBlock(msg.Request)
			if node.Node_name == node.Primary {
				node.broadcast(msg, "/txreply")
			}
		}
	}
}

func (node *Node) broadcastBlock(msg *qblock.Block) error {
	jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
	if err != nil {
		fmt.Println(err)
	}
	utils.Send(node.PBFT_url+"/request", jsonMsg) // 发送给对应的pbft
	file, _ := utils.Init_log(utils.FLOW_PATH + node.Node_name + ".log")
	defer file.Close()
	log.SetPrefix("PBFT--[REQUEST DONE]----")
	log.Println("broadcast request message, start pbft")
	return nil
}

func (node *Node) addBlock(block *qblock.Block) {
	bc := quantumbc.NewBlockchain(node.Node_name) // 获取账本
	UTXOSet := qbutxo.UTXOSet{                    // 设置utxo
		Blockchain: bc,
	}
	bc.AddBlock(block)
	defer bc.DB.Close() // 关闭数据库
	UTXOSet.Update(block)
	UTXOSet.Reindex()

	if node.Node_name == node.Primary {
		file, _ := utils.Init_log(utils.FLOW_PATH + node.Node_name + ".log")
		defer file.Close()
		log.SetPrefix("UPDATE BLOCKCHAIN-------")
		log.Println("add new block to blockchain")
		log.SetPrefix("RETURN RESULT-----------")
		log.Printf("Returns the result of the transaction\n\n")
	} else {
		file, _ := utils.Init_log(utils.FLOW_PATH + node.Node_name + ".log")
		defer file.Close()
		log.SetPrefix("UPDATE BLOCKCHAIN-------")
		log.Printf("add new block to blockchain\n\n")
	}
	//count := UTXOSet.CountTransactions()
	//fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
	//quantumbc.PrintBlockChain(node.Node_name)
}

func (node *Node) broadcast(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error) // 存放广播结果
	// 将消息广播给其他联盟节点
	for nodeID, url := range node.Node_table {
		if nodeID != node.Node_name { // 不需要向自己进行广播
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
