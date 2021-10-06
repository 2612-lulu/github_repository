package pbft

import (
	"fmt"
	"qblock"
	"qbtx"
	"qkdserv"
	"testing"
	"uss"
	"utils"
)

func TestPBFTConsensus(t *testing.T) {
	fmt.Println("----------【3.pbft】----------------------------------------------------------------------")

	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	qkdserv.Node_name = "C1"
	txInput := qbtx.TXInput{
		Txid:      []byte("egry"),
		Vout:      1,
		Signature: uss.USSToeplitzHashSignMsg{},
		From:      "C1",
	}
	var Inputs []qbtx.TXInput
	Inputs = append(Inputs, txInput)
	txOutput := qbtx.TXOutput{
		Value: 1,
		To:    "P3",
	}
	var outputs []qbtx.TXOutput
	outputs = append(outputs, txOutput)
	tx := qbtx.Transaction{
		ID:   nil,
		Vin:  Inputs,
		Vout: outputs,
	}
	tx.SetID()
	tx.SignTX("C1")
	// 定义State消息
	state := CreateState(1, -1)

	// 定义区块消息，即请求消息
	msgs := make([]*qbtx.Transaction, 0)
	msgs = append(msgs, &tx)
	request := qblock.NewBlock(msgs, nil, 1)
	//log.LogStage("Request", false)

	//fmt.Println("-----------------------【3.pbft共识】PrePrepare-------------------------------------------------")
	qkdserv.Node_name = "P1"
	preprepare, err := state.PrePrePare(request)
	if err == nil {
		//qbtools.LogStage("	Request", true)
		//fmt.Println(len(preprepare.Sign_p.Message))
		utils.LogStage("	Pre-prepare", false)
	}
	//fmt.Println("-----------------------【3.pbft共识】Prepare----------------------------------------------------")
	qkdserv.Node_name = "P2"
	prepare, err := state.PrePare(preprepare)
	if err == nil {
		//qbtools.LogStage("	Pre-prepare", true)
		utils.LogStage("	prepare", false)
	}
	qkdserv.Node_name = "P3"
	_, err = state.PrePare(preprepare)
	if err == nil {
		//qbtools.LogStage("	Pre-prepare", true)
		utils.LogStage("	prepare", false)
	}
	qkdserv.Node_name = "P4"
	_, err = state.PrePare(preprepare)
	if err == nil {
		//qbtools.LogStage("	Pre-prepare", true)
		utils.LogStage("	prepare", false)
	}
	//fmt.Println("-----------------------【3.pbft共识】Commit-----------------------------------------------------")
	qkdserv.Node_name = "P3"
	commit, _ := state.Commit(prepare)
	if commit != nil {
		//qbtools.LogStage("	Prepare", true)
		utils.LogStage("	Commit", false)
	}
	qkdserv.Node_name = "P4"
	_, err = state.Commit(prepare)
	if err == nil {
		//qbtools.LogStage("	Prepare", true)
		utils.LogStage("	Commit", false)
	}
	qkdserv.Node_name = "P1"
	_, err = state.Commit(prepare)
	if err == nil {
		//qbtools.LogStage("	Prepare", true)
		utils.LogStage("	Commit", false)
	}

	//fmt.Println("-----------------------【3.pbft共识】Reply------------------------------------------------------")
	qkdserv.Node_name = "P1"
	reply, _ := state.Reply(commit)
	if reply != nil {
		//qbtools.LogStage("	Commit", true)
		utils.LogStage("	Reply", false)
	}
}
