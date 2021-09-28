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
			qbtools.Init_log(CLIENT_LOG_PATH + "result.log")
			fmt.Println(msgs)
			log.SetPrefix("【transaction succes】")
			log.Println("transcation success")
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
