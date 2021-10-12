package qbnode

import (
	"fmt"
	"log"
	"pbft"
	"qb/qbutxo"
	"qb/qbwallet"
	"qb/quantumbc"
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
	tx.USSTransactionSign(node.Node_name)
	file, _ := utils.Init_log(NODE_LOG_PATH + "resolve_" + node.Node_name + ".log")
	log.SetPrefix("[generate tx]")
	log.Println("get a tx and prepare to send it to the Primary")
	defer file.Close()

	return tx
}

func (node *Node) resolveTXreply(msg *pbft.ReplyMsg) {
	w := qbwallet.NewWallet(node.Node_name)
	for _, tx := range msg.Request.Transactions {
		if string(w.Addr) == tx.TX_vin[0].TX_src {
			fmt.Println("transaction success")
			// 获取余额
			bc := quantumbc.NewBlockchain("P1") // 获取当前全账本
			UTXOSet := qbutxo.UTXOSet{
				Blockchain: bc,
			}
			defer bc.DB.Close()

			balance := 0 // 定义余额
			UTXOs := UTXOSet.FindUTXO(string(w.Addr))

			for _, out := range UTXOs {
				balance += out.TX_value
			}
			fmt.Printf("Balance of '%s': %d\n", w.Addr, balance)
		}
	}
}
