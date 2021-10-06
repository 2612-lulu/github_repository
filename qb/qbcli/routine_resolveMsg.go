package qbcli

import (
	"fmt"
	"log"
	"pbft"
	"qbtx"
	"utils"
)

// 线程3：resolveMsg，用于对收到的信息作具体处理
func (client *Client) resolveMsg() {
	for {
		msg := <-client.MsgDelivery // 从调度器通道中获取缓存信息
		switch msg := msg.(type) {
		case *qbtx.Transaction:
			tx := client.SignTranscation(msg)
			client.MsgBroadcast <- tx
		case *pbft.ReplyMsg:
			client.resolveTXreply(msg)
		}
	}
}

// genTranscationMsg,将用户输入的内容处理为交易信息结构
// 参数：用户输入的内容string
// 返回值：交易信息*block.Transaction
func (client *Client) SignTranscation(tx *qbtx.Transaction) *qbtx.Transaction {
	tx.SignTX(client.Client_name)
	file, _ := utils.Init_log(CLIENT_LOG_PATH + "resolve_" + client.Client_name + ".log")
	log.SetPrefix("【generate tx】")
	log.Println("get a tx and prepare to send it to the Primary")
	defer file.Close()
	return tx
}

func (client *Client) resolveTXreply(msg *pbft.ReplyMsg) {
	fmt.Println("get a reply")

}
