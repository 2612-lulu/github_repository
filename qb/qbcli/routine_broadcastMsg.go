package qbcli

import (
	"encoding/json"
	"fmt"
	"qbtx"
	"utils"
)

// 线程1：broasdcastMsg，用于广播交易信息
func (client *Client) broadcastMsg() {
	for {
		msg := <-client.MsgBroadcast
		switch msg := msg.(type) {
		case *qbtx.Transaction:
			if client.CurrentState == Idle { // 如果未共识
				client.Transaction = msg          // 更新待交易列表
				jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
				if err != nil {
					fmt.Println(err)
				}
				utils.Send(client.Node_table[client.View.Primary]+"/transaction", jsonMsg)
				client.CurrentState = PrePrepared // 更改状态
			} else {
				fmt.Println("The last transaction didn't finish,please wait")
			}
		}
	}
}
