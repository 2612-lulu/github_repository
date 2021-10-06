// pbft包，实现了pbft共识的相关处理
// 创建人：zhanglu
// 创建时间：2021/08/15
package pbft

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"qblock"
	"qbtx"
	"qkdserv"
	"strconv"
	"time"
	"uss"
	"utils"
)

// pbft状态标识
type State struct {
	View                 int64      // 视图号
	Msg_logs             MsgLogs    // 缓存数据
	Last_sequence_number int64      // 上次共识序列号
	Current_stage        Stage      // 当前状态
	CommittedMessage     *CommitMsg // 达成共识的commit消息
}

// pbft缓存数据，用于存放pbft过程中的各类消息
type MsgLogs struct {
	ReqMsg        *qblock.Block         // 存放request消息
	PreparedMsgs  map[int64]*PrepareMsg // 存放prepared消息
	CommittedMsgs map[int64]*CommitMsg  // 存放committed消息
	ReplyMsgs     map[int64]*ReplyMsg   // 存放Reply消息
}

// 视图号
type View struct {
	ID      int64  // 视图号
	Primary string // 主节点
}

type Stage int

// 状态标识
const (
	Idle        Stage = iota // Idle=0，节点已成功创建，但共识过程尚未启动。
	PrePrepared              // PrePrepared=1，RequestMsg已成功处理。节点已准备好进入prepare阶段。
	Prepared                 // Prepared=2，Same with `prepared` stage explained in the original paper.
	Committed                // Committed=3，Same with `committed-local` stage explained in the original paper.
)

// 错误log存放路径
const LOG_ERROR_PATH = "../pbft/errorlog/error_"

// CreateState，创建共识状态。如果不存在lastSequenceNumber，则lastSequenceNumber=-1
// 参数：视图号int64，上次共识的序列号int64
// 返回值：pbft状态State
func CreateState(view int64, lastSequenceNumber int64) *State {
	return &State{
		View: view, // 当前视图号，为主节点编号
		Msg_logs: MsgLogs{ // 初始化
			ReqMsg:        new(qblock.Block),
			PreparedMsgs:  make(map[int64]*PrepareMsg),
			CommittedMsgs: make(map[int64]*CommitMsg),
			ReplyMsgs:     make(map[int64]*ReplyMsg),
		},
		CommittedMessage:     nil,
		Last_sequence_number: lastSequenceNumber, // 上一个序列号
		Current_stage:        Idle,               // 目前状态，节点创立，即将进入共识
	}
}

// State.PrePrePare，进入共识，由主节点进行消息处理：客户端Request——>主节点PrePrePare——>从节点
// 参数：请求消息*block.Block
// 返回值：预准备消息*PrePrepareMsg，处理错误error
func (state *State) PrePrePare(request *qblock.Block) (*PrePrepareMsg, error) {
	state.Msg_logs.ReqMsg = request // 记录request消息到state的log中
	msg := request
	if state.verifyRequest(msg.Transactions) { // 如果每条交易信息验签成功
		sequenceID := time.Now().UnixNano() // 使用时间戳作为暂时序列号
		if state.Last_sequence_number != -1 {
			sequenceID = state.Last_sequence_number + 1 // 主节点每开始一次共识，序列号+1
		}
		digest_msg, _ := json.Marshal(msg)
		// 定义一个preprepare消息
		preprepare := &PrePrepareMsg{
			View:            state.View,               // 获取视图号
			Sequence_number: sequenceID,               // 为其分配序列号
			Digest_m:        utils.Digest(digest_msg), // 交易信息摘要
			Sign_p: uss.USSToeplitzHashSignMsg{ // 签名信息
				Sign_index: qkdserv.QKDSignMatrixIndex{
					Sign_dev_id:  utils.GetNodeIDTable(qkdserv.Node_name),
					Sign_task_sn: uss.GenSignTaskSN(16),
				},
				Main_row_num: qkdserv.QKDSignRandomMainRowNum{
					Sign_Node_Name: qkdserv.Node_name,
					Main_Row_Num:   0, // 签名主行号，签名时默认为0
				},
				Sign_counts: N - 1,
				Sign_len:    16,
			},
			Request: new(qblock.Block), // 将请求消息附在preprepare中广播给所有从节点
		}
		preprepare.Request = request
		preprepare.Sign_p.Message = preprepare.signMessageEncode()
		preprepare.Sign_p = uss.Sign(preprepare.Sign_p.Sign_index, preprepare.Sign_p.Sign_counts,
			preprepare.Sign_p.Sign_len, preprepare.Sign_p.Message)
		state.Current_stage = PrePrepared
		return preprepare, nil
	} else {
		return nil, errors.New("request message is corrupted")
	}
}

