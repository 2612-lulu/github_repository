//pbft包，实现了pbft共识的相关处理
//创建人：zhanglu
//创建时间：2021/08/15
package pbft

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"qb/qbtools"
	"qb/qkdserv"
	"qb/uss"
	"strconv"
	"time"
)

type State struct {
	View                 int64 // 视图号
	Msg_logs             MsgLogs
	Last_sequence_number int64 // 上次共识序列号
	CurrentStage         Stage
}

type MsgLogs struct {
	ReqMsg        *RequestMsg          // 存放request消息
	PreparedMsgs  map[int64]PrepareMsg // 存放prepared消息
	CommittedMsgs map[int64]CommitMsg  // 存放committed消息
}

type Stage int

const (
	Idle        Stage = iota // Idle=0，节点已成功创建，但共识过程尚未启动。
	PrePrepared              // PrePrepared=1，RequestMsg已成功处理。节点已准备好进入prepare阶段。
	Prepared                 // Prepared=2，Same with `prepared` stage explained in the original paper.
	Committed                // Committed=3，Same with `committed-local` stage explained in the original paper.
)

// N=3F+1，本程序中N=4，即F=1
const F = 1 //f，容忍无效或者恶意节点数
const N = 4

// CreateState，如果不存在lastSequenceNumber，则lastSequenceNumber=-1
func CreateState(view int64, lastSequenceNumber int64) *State {
	return &State{
		View: view, // 当前视图号，为主节点编号
		Msg_logs: MsgLogs{ // 初始化
			ReqMsg:        nil,
			PreparedMsgs:  make(map[int64]PrepareMsg),
			CommittedMsgs: make(map[int64]CommitMsg),
		},
		Last_sequence_number: lastSequenceNumber, // 上一个序列号
		CurrentStage:         Idle,               // 目前状态，节点创立，即将进入共识
	}
}

// 生成请求消息，只有客户端可以调用该函数
func (state *State) GenReqMsg(operation string, node_name [2]byte) (*RequestMsg, bool) {
	request := RequestMsg{}
	request.Time_stamp = time.Now().UnixNano()            // 获取当前时间戳
	request.Client_id = qbtools.GetNodeIDTable(node_name) // 获取签名者ID
	request.Operation_type = []byte("transaction")        // 交易类型目前默认只有交易一种类型
	request.M = []byte(operation)                         // 具体操作请求
	request.Digest_m = Digest(request.M)                  // 操作请求消息摘要

	// 确定preprepare消息的签名信息,签名者主行号信息可不定义，为0即可
	request.Sign_client.Sign_index.Sign_dev_id = request.Client_id      // 签名者ID
	request.Sign_client.Sign_index.Sign_task_sn = uss.GenSignTaskSN(16) // 签名序列号
	request.Sign_client.Sign_counts = N                                 // 验签者的数量，等于节点数量
	request.Sign_client.Sign_len = 16                                   // 签名的单位长度，一般默认为16
	request.Sign_client.Main_row_num.Sign_Node_Name = node_name         // 签名者节点号
	request.Sign_client.Main_row_num.Main_Row_Num = 0                   // 签名主行号，签名时默认为0
	request.Sign_client.Message, _ = request.signMessageEncode()        // 获取preprepare阶段待签名消息
	// 获取Pre-prepare消息的签名
	request.Sign_client = uss.Sign(request.Sign_client.Sign_index,
		request.Sign_client.Sign_counts, request.Sign_client.Sign_len, request.Sign_client.Message)
	state.Msg_logs.ReqMsg = &request // 记录request消息到state的log中

	return &request, true

}

