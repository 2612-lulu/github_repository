// pbft包，实现了pbft共识的相关处理
// 创建人：zhanglu
// 创建时间：2021/08/15
package pbft

import (
	"qblock"
)

// pbft状态标识
type State struct {
	View                 View     // 视图号
	Msg_logs             *MsgLogs // 缓存数据
	Last_sequence_number int64    // 上次共识序列号
	Current_stage        Stage    // 当前状态
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
	ID      int64  `json:"viewid"`  // 视图号
	Primary string `json:"primary"` // 主节点
	F       int64  `json:"f"`       // 可容忍错误节点的数量
}

// 状态标识
type Stage int

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
		View: View{
			ID: view, // 当前视图号，为主节点编号
		},
		Msg_logs: &MsgLogs{ // 初始化
			ReqMsg:        nil,
			PreparedMsgs:  make(map[int64]*PrepareMsg),
			CommittedMsgs: make(map[int64]*CommitMsg),
			ReplyMsgs:     make(map[int64]*ReplyMsg),
		},
		Last_sequence_number: lastSequenceNumber, // 上一个序列号
		Current_stage:        Idle,               // 目前状态，节点创立，即将进入共识
	}
}
