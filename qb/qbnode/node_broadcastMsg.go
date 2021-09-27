package qbnode

import (
	"encoding/json"
	"fmt"
	"log"
	"qb/qbtools"
	"qb/qbtx"
)

// 线程1：broasdcastMsg，用于广播交易信息
func (node *Node) broadcastMsg() {
	for {
		msg := <-node.MsgBroadcast
		switch msg := msg.(type) {
		case *qbtx.Transaction:
			jsonMsg, err := json.Marshal(msg) // 将msg信息编码成json格式
			if err != nil {
				fmt.Println(err)
			}
			addr := msg.Vin[0].From
			to_client := node.Addr_table[addr]
			qbtools.Send(node.Client_table[to_client]+"/txmessage", jsonMsg)

			qbtools.Init_log(LOG_PATH + "broadcast_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[send TXmessage]")
			log.Println("send a transcation message to the client")
		}
	}
}
