package main

import (
	"encoding/hex"
	"fmt"
	"qb/pbft"
	"qb/qkdserv"
	"qb/uss"
	"testing"
)

func TestPBFTConsensus(t *testing.T) {
	fmt.Println("-----------------------【pbft共识】PrePrepare-------------------------------------------------")
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)

	qkdserv.Node_name = [2]byte{'C', '1'} // 签名者是C1
	// 定义State消息
	state := pbft.State{}
	state.View = 1
	state.CurrentStage = 0
	o := "gbsdaifgiafgiegfgibv"
	// 生成request消息
	request, _ := state.GenReqMsg(o, qkdserv.Node_name)

	// 获取pre-prepare消息
	qkdserv.Node_name = [2]byte{'P', '1'} // P1是主节点
	preprepare, ok := state.PrePrePare(request)
	if ok {
		fmt.Print("\nview=", preprepare.View)
		fmt.Print("\nn=", preprepare.Sequence_number)
		fmt.Print("\nd=", preprepare.Digest_m)
		fmt.Print("\nsign=", hex.EncodeToString(preprepare.Sign_p.Sign))
		fmt.Println("	request 校验成功")

	}
	fmt.Println("-----------------------【pbft共识】Prepare----------------------------------------------------")
	qkdserv.Node_name = [2]byte{'P', '2'} // P2是从节点
	prepare, ok := state.PrePare(preprepare)
	if ok {
		fmt.Print("view=", prepare.View)
		fmt.Print("\nn=", prepare.Sequence_number)
		fmt.Print("\nd=", prepare.Digest_m)
		fmt.Print("\ni=", prepare.Node_i)
		fmt.Print("\nsign=", hex.EncodeToString(prepare.Sign_i.Sign))
		fmt.Println("	preprepare校验成功")
	}
	fmt.Println("-----------------------【pbft共识】Commit-----------------------------------------------------")
	qkdserv.Node_name = [2]byte{'P', '3'} // P3是从节点
	var commit *pbft.CommitMsg
	if state.VerifyPrepareMsg(prepare) {
		commit = state.GetCommitMsg(prepare)
		//fmt.Print("sign=", hex.EncodeToString(commit.Sign_i.Sign))
		fmt.Println("	prepare校验成功")
	}
	fmt.Println("-----------------------【pbft共识】Reply------------------------------------------------------")
	qkdserv.Node_name = [2]byte{'P', '4'} // P3是从节点
	if state.VerifyCommitMsg(commit) {
		fmt.Println("	commit校验成功")
	}
	qkdserv.Node_name = [2]byte{'P', '1'}
	reply := state.GetReplyMsg(commit)

	qkdserv.Node_name = [2]byte{'C', '1'}
	if uss.VerifySign(reply.Sign_i) { // 校验上一阶段节点签名
		fmt.Println("	验签成功")
	}
}
