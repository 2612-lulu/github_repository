package qbnode

import (
	"log"
	"qb/qbtools"
	"qb/qbtx"
	"qb/qbutxo"
	"qb/qbwallet"
	"qb/quantumbc"
)

// 线程3：resolveMsg，用于对收到的信息作具体处理
func (node *Node) resolveMsg() {
	for {
		msg := <-node.MsgDelivery // 从调度器通道中获取缓存信息
		switch msg := msg.(type) {
		case *qbtx.ToGenTx:
			tx := node.genTranscationMsg(msg)
			node.MsgBroadcast <- tx
		}
	}
}

// genTranscationMsg,将用户输入的内容处理为交易信息结构
// 参数：用户输入的内容string
// 返回值：交易信息*block.Transaction
func (node *Node) genTranscationMsg(tx *qbtx.ToGenTx) *qbtx.Transaction {
	if !qbwallet.ValidateAddress(tx.From) { // 检验交易发送地址
		qbtools.Init_log(NODE_LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[resolve tx error]")
		log.Panic("ERROR: Sender address is not valid")
	}
	if !qbwallet.ValidateAddress(tx.To) { // 检验交易目的地址
		qbtools.Init_log(NODE_LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[resolve tx error]")
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := quantumbc.NewBlockchain(node.Node_name) // 获取账本
	UTXOSet := qbutxo.UTXOSet{                    // 设置utxo
		Blockchain: bc,
	}
	defer bc.DB.Close()
	transaction := qbutxo.NewUTXOTransaction(tx.From, tx.To, tx.Value, &UTXOSet)

	return transaction
}
