package qbtx

import (
	"fmt"
	"qkdserv"
	"testing"
	"uss"
)

func TestQBTX(t *testing.T) {
	fmt.Println("----------【Transaciton】——Sign------------------------------------------------------------")
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	// 定义使用该程序的参与者名称，正常使用时，该参数由命令行输入，此处只是为了测试使用
	qkdserv.Node_name = "C1"
	txInput := TXInput{
		Txid:      []byte("egry"),
		Vout:      1,
		Signature: uss.USSToeplitzHashSignMsg{},
		From:      "C1",
	}
	var Inputs []TXInput
	Inputs = append(Inputs, txInput)
	txOutput := TXOutput{
		Value: 1,
		To:    "P3",
	}
	var outputs []TXOutput
	outputs = append(outputs, txOutput)
	tx := Transaction{
		ID:   nil,
		Vin:  Inputs,
		Vout: outputs,
	}
	tx.SetID()
	tx.SignTX("C1")
	fmt.Println("sign success")
	qkdserv.Node_name = "P1"
	if tx.VerifyTX() {
		fmt.Println("sign and verify of tx success")
	}

}
