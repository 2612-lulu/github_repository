package qbtx

import (
	"fmt"
	"qkdserv"
	"testing"
	"uss"
)

func TestQBTX(t *testing.T) {
	fmt.Println("----------【Transaciton】——SignTX && VerifyTXsign------------------------------------------------------------")
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	qkdserv.Node_name = "C1"
	N = 4 //定义验签者数量
	txInput := TXInput{
		Refer_tx_id:       []byte("egry"),
		Refer_tx_id_index: 1,
		TX_uss_sign:       uss.USSToeplitzHashSignMsg{},
		TX_src:            "C1",
	}
	var Inputs []TXInput
	Inputs = append(Inputs, txInput)
	txOutput := TXOutput{
		TX_value: 1,
		TX_dst:   "P3",
	}
	var outputs []TXOutput
	outputs = append(outputs, txOutput)
	tx := Transaction{
		TX_id:   nil,
		TX_vin:  Inputs,
		TX_vout: outputs,
	}
	tx.SetID()
	tx.USSTransactionSign("C1")
	fmt.Println("sign success")
	qkdserv.Node_name = "P1"
	if tx.VerifyUSSTransactionSign() {
		fmt.Println("sign and verify of tx success")
	}

}
