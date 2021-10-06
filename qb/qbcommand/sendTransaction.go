package qbcommand

import (
	"fmt"
	"log"
	"qb/qbcli"
	"qb/qbutxo"
	"qb/qbwallet"
	"qb/quantumbc"
	"utils"
)

func (command *COMM) transaction(tx_from, tx_to, nodeID string, tx_amount int) {
	client := qbcli.NewClient(nodeID) // 开启节点
	file, _ := utils.Init_log(qbcli.CLIENT_LOG_PATH + "resolve_" + client.Client_name + ".log")
	log.SetPrefix("[resolve tx error]")
	defer file.Close()
	if !qbwallet.ValidateAddress(tx_from) { // 检验交易发送地址
		log.Panic("ERROR: Sender address is not valid")
	}
	if !qbwallet.ValidateAddress(tx_to) { // 检验交易目的地址
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := quantumbc.NewBlockchain(client.View.Primary) // 获取账本
	UTXOSet := qbutxo.UTXOSet{                         // 设置utxo
		Blockchain: bc,
	}
	transaction := qbutxo.NewUTXOTransaction(tx_from, tx_to, client.Client_name, tx_amount, &UTXOSet)
	bc.DB.Close() // 关闭数据库
	transaction.PrintTransaction()

	client.MsgBroadcast <- transaction
	fmt.Println("put into channel")
	client.Httplisten() // 开启http

}
