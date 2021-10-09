package qbtx

import (
	"bytes"
	"encoding/gob"
	"log"
)

// TXOutput，交易输出结构
type TXOutput struct {
	Value int    `json:"value"` // 输出金额
	To    string `json:"to"`    // 接收方
}

// NewTXOutput，初始化交易输出项
// 参数：交易数值int，接收方地址string
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, address}
	return txo
}

// TXOutputs
type TXOutputs struct {
	Outputs []TXOutput
}

// SerializeOutput，交易输出项序列化
func (outs TXOutputs) SerializeOutputs() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// DeserializeOutputs，交易输出项反序列化
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}
	return outputs
}
