package pbft

import (
	"bytes"
	"encoding/binary"
	"qblock"
	"uss"
)

const F = 1       // F，容忍无效或者恶意节点数
const N = 3*F + 1 // N=3F+1，本程序中N=4

// Reply消息，由各节点发往客户端
type ReplyMsg struct {
	View        int64                      // 当前视图编号
	Time_stamp  int64                      // 相应request的时间戳
	Client_name string                     // 客户端名称
	Node_i      int64                      // 当前节点编号
	Result      bool                       // 执行request操作的结果
	Sign_i      uss.USSToeplitzHashSignMsg // 当前节点对Commit消息中v,t,c,i,r的签名
	Request     qblock.Block
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
func (obj *PrePrepareMsg) signMessageEncode() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	return buf.Bytes()
}

// PrepareMsg.signMessageEncode,对准备消息编码，形成待签名消息
// 参数：准备消息PrepareMsg
// 返回值：待签名消息[]byte
func (obj *PrepareMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

// CommitMsg.signMessageEncode,对提交消息编码，形成待签名消息
// 参数：提交消息CommitMsg
// 返回值：待签名消息[]byte
func (obj *CommitMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Sequence_number)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	return buf.Bytes(), nil
}

// ReplyMsg.signMessageEncode,对应答消息编码，形成待签名消息
// 参数：应答消息ReplyMsg
// 返回值：待签名消息[]byte
func (obj *ReplyMsg) signMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.View)
	binary.Write(buf, binary.LittleEndian, obj.Time_stamp)
	binary.Write(buf, binary.LittleEndian, []byte(obj.Client_name))
	binary.Write(buf, binary.LittleEndian, obj.Node_i)
	binary.Write(buf, binary.LittleEndian, obj.Result)
	return buf.Bytes(), nil
}
