package main

import (
	"encoding/hex"
	"fmt"
	"qb/qkdserv"
	"qb/uss"
	"testing"
)

// 测试qkdserv
func TestQKDServ(t *testing.T) {
	fmt.Println("-----------------------------------------------------------------------------------------------------------")
	fmt.Println("--------------------------!!!!QKDServ测试!!!!--------------------------------------------------------------")
}

// 测试接口函数一：共享安全随机数
func TestQKDSecRandomShare(t *testing.T) {
	fmt.Println("----------【接口函数测试一】——QKDSecRandomShare--------------------------------------------------------------------")

}

// 测试接口函数二：读取安全随机数
func TestQKDReadSecRandom(t *testing.T) {
	fmt.Println("----------【接口函数测试二】——QKDReadSecRandom--------------------------------------------------------------------")
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	var i uint32
	// 定义使用该程序的参与者名称，正常使用时，该参数由命令行输入，此处只是为了测试使用
	qkdserv.Node_name = []byte("P1")
	// 定义签名索引
	SignIndex := qkdserv.QKDSignMatrixIndex{}
	id := []byte("XHSGDFAYQHDJ2163")
	for i = 0; i < 16; i++ {
		SignIndex.Sign_dev_id[i] = id[i]
	}
	SN := uss.GenSignTaskSN(16)
	for i = 0; i < 16; i++ {
		SignIndex.Sign_task_sn[i] = SN[i]
	}
	// 定义主行号相关信息
	SignMainRowNum := qkdserv.QKDSignRandomMainRowNum{}
	SignMainRowNum.Sign_Node_Name = []byte("C1")
	SignMainRowNum.Main_Row_Num = 0 // 主行号默认设置为0
	SignMainRowNum.Counts = 4
	SignMainRowNum.Unit_len = 16

	// 读取安全随机数
	qkdserv.QKD_sign_random_matrix_pool[SignIndex] = qkdserv.QKDReadSecRandom(SignIndex, SignMainRowNum)
	PrintVerifyMatrix(qkdserv.QKD_sign_random_matrix_pool[SignIndex])
}

// 打印签名密钥矩阵
func PrintVerifyMatrix(VerifyMatrix qkdserv.QKDSignRandomsMatrix) {
	// fmt.Println("	Main_row_num=", VerifyMatrix.Main_row_num)
	// fmt.Println("	Row_counts=", VerifyMatrix.Row_counts)
	for i := 0; i < int(VerifyMatrix.Row_counts); i++ {
		fmt.Printf("	第%d行,第%d列随机数=", i+1, VerifyMatrix.Sign_randoms[i].Column_num)
		fmt.Println(hex.EncodeToString(VerifyMatrix.Sign_randoms[i].Randoms))
	}
}
