package qbtx

import (
	"bytes"
	"encoding/gob"
	"log"
	"uss"
)

// TXInput，交易输入结构，表示交易金额的来源，一笔交易可以是由一个用户发起，但是是该用户的多笔零钱发起的一定金额的转账；也可以是多个用户发起的一笔共同转账
type TXInput struct {
	Refer_tx_id       []byte                     `json:"ReferTXid"`      // 引用交易ID，一个输入必须引用一个输出
	Refer_tx_id_index int                        `json:"ReferTXidIndex"` // 引用的交易输出编号，引用的交易中具体的某一个output
	TX_uss_sign       uss.USSToeplitzHashSignMsg `json:"TxUssSign"`      // 签名，花钱的人要证明这些钱是属于它的，会在交易确认的时候，校验这个签名
	TX_src            string                     `json:"TXsrc"`          // 交易来源
}

// SerializeInput，交易输入项序列化
// 参数：交易输入项
// 返回值：序列化结果bool
func (tx_input TXInput) SerializeInput() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(tx_input)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeInput，交易输入项反序列化
// 参数：序列化结果[]byte
// 返回值：交易输入项
func DeserializeInput(data []byte) TXInput {
	var tx_input TXInput

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&tx_input)
	if err != nil {
		log.Panic(err)
	}
	return tx_input
}
