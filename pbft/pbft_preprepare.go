package pbft

import (
	"encoding/hex"
	"log"
	"qblock"
	"qbtx"
	"qkdserv"
	"time"
	"uss"
	"utils"
)

// State.PrePrePare，进入共识，由主节点进行消息处理：客户端Request——>主节点PrePrePare——>从节点
// 参数：请求消息*qblock.Block
// 返回值：预准备消息*PrePrepareMsg
func (state *State) PrePrePare(request *qblock.Block) *PrePrepareMsg {
	state.Msg_logs.ReqMsg = request // 记录request消息到state的log中
	msg := request
	if state.verifyRequestTX(msg.Transactions) { // 如果每条交易信息验签成功
		sequenceID := time.Now().UnixNano() // 使用时间戳作为暂时序列号
		if state.Last_sequence_number != -1 {
			sequenceID = state.Last_sequence_number + 1 // 主节点每开始一次共识，序列号+1
		}
		digest_msg := msg.SerializeBlock()
		// 定义一个preprepare消息
		preprepare := &PrePrepareMsg{
			View:            state.View.ID,            // 获取视图号
			Sequence_number: sequenceID,               // 为其分配序列号
			Digest_m:        utils.Digest(digest_msg), // 交易信息摘要
			Sign_p: uss.USSToeplitzHashSignMsg{ // 签名信息
				Sign_index: qkdserv.QKDSignMatrixIndex{
					Sign_dev_id:  utils.GetNodeID(qkdserv.Node_name),
					Sign_task_sn: uss.GenSignTaskSN(16),
				},
				Main_row_num: qkdserv.QKDSignRandomMainRowNum{
					Sign_node_name:    qkdserv.Node_name,
					Main_row_num:      0, // 签名主行号，签名时默认为0
					Random_row_counts: uint32(N - 1),
					Random_unit_len:   16,
				},
				USS_counts:   uint32(N - 1),
				USS_unit_len: 16,
			},
			Request: nil, // 将请求消息附在preprepare中广播给所有从节点
		}
		preprepare.Request = request
		preprepare.Sign_p.USS_message, _ = preprepare.signMessageEncode()
		preprepare.Sign_p = uss.UnconditionallySecureSign(preprepare.Sign_p.Sign_index,
			preprepare.Sign_p.USS_counts, preprepare.Sign_p.USS_unit_len, preprepare.Sign_p.USS_message)
		state.Current_stage = PrePrepared

		file, _ := utils.Init_log(utils.SIGN_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[PBFT-PREPREPARE SIGN]")
		log.Println("Index of uss:", hex.EncodeToString(preprepare.Sign_p.Sign_index.Sign_task_sn[:]))
		log.Println("plaintext:", hex.EncodeToString(preprepare.Sign_p.USS_message))
		log.Println("signature:", hex.EncodeToString(preprepare.Sign_p.USS_signature))
		log.Printf("sign of preprepare message success\n\n\n")
		defer file.Close()
		return preprepare
	} else {
		return nil
	}
}

// State.VerifyRequestTX，验证请求消息中每条交易信息的正确性
// 参数：交易消息[]*qbtx.Transaction
// 返回值：验证结果bool
func (state *State) verifyRequestTX(txs []*qbtx.Transaction) bool {

	//TODO:验证每条交易信息的签名
	verify_num := 0
	for _, tx := range txs {
		if tx.VerifyUSSTransactionSign() { // 验证签名正确性
			file, _ := utils.Init_log(utils.VERIFY_PATH + qkdserv.Node_name + ".log")
			log.SetPrefix("[STAGE-PrePrepare/Prepare:VERIFY of Transaction SIGN]")
			log.Println("transaciton ID:", hex.EncodeToString(tx.TX_id))
			defer file.Close()
			for _, vin := range tx.TX_vin {
				log.Println("Index of uss:", hex.EncodeToString(vin.TX_uss_sign.Sign_index.Sign_task_sn[:]))
			}
			log.Printf("Verify of transaction sign success\n\n")

			verify_num++
		}
	}
	if verify_num == len(txs) {
		return true
	} else {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Pre-prepare error]")
		log.Println("The verify of client sign is false!!!")
		defer file.Close()
		return false
	}
}