// State.VerifyRequest，验证请求消息中每条交易信息的正确性
// 参数：交易消息[]*block.Transaction
// 返回值：验证结果bool
func (state *State) verifyRequest(request []*qbtx.Transaction) bool {
	//TODO:验证每条交易信息的签名
	verify_num := 0
	for _, reqMsg := range request {
		if reqMsg.VerifyTX() { // 验证签名正确性
			verify_num++
		}
	}
	if verify_num == len(request) {
		return true
	} else {
		file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
		log.SetPrefix("[Pre-prepare error]")
		log.Println("The verify of client sign is false!!!")
		defer file.Close()
		return false
	}
}

// State.PrePare，进入准备阶段，从节点处理pre-prepare消息：从节点PrePrepareMsg——>各节点PrepareMsg
// 参数：预准备消息*PrePrepareMsg
// 返回值：准备消息*PrepareMsg，处理错误error
func (state *State) PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, error) {
	state.Msg_logs.ReqMsg = preprepare.Request  // 将request消息提取出来记录到state中
	if !state.verifyPrePrepareMsg(preprepare) { // 校验受到的pre-prepare是否通过
		return nil, errors.New("preprepare message is corrupted")
	} else {
		i, _ := strconv.ParseInt(qkdserv.Node_name[1:], 10, 64) // 获取节点编号
		prepare := &PrepareMsg{                                 // 定义一个prepare消息
			View:            preprepare.View,            // 获取视图号
			Sequence_number: preprepare.Sequence_number, // 获取索引号
			Digest_m:        preprepare.Digest_m,        // 获取消息摘要
			Node_i:          i,
			Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
				Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
					Sign_dev_id:  utils.GetNodeIDTable(qkdserv.Node_name), // 签名者ID
					Sign_task_sn: uss.GenSignTaskSN(16),                   // 签名序列号
				},
				Main_row_num: qkdserv.QKDSignRandomMainRowNum{
					Sign_Node_Name: qkdserv.Node_name, // 签名者节点号
					Main_Row_Num:   0,                 // 签名主行号，签名时默认为0
				},
				Sign_counts: N - 1, // 验签者的数量
				Sign_len:    16,    // 签名的单位长度，一般默认为16
			},
		}
		prepare.Sign_i.Message, _ = prepare.signMessageEncode() // 获取prepare阶段待签名消息
		// prepare消息的签名
		prepare.Sign_i = uss.Sign(prepare.Sign_i.Sign_index,
			prepare.Sign_i.Sign_counts, prepare.Sign_i.Sign_len, prepare.Sign_i.Message)
		state.Msg_logs.PreparedMsgs[i] = prepare // 将节点自己产生的prepare消息写入log，以便后续进行投票校验
		state.Current_stage = PrePrepared        // 此时状态改变为PrePrepared
		return prepare, nil
	}
}

