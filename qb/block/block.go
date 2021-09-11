//block包，定义区块链数据结构，实现区块链打包
//创建人：zhanglu
//创建时间：2021/08/25
package block

import (
	"bytes"
	"encoding/binary"
	"time"

	"qb/uss"
)

// 区块包含的最小交易数量
const BLOCK_LENGTH = 5

// 区块结构
type Block struct {
	//Version       int64          // 当前版本
	//Round         int64          // 当前轮数
	Time_stamp int64 // 系统当前时间
	//Num           int64          // 区块序号
	Transactions []*Transaction // 用于共识的交易信息
	//PrevBlockHash []byte         // 前一区块hash值
}

// 用于共识的交易结构
type Transaction struct {
	Time_stamp           int64                      // 时间戳，发出请求时客户端本地时钟的值
	Name                 string                     // 客户端名称
	TransactionOperation                            // 请求的具体操作
	Sign_client          uss.USSToeplitzHashSignMsg // 客户端对request消息的签名
}

// 交易内容
type TransactionOperation struct {
	Transaction_message []byte // 消息
	Digest_m            []byte // 消息m的摘要值
}

// Transaction.signMessageEncode,对交易消息编码，形成待签名消息
// 参数：交易消息Transaction
// 返回值：待签名消息[]byte
func (obj *Transaction) SignMessageEncode() ([]byte, error) {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, obj.Time_stamp)
	binary.Write(buf, binary.LittleEndian, []byte(obj.Name))
	binary.Write(buf, binary.LittleEndian, obj.Transaction_message)
	binary.Write(buf, binary.LittleEndian, obj.Digest_m)
	return buf.Bytes(), nil
}

// CreateBlock，生成区块
// 参数：交易信息数组[]*Transaction
// 返回值：区块*Block
func CreateBlock(transcation []*Transaction) *Block {

	block := Block{
		Time_stamp:   time.Now().UnixNano(),
		Transactions: transcation,
	}
	return &block
}
