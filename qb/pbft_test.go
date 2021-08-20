package main

import (
	"fmt"
	"testing"
)

func TestPBFTConsensus(t *testing.T) {
	fmt.Println("-----------------------【pbft共识】PrePrepare-------------------------------------------------")
	/*qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	qkdserv.Node_name = [2]byte{'C', '1'} // 签名者是C1
	// 定义State消息
	state := pbft.State{}
	state.View = 1
	state.CurrentStage = 2347
	o := "gbsdaifgiafgiegfgibv"
	// 生成request消息
	request, _ := state.GenReqMsg(o, qkdserv.Node_name)
	*/
	fmt.Println("-----------------------【pbft共识】Prepare----------------------------------------------------")
	fmt.Println("-----------------------【pbft共识】Commit-----------------------------------------------------")
	fmt.Println("-----------------------【pbft共识】Reply------------------------------------------------------")
}
