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

//  State.ReplyMsg，获取reply消息，当收到2f+1个满足要求的commit时，调用此函数
// 参数：提交消息*CommitMsg
// 返回值：应答消息*ReplyMsg
func (state *State) Reply(commit *CommitMsg) *ReplyMsg {
	if !state.verifyCommitMsg(commit) {
		return nil
	} else if state.committed() {
		i, _ := strconv.ParseInt(qkdserv.Node_name[1:], 10, 64)
		_, ok := state.Msg_logs.ReplyMsgs[i] // 检查是否发送过reply消息,避免重复发送
		if !ok {                             // 如果未发送过reply消息
			reply := &ReplyMsg{
				View:        commit.View,                      // 获取视图号
				Time_stamp:  state.Msg_logs.ReqMsg.Time_stamp, // 相应request消息的时间戳
				Client_name: "",                               // 客户端名称
				Node_i:      i,                                // 获取节点编号
				Result:      true,
				Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
					Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
						Sign_dev_id:  utils.GetNodeID(qkdserv.Node_name), // 签名者ID
						Sign_task_sn: uss.GenSignTaskSN(16),              // 签名序列号
					},
					Main_row_num: qkdserv.QKDSignRandomMainRowNum{
						Sign_node_name:    qkdserv.Node_name, // 签名者节点号
						Main_row_num:      0,                 // 签名主行号，签名时默认为0
						Random_row_counts: 1,
						Random_unit_len:   16,
					},
					USS_counts:   1,  // 验签者的数量，客户端验签
					USS_unit_len: 16, // 签名的单位长度，一般默认为16
				},
				Request: state.Msg_logs.ReqMsg,
			}
			reply.Sign_i.USS_message, _ = reply.signMessageEncode()
			// reply消息的签名
			reply.Sign_i = uss.UnconditionallySecureSign(reply.Sign_i.Sign_index,
				reply.Sign_i.USS_counts, reply.Sign_i.USS_unit_len, reply.Sign_i.USS_message)

			state.Msg_logs.ReplyMsgs[i] = reply
			state.Current_stage = Committed

			file, _ := utils.Init_log(utils.SIGN_PATH + qkdserv.Node_name + ".log")
			log.SetPrefix("[PBFT-REPLY      SIGN]")
			log.Println("Index of uss:", hex.EncodeToString(reply.Sign_i.Sign_index.Sign_task_sn[:]))
			log.Println("plaintext:", hex.EncodeToString(reply.Sign_i.USS_message))
			log.Println("signature:", hex.EncodeToString(reply.Sign_i.USS_signature))
			log.Printf("sign of reply message success\n\n")
			defer file.Close()
			return reply
		}
	}
	return nil
}

// State.VerifyCommitMsg，验证收到的commit消息的正确性
// 参数：提交消息*CommitMsg
// 返回值：验证结果bool
func (state *State) verifyCommitMsg(commit *CommitMsg) bool {
	var result bool
	msg := state.Msg_logs.ReqMsg.SerializeBlock()
	digest := utils.Digest(msg) // 计算消息的摘要值

	if state.View.ID != commit.View {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Reply error]")
		defer file.Close()
		log.Println("the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 && state.Last_sequence_number >= commit.Sequence_number {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Reply error]")
		defer file.Close()
		log.Println("the sequenceID of commit message is wrong!")
		result = false
	} else if !bytes.Equal(digest, commit.Digest_m) {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Reply error]")
		defer file.Close()
		log.Println("the verify of digest is wrong!")
		result = false
	} else if !uss.UnconditionallySecureVerifySign(commit.Sign_i) {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Reply error]")
		defer file.Close()
		log.Println("the node_sign is wrong!")
		result = false
	} else {
		file, _ := utils.Init_log(utils.VERIFY_PATH + qkdserv.Node_name + ".log")
		defer file.Close()
		log.SetPrefix("[STAGE-Commit:    VERIFY of CommitMsg SIGN    ]")
		log.Println("Index of uss:", hex.EncodeToString(commit.Sign_i.Sign_index.Sign_task_sn[:]))
		log.Printf("Verify of commit sign success\n\n\n")
		state.Msg_logs.CommittedMsgs[commit.Node_i] = commit
		result = true
	}
	return result
}

// State.commited,验证收到的commit消息是否符合要求
// 参数：共识状态state
// 返回值：验证结果bool
func (state *State) committed() bool {
	if !state.prepared() { // 如果prepare投票未通过，则不能进入commit
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[committed error]")
		defer file.Close()
		log.Println("didn't prepared!")
		return false
	}
	if len(state.Msg_logs.CommittedMsgs) < 2*F+1 { // commit通过的条件是受到2f+1个校验通过的commit,包括自身节点
		return false
	}
	return true
}
