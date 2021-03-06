package qbtx

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

// TXOutput，交易输出结构
type TXOutput struct {
	TX_value int    `json:"TXValue"` // 输出金额
	TX_dst   string `json:"TXdst"`   // 接收方
}

// TXOutputs
type TXOutputs struct {
	Outputs []TXOutput
}

// NewTXOutput，初始化交易输出项
// 参数：交易数值int，接收方地址string
// 返回值：交易输出项
func NewTXOutput(tx_value int, tx_dst string) TXOutput {
	txo := TXOutput{tx_value, tx_dst}
	return txo
}

// SerializeOutput，交易输出项序列化
// 参数：待序列化的交易输出项
// 返回值：序列化结果
func (tx_outputs TXOutputs) SerializeOutputs() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(tx_outputs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// DeserializeOutputs，交易输出项反序列化
// 参数：序列化结果
// 返回值：反序列化的交易输出项数组
func DeserializeOutputs(data []byte) TXOutputs {
	var tx_outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&tx_outputs)

	if err != nil {
		fmt.Println("outputs=", tx_outputs)
		log.Panic(err)
	}
	return tx_outputs
}
