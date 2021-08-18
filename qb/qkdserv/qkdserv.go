//qkdserv包，模拟QKD服务基本功能，包括密钥分发与获取密钥
//创建人：zhanglu
//创建时间：2021/08/04
package qkdserv

import (
	"bytes"
	"fmt"
	"qb/qbtools"
	"strconv"
)

// QKDSecRandomShare，密钥分发
// 参数：签名索引QKDSignMatrixIndex，源、目的ID[16]byte，共享的一行随机数QKDSignRandomsMatrixRow，主行号QKDSignRandomsMatrixRow
// 返回值：分发结果bool
func QKDSecRandomShare(sign_matrix_index QKDSignMatrixIndex, random_src_dev_id,
	random_dst_dev_id [16]byte, random_row QKDSignRandomsMatrixRow, main_row_num QKDSignRandomsMatrixRow) bool {
	return true
}

// QKDReadSecRandom，读取共享密钥
// 参数：签名索引QKDSignMatrixIndex，主行号QKDSignRandomsMatrixRow
// 返回值：用于验签的密钥矩阵QKDSignRandomsMatrix
func QKDReadSecRandom(sign_matrix_index QKDSignMatrixIndex, sign_main_row_num QKDSignRandomMainRowNum) QKDSignRandomsMatrix {
	// 获取主行号
	sign_main_row_num.Main_Row_Num = getMainRowNum(sign_main_row_num, Node_name)

	// 格式化检查：id+SN的长度
	if sign_main_row_num.Main_Row_Num == 0 {
		fmt.Println("	qkdserv error：Only erifysigner can use func QKDReadSecRandom()!! ")
	} else if len(sign_matrix_index.Sign_dev_id) != 16 {
		fmt.Println("	qkdserv error：The length of Sign_dev_id is wrong!! ")
	} else if len(sign_matrix_index.Sign_task_sn) != 16 {
		fmt.Println("	qkdserv error：The length of Sign_task_sn is wrong!! ")
	} else {
		// 计算全阵，模拟实现的时候，需要通过全局种子和sign_matrix_index值来首先计算出随机数矩阵
		sign_randoms_matrix := generateSignRandomsMatrix(sign_matrix_index, sign_main_row_num.Counts, sign_main_row_num.Unit_len)

		// 获得签名密钥矩阵（残阵）
		QKD_sign_random_matrix_pool[sign_matrix_index] = getVerifyMatrix(sign_matrix_index, sign_randoms_matrix, sign_main_row_num)
	}

	return QKD_sign_random_matrix_pool[sign_matrix_index]
}

// generateSignRandomsMatrix，生成签名随机数全阵：通过签名序列号以及所有节点共享的秘密值，产生全矩阵
// 参数：签名索引QKDSignMatrixIndex（id+SN），每行随机数个数uint32，随机数的单位字节长度uint32
// 返回值：随机数矩阵QKDSignRandomsMatrix
func generateSignRandomsMatrix(sign_matrix_index QKDSignMatrixIndex, row_counts uint32, unit_len uint32) QKDSignRandomsMatrix {
	// 计算矩阵中随机数的总个数
	randoms_counts := row_counts * row_counts

	// 用密钥和随机数序列号生成随机数
	_, randoms := qbtools.GenRandomWithPRF([]byte(QKD_KEY), sign_matrix_index.Sign_dev_id, sign_matrix_index.Sign_task_sn, randoms_counts, unit_len)

	// 定义签名密钥全阵
	sign_matrix := QKDSignRandomsMatrix{}
	sign_matrix.Main_row_num = 0        // 表示是全阵
	sign_matrix.Row_counts = row_counts // 验签者个数

	for i := 0; i < int(row_counts); i++ {
		curr_row := QKDSignRandomsMatrixRow{}
		curr_row.Counts = row_counts     // 完整行，随机数数量等于行数/列数
		curr_row.Row_num = uint32(i + 1) // 当前行的行号
		curr_row.Column_num = 0          // 完整行，列号无意义
		curr_row.Unit_len = unit_len     // 每个随机数unit_len字节

		// 取出一行
		start := int(row_counts*unit_len) * i
		end := int(row_counts*unit_len) * (i + 1)
		s := randoms[start:end]
		curr_row.Randoms = append(curr_row.Randoms, s...)

		sign_matrix.Sign_randoms = append(sign_matrix.Sign_randoms, curr_row)
	}

	return sign_matrix
}

