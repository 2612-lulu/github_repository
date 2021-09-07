package main

import (
	"fmt"
	"qb/block"
	"qb/mylog"
	"qb/pbft"
	"qb/qbtools"
	"qb/qkdserv"
	"qb/uss"
	"testing"
	"time"
)

func TestPBFTConsensus(t *testing.T) {
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	qkdserv.Node_name = [2]byte{'C', '1'}

	// 定义State消息
	state := pbft.CreateState(1, -1)

	// 定义单条交易消息
	m := []byte("from c1 to P1 36576")
	tran := &block.Transaction{
		Time_stamp: time.Now().UnixNano(),
		Name:       qkdserv.Node_name, // 签名者是C1
		TransactionOperation: block.TransactionOperation{
			Transaction_message: m,
			Digest_m:            qbtools.Digest(m),
		},
		Sign_client: uss.USSToeplitzHashSignMsg{ // 签名信息
			Sign_index: qkdserv.QKDSignMatrixIndex{
				Sign_dev_id:  qbtools.GetNodeIDTable(qkdserv.Node_name),
				Sign_task_sn: uss.GenSignTaskSN(16),
			},
			Main_row_num: qkdserv.QKDSignRandomMainRowNum{
				Sign_Node_Name: qkdserv.Node_name,
				Main_Row_Num:   0, // 签名主行号，签名时默认为0
			},
			Sign_counts: 4,
			Sign_len:    16,
		},
	}
	tran.Sign_client.Message, _ = tran.SignMessageEncode()
	tran.Sign_client = uss.Sign(tran.Sign_client.Sign_index, tran.Sign_client.Sign_counts,
		tran.Sign_client.Sign_len, tran.Sign_client.Message)

	// 定义区块消息，即请求消息
	msgs := make([]*block.Transaction, 0)
	msgs = append(msgs, tran)
	request := block.Block{
		Time_stamp:   time.Now().UnixNano(),
		Transactions: msgs,
	}
	//log.LogStage("Request", false)

	fmt.Println("-----------------------【pbft共识】PrePrepare-------------------------------------------------")
	qkdserv.Node_name = [2]byte{'P', '1'}
	preprepare, err := state.PrePrePare(&request)
	if err == nil {
		mylog.LogStage("	Request", true)
		//fmt.Println(len(preprepare.Sign_p.Message))
		mylog.LogStage("	Pre-prepare", false)
	}
	fmt.Println("-----------------------【pbft共识】Prepare----------------------------------------------------")
	qkdserv.Node_name = [2]byte{'P', '2'}
	prepare, err := state.PrePare(preprepare)
	if err == nil {
		mylog.LogStage("	Pre-prepare", true)
		mylog.LogStage("	prepare", false)
	}
	qkdserv.Node_name = [2]byte{'P', '3'}
	_, err = state.PrePare(preprepare)
	if err == nil {
		mylog.LogStage("	Pre-prepare", true)
		mylog.LogStage("	prepare", false)
	}
	qkdserv.Node_name = [2]byte{'P', '4'}
	_, err = state.PrePare(preprepare)
	if err == nil {
		mylog.LogStage("	Pre-prepare", true)
		mylog.LogStage("	prepare", false)
	}
	fmt.Println("-----------------------【pbft共识】Commit-----------------------------------------------------")
	qkdserv.Node_name = [2]byte{'P', '3'}
	commit, _ := state.Commit(prepare)
	if commit != nil {
		mylog.LogStage("	Prepare", true)
		mylog.LogStage("	Commit", false)
	}
	qkdserv.Node_name = [2]byte{'P', '4'}
	_, err = state.Commit(prepare)
	if err == nil {
		mylog.LogStage("	Prepare", true)
		mylog.LogStage("	Commit", false)
	}
	qkdserv.Node_name = [2]byte{'P', '1'}
	_, err = state.Commit(prepare)
	if err == nil {
		mylog.LogStage("	Prepare", true)
		mylog.LogStage("	Commit", false)
	}

	fmt.Println("-----------------------【pbft共识】Reply------------------------------------------------------")
	qkdserv.Node_name = [2]byte{'P', '1'}
	reply, _ := state.Reply(commit)
	if reply != nil {
		mylog.LogStage("	Commit", true)
		mylog.LogStage("	Reply", false)
	}
}