// PrePrePare，进入共识，由主节点进行消息处理：客户端Request——>主节点PrePrePare——>从节点
func (state *State) PrePrePare(request *RequestMsg) (*PrePrepareMsg, bool) {
	var result bool
	state.Msg_logs.ReqMsg = request // 记录request消息到state的log中

	preprepare := PrePrepareMsg{}          // 定义一个preprepare消息
	preprepare.View = state.View           // 获取视图号
	preprepare.Digest_m = request.Digest_m // 获取请求消息的摘要
	// 1. 检查客户端签名是否正确
	if uss.VerifySign(request.Sign_client) {
		sequenceID := time.Now().UnixNano() // 使用时间戳作为暂时序列号
		if state.Last_sequence_number != -1 {
			for state.Last_sequence_number >= sequenceID {
				sequenceID = state.Last_sequence_number + 1 // 主节点每开始一次共识，序列号+1
			}
		}
		preprepare.Sequence_number = sequenceID // 为其分配序列号

		// 确定preprepare消息的签名信息,签名者主行号信息可不定义，为0即可
		preprepare.Sign_p.Sign_index.Sign_dev_id = qbtools.GetNodeIDTable(qkdserv.Node_name) // 签名者ID
		preprepare.Sign_p.Sign_index.Sign_task_sn = uss.GenSignTaskSN(16)                    // 签名序列号
		preprepare.Sign_p.Sign_counts = N - 1                                                // 验签者的数量
		preprepare.Sign_p.Sign_len = 16                                                      // 签名的单位长度，一般默认为16
		preprepare.Sign_p.Main_row_num.Sign_Node_Name = qkdserv.Node_name                    // 签名者节点号
		preprepare.Sign_p.Main_row_num.Main_Row_Num = 0                                      // 签名主行号，签名时默认为0
		preprepare.Sign_p.Message, _ = preprepare.signMessageEncode()                        // 获取preprepare阶段待签名消息
		// 获取Pre-prepare消息的签名
		preprepare.Sign_p = uss.Sign(preprepare.Sign_p.Sign_index,
			preprepare.Sign_p.Sign_counts, preprepare.Sign_p.Sign_len, preprepare.Sign_p.Message)
		preprepare.Request = *request // 将请求消息附在preprepare中广播给所有从节点

		state.CurrentStage = PrePrepared // 此时状态改变为PrePrepared
		result = true                    // 客户端验签成功，即将进入prepare阶段
		//fmt.Println("	The verify of ReqMsg is true, get the preprepare message!")

	} else {
		fmt.Println("	pbft-PrePrepare error:The verify of client sign is false!!!")
		result = false // 客户端验签失败
	}

	return &preprepare, result
}

//  PrePare，进入准备阶段，从节点处理pre-prepare消息：从节点PrePrepareMsg——>各节点PrepareMsg
func (state *State) PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, error) {
	state.Msg_logs.ReqMsg = &preprepare.Request // 将request消息提取出来记录到state中
	if !state.VerifyPrePrepareMsg(preprepare) { // 校验受到的pre-prepare是否通过
		return nil, errors.New("pre-prepare message is corrupted")
	}
	prepare := PrepareMsg{}                                                    // 定义一个prepare消息
	prepare.View = preprepare.View                                             // 获取视图号
	prepare.Sequence_number = preprepare.Sequence_number                       // 获取索引号
	prepare.Digest_m = preprepare.Digest_m                                     // 获取消息摘要
	prepare.Node_i, _ = strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64) // 获取节点编号
	// 确定preprepare消息的签名信息,签名者主行号信息可不定义，为0即可
	prepare.Sign_i.Sign_index.Sign_dev_id = qbtools.GetNodeIDTable(qkdserv.Node_name) // 签名者ID
	prepare.Sign_i.Sign_index.Sign_task_sn = uss.GenSignTaskSN(16)                    // 签名序列号
	prepare.Sign_i.Sign_counts = N - 1                                                // 验签者的数量
	prepare.Sign_i.Sign_len = 16                                                      // 签名的单位长度，一般默认为16
	prepare.Sign_i.Main_row_num.Sign_Node_Name = qkdserv.Node_name                    // 签名者节点号
	prepare.Sign_i.Main_row_num.Main_Row_Num = 0                                      // 签名主行号，签名时默认为0
	prepare.Sign_i.Message, _ = prepare.signMessageEncode()                           // 获取prepare阶段待签名消息
	// prepare消息的签名
	prepare.Sign_i = uss.Sign(prepare.Sign_i.Sign_index,
		prepare.Sign_i.Sign_counts, prepare.Sign_i.Sign_len, prepare.Sign_i.Message)
	msg := prepare
	state.Msg_logs.PreparedMsgs[prepare.Node_i] = msg // 将节点自己产生的prepare消息写入log，以便后续进行投票校验
	//fmt.Printf("	%d log prepare\n", prepare.Node_i)
	state.CurrentStage = PrePrepared // 此时状态改变为PrePrepared
	//fmt.Println("	accept Pre-Prepare message，prepared,will enter the commit statge")
	return &prepare, nil
}