// State.VerifyPrePrepareMsg,验证收到的pre-prepare消息的正确性
// 参数：预准备消息*PrePrepareMsg
// 返回值：验证结果bool
func (state *State) verifyPrePrepareMsg(preprepare *PrePrepareMsg) bool {
	var result bool
	msg, _ := json.Marshal(preprepare.Request)
	digest := utils.Digest(msg) // 计算消息的摘要值
	file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
	log.SetPrefix("[Prepare error]")
	defer file.Close()
	// 判断是否符合校验条件
	if state.View != preprepare.View {
		log.Println("the view of preprepare message is wrong!")
		result = false
	} else if state.Last_sequence_number >= preprepare.Sequence_number {
		log.Println("the sequenceID of preprepare message is wrong!")
		log.Println("Last_sequence_number=", state.Last_sequence_number)
		log.Println("preprepare.Sequence_number=", preprepare.Sequence_number)
		result = false
	} else if !bytes.Equal(digest, preprepare.Digest_m) {
		log.Println("the digest is wrong!")
		result = false
	} else if !state.verifyRequest(preprepare.Request.Transactions) {
		log.Println("the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(preprepare.Sign_p) {
		log.Println("the primary_sign is wrong!")
		result = false
	} else {
		result = true
	}
	return result
}

//  State.Commit，所有联盟节点处理收到的prepare消息：各节点prepare——>其余节点commit
// 参数：准备消息*PrepareMsg
// 返回值：提交消息*CommitMsg，处理错误error
func (state *State) Commit(prepare *PrepareMsg) (*CommitMsg, error) {
	if !state.verifyPrepareMsg(prepare) { // 校验收到的prepare消息
		return nil, errors.New("prepare message is corrupted")
	} else if state.prepared() { // 检查是否收到2f+1（含本节点产生的prepare）个通过校验的prepare消息
		i, _ := strconv.ParseInt(qkdserv.Node_name[1:], 10, 64)
		_, ok := state.Msg_logs.CommittedMsgs[i] // 检查是否发送过commit消息
		if !ok {                                 // 如果log中无commit信息，则发送commit
			commit := &CommitMsg{ // 定义一个commit消息
				View:            prepare.View,            // 获取视图号
				Sequence_number: prepare.Sequence_number, // 获取索引号
				Digest_m:        prepare.Digest_m,        // 获取消息摘要
				Node_i:          i,                       // 获取节点编号
				Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
					Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
						Sign_dev_id:  utils.GetNodeIDTable(qkdserv.Node_name), // 签名者ID
						Sign_task_sn: uss.GenSignTaskSN(16),                   // 签名序列号
					},
					Main_row_num: qkdserv.QKDSignRandomMainRowNum{
						Sign_Node_Name: qkdserv.Node_name, // 签名者节点号
						Main_Row_Num:   0,                 // 签名主行号，签名时默认为0
					},
					Sign_counts: N - 1, // 验签者的数量
					Sign_len:    16,    // 签名的单位长度，一般默认为16
				},
			}
			commit.Sign_i.Message, _ = commit.signMessageEncode() // 获取commit阶段待签名消息
			// commit消息的签名
			commit.Sign_i = uss.Sign(commit.Sign_i.Sign_index,
				commit.Sign_i.Sign_counts, commit.Sign_i.Sign_len, commit.Sign_i.Message)

			state.Msg_logs.CommittedMsgs[i] = commit // 将commit写入log，以便后续投票校验
			state.Current_stage = Prepared           // 此时状态改变为Prepared
			return commit, nil
		}
	}
	return nil, nil
}

