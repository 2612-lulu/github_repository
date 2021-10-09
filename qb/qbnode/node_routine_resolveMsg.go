package qbnode

import (
	"fmt"
	"log"
	"pbft"
	"qb/qbwallet"
	"qbtx"
	"utils"
)

// 线程3：resolveMsg，用于对收到的信息作具体处理
func (node *Node) resolveMsg() {
	for {
		msg := <-node.MsgDelivery // 从调度器通道中获取缓存信息
		switch msg := msg.(type) {
		case *qbtx.Transaction:
			tx := node.SignTranscation(msg)
			node.MsgBroadcast <- tx
		case *pbft.ReplyMsg:
			node.resolveTXreply(msg)
		}
	}
}

// genTranscationMsg,将用户输入的内容处理为交易信息结构
// 参数：用户输入的内容string
// 返回值：交易信息*block.Transaction
func (node *Node) SignTranscation(tx *qbtx.Transaction) *qbtx.Transaction {
	tx.SignTX(node.Node_name)
	file, _ := utils.Init_log(CLIENT_LOG_PATH + "resolve_" + node.Node_name + ".log")
	log.SetPrefix("[generate tx]")
	log.Println("get a tx and prepare to send it to the Primary")
	defer file.Close()

	return tx
}

func (node *Node) resolveTXreply(msg *pbft.ReplyMsg) {
	w := qbwallet.NewWallet(node.Node_name)
	for _, tx := range msg.Request.Transactions {
		if string(w.Addr) == tx.Vin[0].From {
			fmt.Println("success")
		}
	}
}