// getMainRowNum，获得主行号
// 参数：主行号信息QKDSignRandomMainRowNum，调用此程序的节点名称[]byte
// 返回值：主行号uint32
func getMainRowNum(SignMainRowNum QKDSignRandomMainRowNum, VerifyNode [2]byte) uint32 {
	var main_row_num uint32
	SignNum, _ := strconv.Atoi(string(SignMainRowNum.Sign_Node_Name[1:])) // 将其转化为数字
	VerifyNum, _ := strconv.Atoi(string(VerifyNode[1:]))

	if bytes.Equal(SignMainRowNum.Sign_Node_Name[:1], []byte("C")) { // 如果签名者是客户端
		if bytes.Equal(VerifyNode[:1], []byte("C")) {
			main_row_num = 0
		} else if bytes.Equal(VerifyNode[:1], []byte("P")) {
			main_row_num = uint32(VerifyNum)
		} else {
			fmt.Println("	qkdserv error:The input arg is wrong!!")
		}
	} else if bytes.Equal(SignMainRowNum.Sign_Node_Name[:1], []byte("P")) { // 如果签名者是联盟节点
		if bytes.Equal(SignMainRowNum.Sign_Node_Name[:], VerifyNode[:]) { // 签名节点调用此程序，则主行号=0
			main_row_num = 0
		} else if bytes.Equal(VerifyNode[:1], []byte("C")) { // 如果验签者是客户端
			main_row_num = 1
		} else if VerifyNum <= SignNum { // 验签者节点下标<=签名者节点下标
			main_row_num = uint32(VerifyNum)
		} else { // 验签者节点下标>签名者节点下标，主行号=验签者节点下标-1
			main_row_num = uint32(VerifyNum - 1)
		}
	} else {
		fmt.Println("	qkdserv error:The name of Sign_dev_id is wrong!!")
	}
	return main_row_num
}

// getVerifyMatrix，获得验签用的密钥矩阵（残阵）:通过全矩阵，得到本节点的残矩阵
// 参数：签名索引QKDSignMatrixIndex，随机数全阵QKDSignRandomsMatrix，主行号信息QKDSignRandomMainRowNum
// 返回值：验签用的随机数残阵QKDSignRandomsMatrix
func getVerifyMatrix(SignIndex QKDSignMatrixIndex, SignRandomsMatrix QKDSignRandomsMatrix, main_row_num QKDSignRandomMainRowNum) QKDSignRandomsMatrix {
	VerifyMatrix := QKDSignRandomsMatrix{}
	VerifyMatrix.Main_row_num = byte(main_row_num.Main_Row_Num) // 主行号
	VerifyMatrix.Row_counts = SignRandomsMatrix.Row_counts      // 矩阵的行数
	SN := int(SignIndex.Sign_task_sn[15])                       // 取SN最后一位字节

	for i := 0; i < int(VerifyMatrix.Row_counts); i++ {
		curr_row := QKDSignRandomsMatrixRow{}
		curr_row.Counts = 1                                                                              // 该行中包含的签名随机数的个数，签名密钥中均为1                                                                   // 随机数个数，因为是验签用的残阵，这里每行有一个随机数
		curr_row.Row_num = uint32(i + 1)                                                                 // 当前行号
		curr_row.Column_num = ((uint32(i+SN) + main_row_num.Main_Row_Num) % VerifyMatrix.Row_counts) + 1 // 列号
		curr_row.Unit_len = main_row_num.Unit_len                                                        // 每个随机数的单位字节长度

		start := int(curr_row.Column_num-1) * int(curr_row.Unit_len)
		end := int(curr_row.Column_num * curr_row.Unit_len)
		s := SignRandomsMatrix.Sign_randoms[i].Randoms[start:end] // 取出这一行中用到的签名密钥
		curr_row.Randoms = append(curr_row.Randoms, s...)

		VerifyMatrix.Sign_randoms = append(VerifyMatrix.Sign_randoms, curr_row)
	}
	return VerifyMatrix
}
