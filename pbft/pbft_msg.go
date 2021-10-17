package pbft

import (
	"encoding/json"
	"qblock"
	"uss"
)

var F int // F，容忍无效或者恶意节点数
var N int // N=3f+1

// Reply消息，由各节点发往客户端
type ReplyMsg struct {
	View        int64                      // 当前视图编号
	Time_stamp  int64                      // 相应request的时间戳
	Client_name string                     // 客户端名称
	Node_i      int64                      // 当前节点编号
	Result      bool                       // 执行request操作的结果
	Sign_i      uss.USSToeplitzHashSignMsg // 当前节点对Commit消息中v,t,c,i,r的签名
	Request     *qblock.Block
}

// PrePrepare消息，由主节点发往从节点
type PrePrepareMsg struct {
	View            int64                      // 当前视图编号
	Sequence_number int64                      // 主节点分配的序列号n,在[h, H]之间，用于对request进行排序？
	Digest_m        []byte                     // 客户端请求消息中m的摘要
	Sign_p          uss.USSToeplitzHashSignMsg // 主节点P对PrePrepare消息中v,n,d的签名
	Request         *qblock.Block
}

// Prepare消息，由从节点发往其他所有节点
type PrepareMsg struct {
	View            int64                      // 当前视图编号
	Sequence_number int64                      // 主节点分配的序列号n,在[h, H]之间，用于对request进行排序？
	Digest_m        []byte                     // 客户端请求消息中m的摘要
	Node_i          int64                      // 当前节点编号
	Sign_i          uss.USSToeplitzHashSignMsg // 当前从节点i对Prepare消息的签名
}

// Commit消息，由任一节点发往其他各节点
type CommitMsg struct {
	View            int64                      // 当前视图编号
	Sequence_number int64                      // 主节点分配的序列号n,在[h, H]之间，用于对request进行排序？
	Digest_m        []byte                     // 客户端请求消息中m的摘要
	Node_i          int64                      // 当前节点编号
	Sign_i          uss.USSToeplitzHashSignMsg // 当前从节点i对Commit消息的签名
}

// PrePrepareMsg.signMessageEncode,对预准备消息编码，形成待签名消息
// 参数：预准备消息PrePrepareMsg
// 返回值：待签名消息[]byte
func (obj *PrePrepareMsg) signMessageEncode() ([]byte, error) {
	type PrePrepare struct {
		View            int64
		Sequence_number int64
		Digest_m        []byte
	}
	preprepare := PrePrepare{
		View:            obj.View,
		Sequence_number: obj.Sequence_number,
		Digest_m:        obj.Digest_m,
	}

	jsonMsg, err := json.Marshal(preprepare) // 将msg信息编码成json格式
	if err != nil {
		return nil, err
	}
	return jsonMsg, nil
}

// PrepareMsg.signMessageEncode,对准备消息编码，形成待签名消息
// 参数：准备消息PrepareMsg
// 返回值：待签名消息[]byte
func (obj *PrepareMsg) signMessageEncode() ([]byte, error) {
	type Prepare struct {
		View            int64
		Sequence_number int64
		Digest_m        []byte
		Node_i          int64
	}
	prepare := Prepare{
		View:            obj.View,
		Sequence_number: obj.Sequence_number,
		Digest_m:        obj.Digest_m,
		Node_i:          obj.Node_i,
	}
	jsonMsg, err := json.Marshal(prepare) // 将msg信息编码成json格式
	if err != nil {
		return nil, err
	}
	return jsonMsg, nil
}

// CommitMsg.signMessageEncode,对提交消息编码，形成待签名消息
// 参数：提交消息CommitMsg
// 返回值：待签名消息[]byte
func (obj *CommitMsg) signMessageEncode() ([]byte, error) {
	type Commit struct {
		View            int64
		Sequence_number int64
		Digest_m        []byte
		Node_i          int64
	}
	commit := Commit{
		View:            obj.View,
		Sequence_number: obj.Sequence_number,
		Digest_m:        obj.Digest_m,
		Node_i:          obj.Node_i,
	}
	jsonMsg, err := json.Marshal(commit) // 将msg信息编码成json格式
	if err != nil {
		return nil, err
	}
	return jsonMsg, nil
}

// ReplyMsg.signMessageEncode,对应答消息编码，形成待签名消息
// 参数：应答消息ReplyMsg
// 返回值：待签名消息[]byte
func (obj *ReplyMsg) signMessageEncode() ([]byte, error) {
	type Reply struct {
		View        int64
		Time_stamp  int64
		Client_name string
		Node_i      int64
		Result      bool
	}
	reply := Reply{
		View:        obj.View,
		Time_stamp:  obj.Time_stamp,
		Client_name: obj.Client_name,
		Node_i:      obj.Node_i,
		Result:      obj.Result,
	}
	jsonMsg, err := json.Marshal(reply) // 将msg信息编码成json格式
	if err != nil {
		return nil, err
	}
	return jsonMsg, nil
}
