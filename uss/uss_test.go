package uss

import (
	"encoding/hex"
	"fmt"
	"testing"

	"qkdserv"
)

// 测试接口函数一：签名
func TestUSSign(t *testing.T) {
	fmt.Println("----------【USS】——Sign------------------------------------------------------------------------")
	// 定义签名索引
	SignIndex := qkdserv.QKDSignMatrixIndex{}
	id := []byte("XHSGDFAYQHDJ2163")
	for i := 0; i < 16; i++ {
		SignIndex.Sign_dev_id[i] = id[i]
	}
	SN := GenSignTaskSN(16)
	for i := 0; i < 16; i++ {
		SignIndex.Sign_task_sn[i] = SN[i]
	}
	// 签名消息m
	m := []byte("afhigehrfgui")
	s := UnconditionallySecureSign(SignIndex, 4, 16, m)
	fmt.Println("sign message:", hex.EncodeToString(s.USS_message[:]))
	fmt.Println("uss signature:", hex.EncodeToString(s.USS_signature))
}

// 测试接口函数二：验签
func TestUSSVerifySign(t *testing.T) {
	fmt.Println("----------【USS】——VerifySign------------------------------------------------------------------")
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	// 定义使用该程序的参与者名称，正常使用时，该参数由命令行输入，此处只是为了测试使用
	qkdserv.Node_name = "P1"
	// 定义签名索引
	SignIndex := qkdserv.QKDSignMatrixIndex{}
	id := []byte("XHSGDFAYQHDJ2163")
	for i := 0; i < 16; i++ {
		SignIndex.Sign_dev_id[i] = id[i]
	}
	SN := GenSignTaskSN(16)
	for i := 0; i < 16; i++ {
		SignIndex.Sign_task_sn[i] = SN[i]
	}
	// 定义签名信息
	m := []byte("4379765")
	uss_sign := UnconditionallySecureSign(SignIndex, 4, 16, m)
	uss_sign.Main_row_num.Sign_node_name = "C1"
	//fmt.Println(hex.EncodeToString(uss_sign.Sign))
	// 验签
	result := UnconditionallySecureVerifySign(uss_sign)
	fmt.Println("result of verify sign:", result)

}
