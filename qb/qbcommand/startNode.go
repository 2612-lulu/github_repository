package qbcommand

import (
	"fmt"
	"log"
	"qb/qbnode"
	"qb/qbutxo"
	"qb/qbwallet"
	"qb/quantumbc"
)

func (command *COMM) startNode(nodeID string) {
	node := qbnode.NewNode(nodeID) // 开启一个联盟节点
	addresses := make([]string, 0)
	for ID, _ := range node.Node_table { // 钱包地址
		w := qbwallet.NewWallet(ID)
		node.Addr_table[string(w.Addr)] = ID
		addresses = append(addresses, string(w.Addr))
	}
	//fmt.Println(addresses)
	// 获取区块链数据库名称
	dbFile := fmt.Sprintf(quantumbc.DBFile, nodeID)
	// 检查是否已创建数据库，如未创建则现在创建
	qbc := quantumbc.Blockchain{}
	if !quantumbc.DBExists(dbFile) {
		qbc = *quantumbc.CreateBlockchain(addresses, nodeID)

		// utxo初始化
		UTXOSet := qbutxo.UTXOSet{
			Blockchain: &qbc,
		}
		UTXOSet.Reindex()
		log.Println("Blockchain didn't exists，have create a new one.")
		qbc.DB.Close() // 关闭账本
	} else {
		log.Println("Blockchain already exists.")
	}
	//quantumbc.PrintBlockChain(nodeID) // 打印当前区块链信息
	node.Httplisten()
}
