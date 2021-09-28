package qbcli

import (
	"encoding/json"
	"fmt"
	"log"
	"qb/pbft"
	"qb/qbtools"
	"qb/qbtx"
)

// 线程1：broasdcastMsg，用于广播交易信息
func (client *Client) broadcastMsg() {
	for {
		msg := <-client.MsgBroadcast
		switch msg := msg.(type) {
		case *qbtx.ToGenTx:
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				fmt.Println(err)
			}
			qbtools.Send(client.Node_table[client.View.Primary]+"/togentx", jsonMsg)

			qbtools.Init_log(CLIENT_LOG_PATH + "broadcast_" + client.Client_name + ".log")
			log.SetPrefix("【send togentx】")
			log.Println("send a transcation message to the Primary node to get utxo")
		case *qbtx.Transaction:
			if client.CurrentState == pbft.Idle { // 如果未共识
				client.Transaction = msg          // 更新待交易列表
				jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
				if err != nil {
					fmt.Println(err)
				}
				qbtools.Send(client.Node_table[client.View.Primary]+"/transaction", jsonMsg)
				client.CurrentState = pbft.PrePrepared // 更改状态
				qbtools.Init_log(CLIENT_LOG_PATH + "broadcast_" + client.Client_name + ".log")
				log.SetPrefix("【send transaction】")
				log.Println("send a transcation to the Primary node")
			} else {
				qbtools.Init_log(CLIENT_LOG_PATH + "broadcast_" + client.Client_name + ".log")
				log.SetPrefix("【send transaction error】")
				log.Println("please wait the last tx finish")
			}
		}
	}
}
