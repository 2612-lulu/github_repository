//pbft包，实现了pbft共识的相关处理
//创建人：zhanglu
//创建时间：2021/08/15
package pbft

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"qb/qbtools"
	"qb/qkdserv"
	"qb/uss"
	"strconv"
	"time"
)

type State struct {
	View                 int64 //视图号
	Msg_logs             MsgLogs
	Last_sequence_number int64
	CurrentStage         Stage
}

type MsgLogs struct {
	ReqMsg *RequestMsg
	//CommitMsgs map[int64]*CommitMsg
	//CommitMsgs  map[string]*VoteMsg
}

type Stage int

const (
	Idle        Stage = iota // Idle=0，节点已成功创建，但共识过程尚未启动。
	PrePrepared              // PrePrepared=1，RequestMsg已成功处理。节点已准备好进入prepare阶段。
	Prepared                 // Prepared=2，Same with `prepared` stage explained in the original paper.
	Committed                // Committed=3，Same with `committed-local` stage explained in the original paper.
)

// N=3F+1，本程序中N=4，即f=1
const F = 1 //f，容忍无效或者恶意节点数
const N = 4

// CreateState，如果不存在lastSequenceNumber，则lastSequenceNumber=-1
func CreateState(view int64, lastSequenceNumber int64) *State {
	return &State{
		View: view, // 当前视图号，为主节点编号
		Msg_logs: MsgLogs{ // 初始化
			ReqMsg: nil,
			//PrepareMsgs: make(map[string]*VoteMsg),
			//CommitMsgs:  make(map[string]*VoteMsg),
		},
		Last_sequence_number: lastSequenceNumber, // 上一个序列号
		CurrentStage:         Idle,               // 目前状态，节点创立，即将进入共识
	}
}

// PrePrePare，进入共识，客户端Request——>主节点PrePrePare——>从节点
func (state *State) PrePrePare(request *RequestMsg) (*PrePrepareMsg, bool) {
	var result bool

	preprepare := PrePrepareMsg{}           // 定义一个preprepare消息
	preprepare.View = state.View            // 获取视图号
	preprepare.Digest_m = Digest(request.M) // 获取请求消息的摘要
	// 1. 检查客户端签名是否正确
	if uss.VerifySign(request.Sign_client) {
		//fmt.Println("	The verify of ReqMsg is true, will enter the prepare statge!")
		sequenceID := time.Now().UnixNano() // 使用时间戳作为序列号
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
		m, _ := preprepare.SignMessageEncode()                                               // 获取preprepare阶段待签名消息
		for i := 0; i < len(m); i++ {
			preprepare.Sign_p.Message[i] = m[i]
		}
		// 获取Pre-prepare消息的签名
		preprepare.Sign_p = uss.Sign(preprepare.Sign_p.Sign_index,
			preprepare.Sign_p.Sign_counts, preprepare.Sign_p.Sign_len, preprepare.Sign_p.Message)
		// 将request消息添加到pre-prepare消息中
		preprepare.request = *request

		state.Msg_logs.ReqMsg = request  // 记录request消息到state的log中
		state.CurrentStage = PrePrepared // 此时状态改变为PrePrepared

		result = true // 客户端验签成功，进入prepare阶段

	} else {
		fmt.Println("	pbft-PrePrepare error:The verify of client sign is false!!!")
		result = false // 客户端验签失败
	}

	return &preprepare, result
}

//  PrePare，进入准备阶段，从节点PrePrepareMsg——>各节点PrepareMsg
func (state *State) PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, bool) {
	digest := Digest(preprepare.request.M) // 计算消息的摘要值

	var result bool // prepare结果

	prepare := PrepareMsg{} // 定义一个prepare消息

	// 判断是否符合校验条件！！！待完善
	if state.View != preprepare.View {
		fmt.Println("	pbft-Prepare error:the view is wrong!")
		result = false
	} else if !bytes.Equal(digest, preprepare.Digest_m) {
		fmt.Println("	pbft-Prepare error:the digest is wrong!")
		result = false
	} else if !uss.VerifySign(preprepare.request.Sign_client) {
		fmt.Println("	pbft-Prepare error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(preprepare.Sign_p) {
		fmt.Println("	pbft-Prepare error:the primary_sign is wrong!")
		result = false
	} else {
		prepare.View = preprepare.View                                             // 获取视图号
		prepare.Sequence_number = preprepare.Sequence_number                       // 获取索引号
		prepare.Digest_m = preprepare.request.Digest_m                             // 获取消息摘要
		prepare.Node_i, _ = strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64) // 获取节点编号
		// 确定preprepare消息的签名信息,签名者主行号信息可不定义，为0即可
		prepare.Sign_i.Sign_index.Sign_dev_id = qbtools.GetNodeIDTable(qkdserv.Node_name) // 签名者ID
		prepare.Sign_i.Sign_index.Sign_task_sn = uss.GenSignTaskSN(16)                    // 签名序列号
		prepare.Sign_i.Sign_counts = N - 1                                                // 验签者的数量
		prepare.Sign_i.Sign_len = 16                                                      // 签名的单位长度，一般默认为16
		prepare.Sign_i.Main_row_num.Sign_Node_Name = qkdserv.Node_name                    // 签名者节点号
		prepare.Sign_i.Main_row_num.Main_Row_Num = 0                                      // 签名主行号，签名时默认为0
		sign_m, _ := prepare.SignMessageEncode()                                          // 获取preprepare阶段待签名消息
		for i := 0; i < len(sign_m); i++ {
			prepare.Sign_i.Message[i] = sign_m[i]
		}
		// prepare消息的签名
		prepare.Sign_i = uss.Sign(prepare.Sign_i.Sign_index,
			prepare.Sign_i.Sign_counts, prepare.Sign_i.Sign_len, prepare.Sign_i.Message)

		state.CurrentStage = Prepared // 此时状态改变为Prepared
		//fmt.Println("	accept Pre-Prepare message，prepared,will enter the commit statge")
		result = true // 接受PrePrepare消息，可以进入prepare转发消息给其他节点
	}
	return &prepare, result
}

