package main

import (
	"encoding/hex"
	"fmt"
	"qb/qkdserv"
	"qb/uss"
	"testing"
)

// 测试uss
func TestUSS(t *testing.T) {
	fmt.Println("-----------------------------------------------------------------------------------------------------------")
	fmt.Println("--------------------------!!!!USS测试!!!!--------------------------------------------------------------")
}

// 测试接口函数一：签名
func TestUSSign(t *testing.T) {
	fmt.Println("----------【接口函数测试一】——Sign--------------------------------------------------------------------")

	// 定义签名索引
	SignIndex := qkdserv.QKDSignMatrixIndex{}
	id := []byte("XHSGDFAYQHDJ2163")
	for i := 0; i < 16; i++ {
		SignIndex.Sign_dev_id[i] = id[i]
	}
	SN := uss.GenSignTaskSN(16)
	for i := 0; i < 16; i++ {
		SignIndex.Sign_task_sn[i] = SN[i]
	}
	// 签名消息m
	var m [1024]byte
	for i := 0; i < 1024; i++ {
		m[i] = byte(i)
	}
	uss := uss.Sign(SignIndex, 4, 16, m)
	/*fmt.Println(uss.Sign_counts)
	fmt.Println(uss.Sign_len)
	fmt.Println(hex.EncodeToString(uss.Message[:]))*/
	fmt.Println(hex.EncodeToString(uss.Sign))
}

// 测试接口函数二：验签
func TestUSSVerifySign(t *testing.T) {
	fmt.Println("----------【接口函数测试二】——VerifySign--------------------------------------------------------------------")
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	// 定义使用该程序的参与者名称，正常使用时，该参数由命令行输入，此处只是为了测试使用
	qkdserv.Node_name = [2]byte{'P', '1'}
	// 定义签名索引
	SignIndex := qkdserv.QKDSignMatrixIndex{}
	id := []byte("XHSGDFAYQHDJ2163")
	for i := 0; i < 16; i++ {
		SignIndex.Sign_dev_id[i] = id[i]
	}
	SN := uss.GenSignTaskSN(16)
	for i := 0; i < 16; i++ {
		SignIndex.Sign_task_sn[i] = SN[i]
	}
	// 定义签名信息
	var m [1024]byte
	for i := 0; i < 1024; i++ {
		m[i] = byte(i)
	}
	uss_sign := uss.Sign(SignIndex, 4, 16, m)
	uss_sign.Main_row_num.Sign_Node_Name = [2]byte{'C', '1'}
	//fmt.Println(hex.EncodeToString(uss_sign.Sign))
	// 验签
	result := uss.VerifySign(uss_sign)
	fmt.Println(result)

}
