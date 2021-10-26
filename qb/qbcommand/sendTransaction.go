package qbcommand

import (
	"encoding/hex"
	"log"
	"qb/qbnode"
	"qb/qbutxo"
	"qb/qbwallet"
	"qb/quantumbc"
	"utils"
)

func (command *COMM) transaction(tx_from, tx_to, nodeID string, tx_amount int) {
	node := qbnode.NewNode(nodeID) // 开启节点
	file, _ := utils.Init_log(qbnode.NODE_LOG_PATH + node.Node_name + ".log")
	log.SetPrefix("[resolve tx error]")
	defer file.Close()
	if !qbwallet.ValidateAddress(tx_from) { // 检验交易发送地址
		log.Panic("ERROR: Sender address is not valid")
	}
	if !qbwallet.ValidateAddress(tx_to) { // 检验交易目的地址
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := quantumbc.NewBlockchain(node.Primary) // 获取账本
	UTXOSet := qbutxo.UTXOSet{                  // 设置utxo
		Blockchain: bc,
	}
	transaction := qbutxo.NewUTXOTransaction(tx_from, tx_to, node.Node_name, tx_amount, &UTXOSet)
	bc.DB.Close() // 关闭数据库
	transaction.PrintTransaction()
	file, _ = utils.Init_log(utils.SIGN_PATH + nodeID + ".log")
	log.SetPrefix("[TRANSACTION SIGN]")
	log.Println("transaciton ID:", hex.EncodeToString(transaction.TX_id))
	defer file.Close()
	for _, vin := range transaction.TX_vin {
		log.Println("Index of uss:", hex.EncodeToString(vin.TX_uss_sign.Sign_index.Sign_task_sn[:]))
		log.Println("plaintext:", hex.EncodeToString(vin.TX_uss_sign.USS_message))
		log.Println("signature:", hex.EncodeToString(vin.TX_uss_sign.USS_signature))
	}
	log.Printf("Sign of transaction success\n\n")

	node.MsgBroadcast <- transaction
	node.Httplisten() // 开启http
}
