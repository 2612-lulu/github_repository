package qbcli

import (
	"fmt"
	"log"
	"qb/pbft"
	"qb/qbtools"
	"qb/qbtx"
)

// 线程3：resolveMsg，用于对收到的信息作具体处理
func (client *Client) resolveMsg() {
	for {
		msgs := <-client.MsgDelivery // 从调度器通道中获取缓存信息
		switch msgs := msgs.(type) {
		case []*pbft.ReplyMsg:
			i := 0
			for _, reply := range msgs {
				if reply.Result {
					i++
				}
			}
			if i >= 3 {
				fmt.Println("====================【transaction success】==============================")
				from := client.Transaction.Vin[0].From
				for index, vout := range client.Transaction.Vout {
					if from != vout.To {
						fmt.Printf("==transaction %d\n", index+1)
						fmt.Printf("\tFrom: %s\n", from)
						fmt.Printf("\tTo:   %s\n", vout.To)
						fmt.Printf("\tValue:%d\n", vout.Value)
					}
				}
				qbtools.Init_log(CLIENT_LOG_PATH + "resolve_" + client.Client_name + ".log")
				log.SetPrefix("【transaction succes】")
				log.Println("transcation success")

				client.Transaction = new(qbtx.Transaction) // 清空待交易列表
				client.ReplyMsgs = make([]*pbft.ReplyMsg, 0)
				client.CurrentState = pbft.Idle // 更改状态，此后可以发送新的交易
			} else {
				qbtools.Init_log(CLIENT_LOG_PATH + "resolve_" + client.Client_name + ".log")
				log.SetPrefix("【transaction error】")
				log.Println("number of reply is wrong")
				client.ReplyMsgs = make([]*pbft.ReplyMsg, 0)
			}
		case *qbtx.Transaction:
			qbtools.Init_log(CLIENT_LOG_PATH + "resolve_" + client.Client_name + ".log")
			log.SetPrefix("【generate tx】")
			log.Println("get a tx and prepare to send it to the Primary")
			tx := client.SignTranscation(msgs)
			client.MsgBroadcast <- tx
		}
	}
}

// genTranscationMsg,将用户输入的内容处理为交易信息结构
// 参数：用户输入的内容string
// 返回值：交易信息*block.Transaction
func (client *Client) SignTranscation(tx *qbtx.Transaction) *qbtx.Transaction {
	tx.SignTX(client.Client_name)
	return tx
}
