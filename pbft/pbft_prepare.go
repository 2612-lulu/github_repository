package pbft

import (
	"bytes"
	"encoding/hex"
	"log"
	"qkdserv"
	"strconv"
	"uss"
	"utils"
)

// State.PrePare，进入准备阶段，从节点处理pre-prepare消息：从节点PrePrepareMsg——>各节点PrepareMsg
// 参数：预准备消息*PrePrepareMsg
// 返回值：准备消息*PrepareMsg
func (state *State) PrePare(preprepare *PrePrepareMsg) *PrepareMsg {
	state.Msg_logs.ReqMsg = preprepare.Request  // 将request消息提取出来记录到state中
	if !state.verifyPrePrepareMsg(preprepare) { // 校验受到的pre-prepare是否通过
		return nil
	} else {
		i, _ := strconv.ParseInt(qkdserv.Node_name[1:], 10, 64) // 获取节点编号
		prepare := &PrepareMsg{                                 // 定义一个prepare消息
			View:            preprepare.View,            // 获取视图号
			Sequence_number: preprepare.Sequence_number, // 获取索引号
			Digest_m:        preprepare.Digest_m,        // 获取消息摘要
			Node_i:          i,
			Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
				Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
					Sign_dev_id:  utils.GetNodeID(qkdserv.Node_name), // 签名者ID
					Sign_task_sn: uss.GenSignTaskSN(16),              // 签名序列号
				},
				Main_row_num: qkdserv.QKDSignRandomMainRowNum{
					Sign_node_name: qkdserv.Node_name, // 签名者节点号
					Main_row_num:   0,                 // 签名主行号，签名时默认为0
				},
				USS_counts:   uint32(N - 1), // 验签者的数量
				USS_unit_len: 16,            // 签名的单位长度，一般默认为16
			},
		}
		prepare.Sign_i.USS_message, _ = prepare.signMessageEncode() // 获取prepare阶段待签名消息
		// prepare消息的签名
		prepare.Sign_i = uss.UnconditionallySecureSign(prepare.Sign_i.Sign_index,
			prepare.Sign_i.USS_counts, prepare.Sign_i.USS_unit_len, prepare.Sign_i.USS_message)
		state.Msg_logs.PreparedMsgs[i] = prepare // 将节点自己产生的prepare消息写入log，以便后续进行投票校验
		//state.Current_stage = PrePrepared        // 此时状态改变为PrePrepared
		file, _ := utils.Init_log(utils.SIGN_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[PBFT-PREPARE    SIGN]")
		log.Println("Index of uss:", hex.EncodeToString(prepare.Sign_i.Sign_index.Sign_task_sn[:]))
		log.Println("plaintext:", hex.EncodeToString(prepare.Sign_i.USS_message))
		log.Println("signature:", hex.EncodeToString(prepare.Sign_i.USS_signature))
		log.Printf("sign of prepare message success\n\n\n")
		defer file.Close()
		return prepare
	}
}

// State.VerifyPrePrepareMsg,验证收到的pre-prepare消息的正确性
// 参数：预准备消息*PrePrepareMsg
// 返回值：验证结果bool
func (state *State) verifyPrePrepareMsg(preprepare *PrePrepareMsg) bool {
	var result bool
	msg := preprepare.Request.SerializeBlock()
	digest := utils.Digest(msg) // 计算消息的摘要值

	// 判断是否符合校验条件
	if state.View.ID != preprepare.View {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Prepare error]")
		defer file.Close()
		log.Println("the view of preprepare message is wrong!")
		result = false
	} else if state.Last_sequence_number >= preprepare.Sequence_number {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Prepare error]")
		defer file.Close()
		log.Println("the sequenceID of preprepare message is wrong!")
		result = false
	} else if !bytes.Equal(digest, preprepare.Digest_m) {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Prepare error]")
		defer file.Close()
		log.Println("the digest is wrong!")
		result = false
	} else if !uss.UnconditionallySecureVerifySign(preprepare.Request.Block_uss) {
		if !state.verifyRequestTX(preprepare.Request.Transactions) {
			file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
			log.SetPrefix("[Prepare error]")
			defer file.Close()
			log.Println("the client_sign is wrong!")
			result = false
		}
	} else if !uss.UnconditionallySecureVerifySign(preprepare.Sign_p) {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Prepare error]")
		defer file.Close()
		log.Println("the primary_sign is wrong!")
		result = false
	} else if !state.verifyRequestTX(preprepare.Request.Transactions) {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Prepare error]")
		defer file.Close()
		log.Println("the tx is wrong!")
		result = false
	} else {
		file, _ := utils.Init_log(utils.VERIFY_PATH + qkdserv.Node_name + ".log")
		defer file.Close()
		log.SetPrefix("[STAGE-Prepare:   VERIFY of BLOCK SIGN        ]")
		log.Println("Index of uss:", hex.EncodeToString(preprepare.Request.Block_uss.Sign_index.Sign_task_sn[:]))
		log.Printf("Verify of block sign success\n\n")
		log.SetPrefix("[STAGE-Prepare:   VERIFY of PrePrepareMsg SIGN]")
		log.Println("Index of uss:", hex.EncodeToString(preprepare.Sign_p.Sign_index.Sign_task_sn[:]))
		log.Printf("Verify of preprepare sign success\n\n")
		result = true
	}
	return result
}