//  GetCommitMsg，提交阶段用于获取commit消息，当收到2f个验证通过的prepare消息时调用此函数
func (state *State) GetCommitMsg(prepare *PrepareMsg) *CommitMsg {
	commit := CommitMsg{}                                                     // 定义一个commit消息
	commit.View = prepare.View                                                // 获取视图号
	commit.Sequence_number = prepare.Sequence_number                          // 获取索引号
	commit.Digest_m = prepare.Digest_m                                        // 获取消息摘要
	commit.Node_i, _ = strconv.ParseInt(string(qkdserv.Node_name[1]), 10, 64) // 获取节点编号
	// 确定preprepare消息的签名信息,签名者主行号信息可不定义，为0即可
	commit.Sign_i.Sign_index.Sign_dev_id = qbtools.GetNodeIDTable(qkdserv.Node_name) // 签名者ID
	commit.Sign_i.Sign_index.Sign_task_sn = uss.GenSignTaskSN(16)                    // 签名序列号
	commit.Sign_i.Sign_counts = N - 1                                                // 验签者的数量
	commit.Sign_i.Sign_len = 16                                                      // 签名的单位长度，一般默认为16
	commit.Sign_i.Main_row_num.Sign_Node_Name = qkdserv.Node_name                    // 签名者节点号
	commit.Sign_i.Main_row_num.Main_Row_Num = 0                                      // 签名主行号，签名时默认为0
	sign_m, _ := commit.SignMessageEncode()                                          // 获取preprepare阶段待签名消息
	for i := 0; i < len(sign_m); i++ {
		commit.Sign_i.Message[i] = sign_m[i]
	}
	// commit消息的签名
	commit.Sign_i = uss.Sign(commit.Sign_i.Sign_index,
		commit.Sign_i.Sign_counts, commit.Sign_i.Sign_len, commit.Sign_i.Message)

	state.CurrentStage = Committed // 此时状态改变为Committed
	return &commit
}

//  GetReplyMsg，获取reply消息，当收到2f+1个满足要求的commit时，调用此函数
func (state *State) GetReplyMsg(commit *CommitMsg) *ReplyMsg {
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
	m, _ := reply.SignMessageEncode()
	for i := 0; i < len(m); i++ {
		reply.Sign_i.Message[i] = m[i]
	}
	// reply消息的签名
	reply.Sign_i = uss.Sign(reply.Sign_i.Sign_index,
		reply.Sign_i.Sign_counts, reply.Sign_i.Sign_len, reply.Sign_i.Message)

	return &reply
}

func (state *State) VerifyPrepareMsg(prepare *PrepareMsg) bool {
	var result bool
	digest := Digest(state.Msg_logs.ReqMsg.M) // 计算消息的摘要值

	if state.View != prepare.View {
		fmt.Println("	pbft-Prepare error:the view is wrong!")
		result = false
	} else if !bytes.Equal(digest, prepare.Digest_m) {
		fmt.Println("	pbft-Prepare error:the digest is wrong!")
		result = false
	} else if !uss.VerifySign(state.Msg_logs.ReqMsg.Sign_client) {
		fmt.Println("	pbft-Prepare error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(prepare.Sign_i) {
		fmt.Println("	pbft-Prepare error:the primary_sign is wrong!")
		result = false
	} else {
		result = true

	}
	return result
}

func (state *State) VerifyCommitMsg(commit *CommitMsg) bool {
	var result bool
	digest := Digest(state.Msg_logs.ReqMsg.M) // 计算消息的摘要值

	if state.View != commit.View { // 校验视图号
		fmt.Println("	pbft-Commit error:the view is wrong!")
		result = false
	} else if !bytes.Equal(digest, commit.Digest_m) { // 校验消息摘要值
		fmt.Println("	pbft-Commit error:the digest is wrong!")
		result = false
	} else if !uss.VerifySign(state.Msg_logs.ReqMsg.Sign_client) { // 校验客户端签名
		fmt.Println("	pbft-Commit error:the client_sign is wrong!")
		result = false
	} else if !uss.VerifySign(commit.Sign_i) { // 校验上一阶段节点签名
		fmt.Println("	pbft-Commit error:the primary_sign is wrong!")
		result = false
	} else {
		result = true
	}
	return result
}

func (obj *RequestMsg) SignMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.Time_stamp)
	binary.Write(buf, binary.LittleEndian, obj.Client_id)
	binary.Write(buf, binary.LittleEndian, obj.Operation_type)
	binary.Write(buf, binary.LittleEndian, obj.M)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	return buf.Bytes(), nil
}

func (obj *PrePrepareMsg) SignMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	return buf.Bytes(), nil
}

func (obj *PrepareMsg) SignMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

func (obj *CommitMsg) SignMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

func (obj *ReplyMsg) SignMessageEncode() ([]byte, error) {
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