// State.VerifyPrepareMsg，验证收到的prepare消息的正确性
// 参数：准备消息*PrepareMsg
// 返回值：验证结果bool
func (state *State) verifyPrepareMsg(prepare *PrepareMsg) bool {
	var result bool
	msg, _ := json.Marshal(state.Msg_logs.ReqMsg)
	digest := utils.Digest(msg) // 计算消息的摘要值
	file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
	log.SetPrefix("[Commit error]")
	defer file.Close()
	if state.View != prepare.View {
		log.Println("the view of prepare message is wrong!")
		result = false
	} else if state.Last_sequence_number >= prepare.Sequence_number {
		log.Println("the sequenceID of prepare message is wrong!")
		log.Println("Last_sequence_number=", state.Last_sequence_number)
		log.Println("prepare.Sequence_number=", prepare.Sequence_number)
		result = false
	} else if !bytes.Equal(digest, prepare.Digest_m) {
		log.Println("the verify of digest is wrong!")
		result = false
	} else if !state.verifyRequest(state.Msg_logs.ReqMsg.Transactions) {
		log.Println("the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(prepare.Sign_i) {
		log.Println("the node_sign is wrong!")
		result = false
	} else {
		state.Msg_logs.PreparedMsgs[prepare.Node_i] = prepare
		result = true
	}
	return result
}

// State.prepared,验证收到的prepare消息是否符合要求
// 参数：共识状态state
// 返回值：验证结果bool
func (state *State) prepared() bool {
	file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
	log.SetPrefix("[prepared error]")
	defer file.Close()
	if state.Msg_logs.ReqMsg == nil {
		log.Println("request of state is nil")
		return false
	}

	if len(state.Msg_logs.PreparedMsgs) < 2*F {
		//log.Println("didn't receive 2*f prepared message,please wait")
		return false
	}
	return true
}

//  State.ReplyMsg，获取reply消息，当收到2f+1个满足要求的commit时，调用此函数
// 参数：提交消息*CommitMsg
// 返回值：应答消息*ReplyMsg，处理错误error
func (state *State) Reply(commit *CommitMsg) (*ReplyMsg, error) {
	if !state.verifyCommitMsg(commit) {
		return nil, errors.New("commit message is corrupted")
	} else if state.committed() {
		i, _ := strconv.ParseInt(qkdserv.Node_name[1:], 10, 64)
		_, ok := state.Msg_logs.ReplyMsgs[i] // 检查是否发送过reply消息
		if !ok {                             // 如果未发送过reply消息
			reply := &ReplyMsg{
				View:        commit.View,                      // 获取视图号
				Time_stamp:  state.Msg_logs.ReqMsg.Time_stamp, // 相应request消息的时间戳
				Client_name: "",                               // 客户端名称
				Node_i:      i,                                // 获取节点编号
				Result:      true,
				Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
					Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
						Sign_dev_id:  utils.GetNodeIDTable(qkdserv.Node_name), // 签名者ID
						Sign_task_sn: uss.GenSignTaskSN(16),                   // 签名序列号
					},
					Main_row_num: qkdserv.QKDSignRandomMainRowNum{
						Sign_Node_Name: qkdserv.Node_name, // 签名者节点号
						Main_Row_Num:   0,                 // 签名主行号，签名时默认为0
					},
					Sign_counts: 1,  // 验签者的数量，客户端验签
					Sign_len:    16, // 签名的单位长度，一般默认为16
				},
				Request: *state.Msg_logs.ReqMsg,
			}
			reply.Sign_i.Message, _ = reply.signMessageEncode()
			// reply消息的签名
			reply.Sign_i = uss.Sign(reply.Sign_i.Sign_index,
				reply.Sign_i.Sign_counts, reply.Sign_i.Sign_len, reply.Sign_i.Message)

			state.Msg_logs.ReplyMsgs[i] = reply
			state.CommittedMessage = commit
			state.Current_stage = Committed
			return reply, nil
		}
	}
	return nil, nil
}

// State.VerifyCommitMsg，验证收到的commit消息的正确性
// 参数：提交消息*CommitMsg
// 返回值：验证结果bool
func (state *State) verifyCommitMsg(commit *CommitMsg) bool {
	var result bool
	msg, _ := json.Marshal(state.Msg_logs.ReqMsg)
	digest := utils.Digest(msg) // 计算消息的摘要值
	file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
	log.SetPrefix("[Reply error]")
	defer file.Close()
	if state.View != commit.View {
		log.Println("the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 && state.Last_sequence_number >= commit.Sequence_number {
		log.Println("the sequenceID of commit message is wrong!")
		result = false
	} else if !bytes.Equal(digest, commit.Digest_m) {
		log.Println("the verify of digest is wrong!")
		result = false
	} else if !state.verifyRequest(state.Msg_logs.ReqMsg.Transactions) {
		log.Println("the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(commit.Sign_i) {
		log.Println("the node_sign is wrong!")
		result = false
	} else {
		state.Msg_logs.CommittedMsgs[commit.Node_i] = commit
		result = true
	}
	return result
}

// State.commited,验证收到的commit消息是否符合要求
// 参数：共识状态state
// 返回值：验证结果bool
func (state *State) committed() bool {
	file, _ := utils.Init_log(LOG_ERROR_PATH + qkdserv.Node_name + ".log")
	log.SetPrefix("[committed error]")
	defer file.Close()
	if !state.prepared() { // 如果prepare投票未通过，则不能进入commit
		log.Println("didn't prepared!")
		return false
	}
	if len(state.Msg_logs.CommittedMsgs) < 2*F+1 { // commit通过的条件是受到2f+1个校验通过的commit,包括自身节点
		//log.Println("didn't receive 2*f committed message,please wait!")
		return false
	}
	return true
}
