//pbft包，实现了pbft共识的相关处理
//创建人：zhanglu
//创建时间：2021/08/15
package pbft

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"qb/block"
	"qb/qbtools"
	"qb/qkdserv"
	"qb/uss"
)

type State struct {
	View                 int64 // 视图号
	Msg_logs             MsgLogs
	Last_sequence_number int64 // 上次共识序列号
	Current_stage        Stage
	CommittedMessage     *CommitMsg // 达成共识的commit消息
}

type MsgLogs struct {
	ReqMsg        *block.Block          // 存放request消息
	PreparedMsgs  map[int64]*PrepareMsg // 存放prepared消息
	CommittedMsgs map[int64]*CommitMsg  // 存放committed消息
	ReplyMsgs     map[int64]*ReplyMsg   // 存放Reply消息
}

// N=3F+1，本程序中N=4，即F=1
const F = 1 //F，容忍无效或者恶意节点数
const N = 4

type Stage int

const (
	Idle        Stage = iota // Idle=0，节点已成功创建，但共识过程尚未启动。
	PrePrepared              // PrePrepared=1，RequestMsg已成功处理。节点已准备好进入prepare阶段。
	Prepared                 // Prepared=2，Same with `prepared` stage explained in the original paper.
	Committed                // Committed=3，Same with `committed-local` stage explained in the original paper.
)

// CreateState，如果不存在lastSequenceNumber，则lastSequenceNumber=-1
func CreateState(view int64, lastSequenceNumber int64) *State {
	return &State{
		View: view, // 当前视图号，为主节点编号
		Msg_logs: MsgLogs{ // 初始化
			ReqMsg:        nil,
			PreparedMsgs:  make(map[int64]*PrepareMsg),
			CommittedMsgs: make(map[int64]*CommitMsg),
			ReplyMsgs:     make(map[int64]*ReplyMsg),
		},
		CommittedMessage:     nil,
		Last_sequence_number: lastSequenceNumber, // 上一个序列号
		Current_stage:        Idle,               // 目前状态，节点创立，即将进入共识
	}
}

// PrePrePare，进入共识，由主节点进行消息处理：客户端Request——>主节点PrePrePare——>从节点
func (state *State) PrePrePare(request *block.Block) (*PrePrepareMsg, error) {
	state.Msg_logs.ReqMsg = request // 记录request消息到state的log中

	if state.verifyRequest(request.Transactions) { // 如果每条交易信息验签成功
		sequenceID := time.Now().UnixNano() // 使用时间戳作为暂时序列号
		if state.Last_sequence_number != -1 {
			sequenceID = state.Last_sequence_number + 1 // 主节点每开始一次共识，序列号+1
		}

		digest_msg, _ := json.Marshal(request)
		// 定义一个preprepare消息
		preprepare := &PrePrepareMsg{
			View:            state.View,                 // 获取视图号
			Sequence_number: sequenceID,                 // 为其分配序列号
			Digest_m:        qbtools.Digest(digest_msg), // 交易信息摘要
			Sign_p: uss.USSToeplitzHashSignMsg{ // 签名信息
				Sign_index: qkdserv.QKDSignMatrixIndex{
					Sign_dev_id:  qbtools.GetNodeIDTable(qkdserv.Node_name),
					Sign_task_sn: uss.GenSignTaskSN(16),
				},
				Main_row_num: qkdserv.QKDSignRandomMainRowNum{
					Sign_Node_Name: qkdserv.Node_name,
					Main_Row_Num:   0, // 签名主行号，签名时默认为0
				},
				Sign_counts: N - 1,
				Sign_len:    16,
			},
			Request: *request, // 将请求消息附在preprepare中广播给所有从节点
		}
		preprepare.Sign_p.Message = preprepare.signMessageEncode()
		preprepare.Sign_p = uss.Sign(preprepare.Sign_p.Sign_index, preprepare.Sign_p.Sign_counts,
			preprepare.Sign_p.Sign_len, preprepare.Sign_p.Message)
		state.Current_stage = PrePrepared
		return preprepare, nil
	} else {
		fmt.Println("failing to verify request message")
		return nil, nil
	}
}

