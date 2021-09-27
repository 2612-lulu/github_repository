package qbcommand

import (
	"qb/qbcli"
	"qb/qbtx"
)

func (command *COMM) genTX(tx_from, tx_to, nodeID string, tx_amount int) {
	client := qbcli.NewClient(nodeID)
	tx := qbtx.ToGenTx{
		From:  tx_from,
		To:    tx_to,
		Value: tx_amount,
	}
	client.MsgBroadcast <- &tx
	client.Httplisten() // 开启http
}
