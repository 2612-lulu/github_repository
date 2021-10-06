package qbcommand

import (
	"fmt"
	"log"
	"qb/qbutxo"
	"qb/qbwallet"
	"qb/quantumbc"
)

func (command *COMM) getBalance(address, nodeID string) {
	if !qbwallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := quantumbc.NewBlockchain("P1") // 获取当前全账本
	UTXOSet := qbutxo.UTXOSet{
		Blockchain: bc,
	}
	defer bc.DB.Close()

	balance := 0 // 定义余额
	UTXOs := UTXOSet.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
