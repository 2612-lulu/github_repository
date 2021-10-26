package pbft

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"qblock"
	"qkdserv"
	"reflect"
	"strconv"
	"uss"
	"utils"
)

//  State.Commit，所有联盟节点处理收到的prepare消息：各节点prepare——>其余节点commit
// 参数：准备消息*PrepareMsg
// 返回值：提交消息*CommitMsg
func (state *State) Commit(prepare *PrepareMsg) *CommitMsg {
	if !state.verifyPrepareMsg(prepare) { // 校验收到的prepare消息
		return nil
	} else if state.prepared() { // 检查是否收到2f+1（含本节点产生的prepare）个通过校验的prepare消息
		i, _ := strconv.ParseInt(qkdserv.Node_name[1:], 10, 64)
		_, ok := state.Msg_logs.CommittedMsgs[i] // 检查是否发送过commit消息
		if !ok {
			commit := &CommitMsg{ // 定义一个commit消息
				View:            prepare.View,            // 获取视图号
				Sequence_number: prepare.Sequence_number, // 获取索引号
				Digest_m:        prepare.Digest_m,        // 获取消息摘要
				Node_i:          i,                       // 获取节点编号
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
			commit.Sign_i.USS_message, _ = commit.signMessageEncode() // 获取commit阶段待签名消息
			// commit消息的签名
			commit.Sign_i = uss.UnconditionallySecureSign(commit.Sign_i.Sign_index,
				commit.Sign_i.USS_counts, commit.Sign_i.USS_unit_len, commit.Sign_i.USS_message)

			state.Msg_logs.CommittedMsgs[i] = commit // 将commit写入log，以便后续投票校验
			//state.Current_stage = Prepared           // 此时状态改变为Prepared

			file, _ := utils.Init_log(utils.SIGN_PATH + qkdserv.Node_name + ".log")
			log.SetPrefix("[PBFT-COMMIT     SIGN]")
			log.Println("Index of uss:", hex.EncodeToString(commit.Sign_i.Sign_index.Sign_task_sn[:]))
			log.Println("plaintext:", hex.EncodeToString(commit.Sign_i.USS_message))
			log.Println("signature:", hex.EncodeToString(commit.Sign_i.USS_signature))
			log.Printf("sign of commit message success\n\n")
			defer file.Close()
			return commit
		}
	}
	return nil
}

// State.VerifyPrepareMsg，验证收到的prepare消息的正确性
// 参数：准备消息*PrepareMsg
// 返回值：验证结果bool
func (state *State) verifyPrepareMsg(prepare *PrepareMsg) bool {
	var result bool
	msg := state.Msg_logs.ReqMsg.SerializeBlock()
	digest := utils.Digest(msg) // 计算消息的摘要值

	if state.View.ID != prepare.View {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Commit error]")
		defer file.Close()
		log.Println("the view of prepare message is wrong!")
		result = false
	} else if state.Last_sequence_number >= prepare.Sequence_number {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Commit error]")
		defer file.Close()
		log.Println("the sequenceID of prepare message is wrong!")
		result = false
	} else if !bytes.Equal(digest, prepare.Digest_m) {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Commit error]")
		defer file.Close()
		log.Println("the verify of digest is wrong!")
		result = false
	} else if !uss.UnconditionallySecureVerifySign(prepare.Sign_i) {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Commit error]")
		defer file.Close()
		log.Println("the node_sign is wrong!")
		result = false
	} else {
		file, _ := utils.Init_log(utils.VERIFY_PATH + qkdserv.Node_name + ".log")
		defer file.Close()
		log.SetPrefix("[STAGE-Commit:    VERIFY of PrepareMsg SIGN   ]")
		log.Println("Index of uss:", hex.EncodeToString(prepare.Sign_i.Sign_index.Sign_task_sn[:]))
		log.Printf("Verify of prepare sign success\n\n")
		state.Msg_logs.PreparedMsgs[prepare.Node_i] = prepare
		result = true
	}
	return result
}

// State.prepared,验证收到的prepare消息是否符合要求
// 参数：共识状态state
// 返回值：验证结果bool
func (state *State) prepared() bool {
	if reflect.DeepEqual(state.Msg_logs.ReqMsg, qblock.Block{}) {
		fmt.Println("request of state is nil")
		return false
	}
	if len(state.Msg_logs.PreparedMsgs) < 2*F {
		/*file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[prepared error]")
		defer file.Close()
		log.Println("didn't receive 2*f prepared message,please wait")*/
		return false
	}
	return true
}
