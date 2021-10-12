package qkdserv

import (
	"encoding/hex"
	"fmt"
	"testing"
)

// 测试接口函数一：共享安全随机数
func TestQKDSecRandomShare(t *testing.T) {
	fmt.Println("----------【QKDserv】——QKDSecRandomShare-----------------------------------------------------------")
}

// 测试接口函数二：读取安全随机数
func TestQKDReadSecRandom(t *testing.T) {
	fmt.Println("----------【QKDserv】——QKDReadSecRandom------------------------------------------------------------")
	// 初始化签名密钥池
	QKD_sign_random_matrix_pool = make(map[QKDSignMatrixIndex]QKDSignRandomsMatrix)
	Node_name = "P163" // 定义使用该程序的参与者名称
	// 定义签名索引
	SignIndex := QKDSignMatrixIndex{}
	id := []byte("XHSGDFAYQHDJ2163")
	for i := 0; i < 16; i++ {
		SignIndex.Sign_dev_id[i] = id[i]
	}
	SN := []byte("1234567890216323")
	for i := 0; i < 16; i++ {
		SignIndex.Sign_task_sn[i] = SN[i]
	}
	// 定义主行号相关信息
	SignMainRowNum := QKDSignRandomMainRowNum{
		Sign_node_name:    "P12",
		Main_row_num:      0,
		Random_row_counts: 4,
		Random_unit_len:   16,
	}
	// 读取安全随机数
	QKD_sign_random_matrix_pool[SignIndex] = QKDReadSecRandom(SignIndex, SignMainRowNum)
	printVerifyMatrix(QKD_sign_random_matrix_pool[SignIndex])
}

// 打印签名密钥矩阵
func printVerifyMatrix(verify_matrix QKDSignRandomsMatrix) {
	fmt.Println("	Main_row_num=", verify_matrix.Main_row_num)
	// fmt.Println("	Row_counts=", VerifyMatrix.Row_counts)
	for i := 0; i < int(verify_matrix.Row_counts); i++ {
		fmt.Printf("	第%d行,第%d列随机数=", i+1, verify_matrix.Sign_randoms[i].Column_num)
		fmt.Println(hex.EncodeToString(verify_matrix.Sign_randoms[i].Randoms))
	}
}
