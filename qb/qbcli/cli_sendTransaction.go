package qbcli

import (
	"fmt"
	"log"
	"qb/qbutxo"
	"qb/qbwallet"
	"qb/quantumbc"
)

func (cli *CLI) sendTX(tx_from, tx_to, nodeID string, tx_amount int) {
	bc := quantumbc.NewBlockchain(nodeID)
	UTXOSet := qbutxo.UTXOSet{
		Blockchain: bc,
	}
	defer bc.DB.Close()
	w := qbwallet.NewWallet()
	log.Println("success to read current blockchain")
	UTXOSet.Reindex()
	log.Println("reindex utxo")
	tx := qbutxo.NewUTXOTransaction(w, tx_to, tx_amount, &UTXOSet)
	fmt.Println(tx.ID)
	fmt.Println("Success!")
}
