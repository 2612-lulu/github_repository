package pbft

import (
	"encoding/json"
	"fmt"
	"os"
	"qblock"
	"qkdserv"
	"testing"
	"utils"
)

func TestPBFTConsensus(t *testing.T) {
	fmt.Println("----------【pbft】----------------------------------------------------------------------")

	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	state := CreateState(1, -1) // 定义State消息
	F = 5
	N = 16
	// 定义区块消息，即请求消息
	file, _ := os.Open("../pbft/request.json") // 打开文件
	defer file.Close()                         // 关闭文件
	decoder := json.NewDecoder(file)
	var block qblock.Block
	err := decoder.Decode(&block) //Decode从输入流读取下一个json编码值并保存在v指向的值里
	if err != nil {
		panic(err)
	}
	request := &block
	utils.LogStage("	Request", false)

	qkdserv.Node_name = "P1"
	preprepare := state.PrePrePare(request)
	if preprepare != nil {
		utils.LogStage("	Pre-prepare", false)
	}
	// prepare
	qkdserv.Node_name = "P2"
	prepare := state.PrePare(preprepare)
	if prepare != nil {
		utils.LogStage("	prepare", false)
	}
	qkdserv.Node_name = "P3"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P4"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P5"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P6"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P7"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P8"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P9"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P10"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P11"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P12"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P13"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P14"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P15"
	state.PrePare(preprepare)
	qkdserv.Node_name = "P16"
	state.PrePare(preprepare)

	// commit
	qkdserv.Node_name = "P3"
	commit := state.Commit(prepare)
	if commit != nil {
		utils.LogStage("	Commit", false)
	}
	qkdserv.Node_name = "P4"
	state.Commit(prepare)
	qkdserv.Node_name = "P1"
	state.Commit(prepare)
	qkdserv.Node_name = "P5"
	state.Commit(prepare)
	qkdserv.Node_name = "P6"
	state.Commit(prepare)
	qkdserv.Node_name = "P7"
	state.Commit(prepare)
	qkdserv.Node_name = "P8"
	state.Commit(prepare)
	qkdserv.Node_name = "P9"
	state.Commit(prepare)
	qkdserv.Node_name = "P10"
	state.Commit(prepare)
	qkdserv.Node_name = "P11"
	state.Commit(prepare)
	qkdserv.Node_name = "P12"
	state.Commit(prepare)
	qkdserv.Node_name = "P13"
	state.Commit(prepare)
	qkdserv.Node_name = "P15"
	state.Commit(prepare)
	qkdserv.Node_name = "P15"
	state.Commit(prepare)
	qkdserv.Node_name = "P16"
	state.Commit(prepare)

	// reply
	qkdserv.Node_name = "P1"
	reply := state.Reply(commit)
	if reply != nil {
		utils.LogStage("	Reply", false)
	}
}
