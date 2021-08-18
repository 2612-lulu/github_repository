package network

import (
	"qb/pbft"
	"time"
)

type Node struct {
	Node_name    [2]byte           // 联盟节点名称
	NodeID       [16]byte          // 联盟节点ID，16字节QKD设备号
	NodeTable    map[string]string // key=nodeID, value=url
	View         *View             // 视图号
	CurrentState *pbft.State       // pbft阶段状态
	//CommittedMsgs []*pbft.RequestMsg
	MsgBuffer   *MsgBuffer
	MsgEntrance chan interface{}
	MsgDelivery chan interface{}
	Alarm       chan bool
}

type MsgBuffer struct {
	ReqMsgs        []*pbft.RequestMsg
	ReplyMsgs      []*pbft.ReplyMsg
	PrePrepareMsgs []*pbft.PrePrepareMsg
	PrepareMsgs    []*pbft.PrepareMsg
	CommitMsgs     []*pbft.CommitMsg
}

type View struct {
	ID      int64   // 视图号
	Primary [2]byte // 主节点
}

const ResolvingTimeDuration = time.Millisecond * 1000 // 1 second.