// VerifyRequest
func (state *State) verifyRequest(request []*block.Transaction) bool {
	//TODO:验证每条交易信息的签名
	verify_num := 0
	for _, reqMsg := range request {
		if uss.VerifySign(reqMsg.Sign_client) {
			verify_num++
		}
	}
	if verify_num == len(request) {
		return true
	} else {
		fmt.Println("	pbft-PrePrepare error:The verify of client sign is false!!!")
		return false
	}
}

func (obj *PrePrepareMsg) signMessageEncode() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	return buf.Bytes()
}

//  PrePare，进入准备阶段，从节点处理pre-prepare消息：从节点PrePrepareMsg——>各节点PrepareMsg
func (state *State) PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, error) {
	state.Msg_logs.ReqMsg = &preprepare.Request // 将request消息提取出来记录到state中
	if !state.verifyPrePrepareMsg(preprepare) { // 校验受到的pre-prepare是否通过
		return nil, errors.New("pre-prepare message is corrupted")
	} else {
		i, _ := strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64) // 获取节点编号
		prepare := &PrepareMsg{                                        // 定义一个prepare消息
			View:            preprepare.View,            // 获取视图号
			Sequence_number: preprepare.Sequence_number, // 获取索引号
			Digest_m:        preprepare.Digest_m,        // 获取消息摘要
			Node_i:          i,
			Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
				Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
					Sign_dev_id:  qbtools.GetNodeIDTable(qkdserv.Node_name), // 签名者ID
					Sign_task_sn: uss.GenSignTaskSN(16),                     // 签名序列号
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

// VerifyPrePrepareMsg,校验受到的pre-prepare消息
func (state *State) verifyPrePrepareMsg(preprepare *PrePrepareMsg) bool {
	var result bool
	msg, _ := json.Marshal(preprepare.Request)
	digest := qbtools.Digest(msg) // 计算消息的摘要值
	// 判断是否符合校验条件
	if state.View != preprepare.View {
		fmt.Println("	pbft-Prepare error:the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 && state.Last_sequence_number >= preprepare.Sequence_number {
		fmt.Println("	pbft-Prepare error:the sequenceID is wrong!")
		result = false
	} else if !bytes.Equal(digest, preprepare.Digest_m) {
		fmt.Println("	pbft-Prepare error:the digest is wrong!")
		result = false
	} else if !state.verifyRequest(preprepare.Request.Transactions) {
		fmt.Println("	pbft-Prepare error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(preprepare.Sign_p) {
		fmt.Println("	pbft-Prepare error:the primary_sign is wrong!")
		result = false
	} else {
		result = true
	}
	return result
}

func (obj *PrepareMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

//  Commit，所有联盟节点处理收到的prepare消息：各节点prepare——>其余节点commit
func (state *State) Commit(prepare *PrepareMsg) (*CommitMsg, error) {
	if !state.verifyPrepareMsg(prepare) { // 校验收到的prepare消息
		return nil, errors.New("prepare message is corrupted")
	} else if state.prepared() { // 检查是否收到2f+1（含本节点产生的prepare）个通过校验的prepare消息
		i, _ := strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64)
		_, ok := state.Msg_logs.CommittedMsgs[i] // 检查是否发送过commit消息
		if !ok {                                 // 如果log中无commit信息，则发送commit
			commit := &CommitMsg{ // 定义一个commit消息
				View:            prepare.View,            // 获取视图号
				Sequence_number: prepare.Sequence_number, // 获取索引号
				Digest_m:        prepare.Digest_m,        // 获取消息摘要
				Node_i:          i,                       // 获取节点编号
				Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
					Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
						Sign_dev_id:  qbtools.GetNodeIDTable(qkdserv.Node_name), // 签名者ID
						Sign_task_sn: uss.GenSignTaskSN(16),                     // 签名序列号
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
		} else {
			return nil, nil
		}
	}
	return nil, nil
}

// VerifyPrepareMsg，校验受到的prepare消息
func (state *State) verifyPrepareMsg(prepare *PrepareMsg) bool {
	var result bool
	msg, _ := json.Marshal(state.Msg_logs.ReqMsg)
	digest := qbtools.Digest(msg) // 计算消息的摘要值
	if state.View != prepare.View {
		fmt.Println("	pbft-Commit error:the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 && state.Last_sequence_number >= prepare.Sequence_number {
		fmt.Println("	pbft-Commit error:the sequenceID is wrong!")
		result = false
	} else if !bytes.Equal(digest, prepare.Digest_m) {
		fmt.Println("	pbft-Commit error:the digest is wrong!")
		result = false
	} else if !state.verifyRequest(state.Msg_logs.ReqMsg.Transactions) {
		fmt.Println("	pbft-Commit error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(prepare.Sign_i) {
		fmt.Println("	pbft-Commit error:the node_sign is wrong!")
		result = false
	} else {
		state.Msg_logs.PreparedMsgs[prepare.Node_i] = prepare
		result = true
	}
	return result
}

func (state *State) prepared() bool {
	if state.Msg_logs.ReqMsg == nil {
		fmt.Println("	pbft-commit error:request of state is nil")
		return false
	}

	if len(state.Msg_logs.PreparedMsgs) < 2*F {
		fmt.Println("	pbft-commit error:didn't receive 2*f prepared message,please wait")
		return false
	}

	return true
}

func (obj *CommitMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

//  ReplyMsg，获取reply消息，当收到2f+1个满足要求的commit时，调用此函数
func (state *State) Reply(commit *CommitMsg) (*ReplyMsg, error) {
	if !state.verifyCommitMsg(commit) {
		return nil, errors.New("commit message is corrupted")
	} else if state.committed() {
		i, _ := strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64)
		_, ok := state.Msg_logs.ReplyMsgs[i] // 检查是否发送过reply消息
		if !ok {                             // 如果未发送过reply消息
			reply := &ReplyMsg{
				View:       commit.View,                      // 获取视图号
				Time_stamp: state.Msg_logs.ReqMsg.Time_stamp, // 相应request消息的时间戳
				// 客户端名称在区块里
				Node_i: i, // 获取节点编号
				Result: true,
				Sign_i: uss.USSToeplitzHashSignMsg{ // 签名信息
					Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
						Sign_dev_id:  qbtools.GetNodeIDTable(qkdserv.Node_name), // 签名者ID
						Sign_task_sn: uss.GenSignTaskSN(16),                     // 签名序列号
					},
					Main_row_num: qkdserv.QKDSignRandomMainRowNum{
						Sign_Node_Name: qkdserv.Node_name, // 签名者节点号
						Main_Row_Num:   0,                 // 签名主行号，签名时默认为0
					},
					Sign_counts: 1,  // 验签者的数量，客户端验签
					Sign_len:    16, // 签名的单位长度，一般默认为16
				},
			}
			reply.Sign_i.Message, _ = reply.signMessageEncode()
			// reply消息的签名
			reply.Sign_i = uss.Sign(reply.Sign_i.Sign_index,
				reply.Sign_i.Sign_counts, reply.Sign_i.Sign_len, reply.Sign_i.Message)

			state.Msg_logs.ReplyMsgs[i] = reply
			state.CommittedMessage = commit
			state.Current_stage = Committed
			return reply, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}
}

func (state *State) verifyCommitMsg(commit *CommitMsg) bool {
	var result bool
	msg, _ := json.Marshal(state.Msg_logs.ReqMsg)
	digest := qbtools.Digest(msg) // 计算消息的摘要值
	if state.View != commit.View {
		fmt.Println("	pbft-Reply error:the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 && state.Last_sequence_number >= commit.Sequence_number {
		fmt.Println("	pbft-Reply error:the sequenceID is wrong!")
		result = false
	} else if !bytes.Equal(digest, commit.Digest_m) {
		fmt.Println("	pbft-Reply error:the digest is wrong!")
		result = false
	} else if !state.verifyRequest(state.Msg_logs.ReqMsg.Transactions) {
		fmt.Println("	pbft-Reply error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(commit.Sign_i) {
		fmt.Println("	pbft-Reply error:the node_sign is wrong!")
		result = false
	} else {
		state.Msg_logs.CommittedMsgs[commit.Node_i] = commit
		result = true
	}
	return result
}

func (state *State) committed() bool {
	if !state.prepared() { // 如果prepare投票未通过，则不能进入commit
		return false
	}
	if len(state.Msg_logs.CommittedMsgs) < 2*F+1 { // commit通过的条件是受到2f+1个校验通过的commit,包括自身节点
		return false
	}
	return true
}

func (obj *ReplyMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Time_stamp)
	binary.Write(buf, binary.LittleEndian, obj.Client_name)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	binary.Write(buf, binary.LittleEndian, obj.Result)
	return buf.Bytes(), nil
}