//  Commit，所有联盟节点处理收到的prepare消息：各节点prepare——>其余节点commit
func (state *State) Commit(prepare *PrepareMsg) (*CommitMsg, error) {
	if !state.VerifyPrepareMsg(prepare) { // 校验收到的prepare消息
		return nil, errors.New("prepare message is corrupted")
	}

	commit := CommitMsg{} // 定义一个commit消息
	if state.prepared() { // 检查是否受到2f+1（含本节点产生的prepare）个通过校验的prepare消息
		commit.View = prepare.View                                                // 获取视图号
		commit.Sequence_number = prepare.Sequence_number                          // 获取索引号
		commit.Digest_m = prepare.Digest_m                                        // 获取消息摘要
		commit.Node_i, _ = strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64) // 获取节点编号

		// 检查是否发送过commit消息
		_, ok := state.Msg_logs.CommittedMsgs[commit.Node_i]
		if !ok { // 如果log中无commit信息，则发送commit
			// 确定preprepare消息的签名信息,签名者主行号信息可不定义，为0即可
			commit.Sign_i.Sign_index.Sign_dev_id = qbtools.GetNodeIDTable(qkdserv.Node_name) // 签名者ID
			commit.Sign_i.Sign_index.Sign_task_sn = uss.GenSignTaskSN(16)                    // 签名序列号
			commit.Sign_i.Sign_counts = N - 1                                                // 验签者的数量
			commit.Sign_i.Sign_len = 16                                                      // 签名的单位长度，一般默认为16
			commit.Sign_i.Main_row_num.Sign_Node_Name = qkdserv.Node_name                    // 签名者节点号
			commit.Sign_i.Main_row_num.Main_Row_Num = 0                                      // 签名主行号，签名时默认为0
			commit.Sign_i.Message, _ = commit.signMessageEncode()                            // 获取commit阶段待签名消息
			// commit消息的签名
			commit.Sign_i = uss.Sign(commit.Sign_i.Sign_index,
				commit.Sign_i.Sign_counts, commit.Sign_i.Sign_len, commit.Sign_i.Message)
			msg := commit
			state.Msg_logs.CommittedMsgs[commit.Node_i] = msg // 将commit写入log，以便后续投票校验

			fmt.Printf("[Prepare-Vote]: %d\n", len(state.Msg_logs.PreparedMsgs))

			state.CurrentStage = Prepared // 此时状态改变为Prepared
			return &commit, nil
		} else if ok { // 如果log中已有commit，表明已发送过commit，无需重复发送
			fmt.Println("	sent commit already")
			return nil, nil
		}
	}
	return nil, nil
}

//  GetReplyMsg，获取reply消息，当收到2f+1个满足要求的commit时，调用此函数
func (state *State) GetReply(commit *CommitMsg) (*ReplyMsg, error) {
	if !state.VerifyCommitMsg(commit) {
		return nil, errors.New("commit message is corrupted")
	}
	if state.committed() {
		reply := ReplyMsg{}
		reply.View = commit.View
		reply.Client_id = state.Msg_logs.ReqMsg.Client_id
		reply.Time_stamp = state.Msg_logs.ReqMsg.Time_stamp
		reply.Node_i, _ = strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64)
		reply.Result = true
		// 确定preprepare消息的签名信息,签名者主行号信息可不定义，为0即可
		reply.Sign_i.Sign_index.Sign_dev_id = qbtools.GetNodeIDTable(qkdserv.Node_name) // 签名者ID
		reply.Sign_i.Sign_index.Sign_task_sn = uss.GenSignTaskSN(16)                    // 签名序列号
		reply.Sign_i.Sign_counts = 1                                                    // 验签者的数量
		reply.Sign_i.Sign_len = 16                                                      // 签名的单位长度，一般默认为16
		reply.Sign_i.Main_row_num.Sign_Node_Name = qkdserv.Node_name                    // 签名者节点号
		reply.Sign_i.Main_row_num.Main_Row_Num = 0                                      // 签名主行号，签名时默认为0
		reply.Sign_i.Message, _ = reply.signMessageEncode()
		// reply消息的签名
		reply.Sign_i = uss.Sign(reply.Sign_i.Sign_index,
			reply.Sign_i.Sign_counts, reply.Sign_i.Sign_len, reply.Sign_i.Message)

		state.CurrentStage = Committed
		return &reply, nil
	}
	return nil, nil
}

func (state *State) VerifyPrePrepareMsg(preprepare *PrePrepareMsg) bool {
	var result bool
	digest := Digest(state.Msg_logs.ReqMsg.M) // 计算消息的摘要值
	// 判断是否符合校验条件
	if state.View != preprepare.View {
		fmt.Println("	pbft-Prepare error:the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 {
		if state.Last_sequence_number >= preprepare.Sequence_number {
			fmt.Println("	pbft-Prepare error:the sequenceID is wrong!")
			result = false
		}
	} else if !bytes.Equal(digest, preprepare.Digest_m) {
		fmt.Println("	pbft-Prepare error:the digest is wrong!")
		result = false
	} else if !uss.VerifySign(state.Msg_logs.ReqMsg.Sign_client) {
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
func (state *State) VerifyPrepareMsg(prepare *PrepareMsg) bool {
	var result bool
	digest := Digest(state.Msg_logs.ReqMsg.M) // 计算消息的摘要值

	if state.View != prepare.View {
		fmt.Println("	pbft-Commit error:the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 {
		if state.Last_sequence_number >= prepare.Sequence_number {
			fmt.Println("	pbft-Commit error:the sequenceID is wrong!")
			result = false
		}
	} else if !bytes.Equal(digest, prepare.Digest_m) {
		fmt.Println("	pbft-Commit error:the digest is wrong!")
		result = false
	} else if !uss.VerifySign(state.Msg_logs.ReqMsg.Sign_client) {
		fmt.Println("	pbft-Commit error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(prepare.Sign_i) {
		fmt.Println("	pbft-Commit error:the primary_sign is wrong!")
		result = false
	} else {
		state.Msg_logs.PreparedMsgs[prepare.Node_i] = *prepare
		fmt.Printf("	reveive prepare from %d\n", prepare.Node_i)
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

func (state *State) VerifyCommitMsg(commit *CommitMsg) bool {
	var result bool
	digest := Digest(state.Msg_logs.ReqMsg.M) // 计算消息的摘要值

	if state.View != commit.View {
		fmt.Println("	pbft-Reply error:the view is wrong!")
		result = false
	} else if state.Last_sequence_number != -1 {
		if state.Last_sequence_number >= commit.Sequence_number {
			fmt.Println("	pbft-Reply error:the sequenceID is wrong!")
			result = false
		}
	} else if !bytes.Equal(digest, commit.Digest_m) {
		fmt.Println("	pbft-Reply error:the digest is wrong!")
		result = false
	} else if !uss.VerifySign(state.Msg_logs.ReqMsg.Sign_client) {
		fmt.Println("	pbft-Reply error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(commit.Sign_i) {
		fmt.Println("	pbft-Reply error:the primary_sign is wrong!")
		result = false
	} else {
		state.Msg_logs.CommittedMsgs[commit.Node_i] = *commit
		result = true
	}
	return result
}

func (state *State) committed() bool {
	if !state.prepared() { // 如果prepare投票未通过，则不能进入commit
		return false
	}
	if len(state.Msg_logs.CommittedMsgs) < 2*F+1 {
		return false
	}
	return true
}

func (obj *RequestMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.Time_stamp)
	binary.Write(buf, binary.LittleEndian, obj.Client_id)
	binary.Write(buf, binary.LittleEndian, obj.Operation_type)
	binary.Write(buf, binary.LittleEndian, obj.M)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	return buf.Bytes(), nil
}

func (obj *PrePrepareMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	return buf.Bytes(), nil
}

func (obj *PrepareMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

func (obj *CommitMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

func (obj *ReplyMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Time_stamp)
	binary.Write(buf, binary.LittleEndian, obj.Client_id)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	binary.Write(buf, binary.LittleEndian, obj.Result)
	return buf.Bytes(), nil
}

// digest，摘要函数
func Digest(m []byte) []byte {
	h := sha256.New()
	h.Write(m)
	digest_m := h.Sum(nil)
	return digest_m
}
