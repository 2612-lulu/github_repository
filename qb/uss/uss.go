//uss包，提供无条件安全签名和验签功能
//创建人：zhanglu
//创建时间：2021/08/04
package uss

import (
	"bytes"
	cryptorand "crypto/rand"
	"io"
	"qb/qbtools"
	"qb/qkdserv"
)

// USS，签名
// 参数：签名索引qkdserv.QKDSignMatrixIndex,每行签名个数uint32，签名单位长度uint32，待签名消息[1024]byte
// 返回值：签名信息USSToeplitzHashSignMsg
func Sign(sign_index qkdserv.QKDSignMatrixIndex, counts,
	unit_len uint32, m [1024]byte) USSToeplitzHashSignMsg {
	// 1.密钥分发
	_, randoms := qbtools.GenRandomWithPRF([]byte(qkdserv.QKD_KEY),
		sign_index.Sign_dev_id, sign_index.Sign_task_sn,
		counts*counts, unit_len) // 产生随机数
	//random_share_result := qkdserv.QKDSecRandomShare() //分发随机数

	// 2.USS签名
	Toeplitz_Matrix = generateToeplitz(sign_index, 16, 1024)
	uss_sign := USSToeplitzHashSignMsg{}
	uss_sign.Sign_index = sign_index
	uss_sign.Main_row_num.Counts = counts
	uss_sign.Main_row_num.Unit_len = unit_len
	uss_sign.Sign_counts = counts
	uss_sign.Sign_len = unit_len
	uss_sign.Message = m
	uss_sign.Sign = ussToeplitzHashSign(Toeplitz_Matrix, randoms, m, counts, unit_len)
	return uss_sign
}

// VerifySign，验签
// 参数：签名索引qkdserv.QKDSignMatrixIndex,签名个数uint32，签名单位长度uint32，待签名消息[1024]byte
// 返回值：验签结果bool
func VerifySign(uss_sign USSToeplitzHashSignMsg) bool {
	Toeplitz_Matrix = generateToeplitz(uss_sign.Sign_index, 16, 1024)

	// 1.先获取签名密钥
	verify_random_matrix := qkdserv.QKDReadSecRandom(uss_sign.Sign_index,
		uss_sign.Main_row_num)

	// 2.验签
	j := 0
	for i := 0; i < int(verify_random_matrix.Row_counts); i++ {
		// 计算签名值
		verify_sign := ussToeplitzHashSign(Toeplitz_Matrix,
			verify_random_matrix.Sign_randoms[i].Randoms,
			uss_sign.Message, 1, uss_sign.Sign_len)
		// 取出对应位置的签名值
		row := verify_random_matrix.Sign_randoms[i].Row_num
		column := verify_random_matrix.Sign_randoms[i].Column_num
		start := ((row-1)*uss_sign.Sign_counts + column - 1) * uss_sign.Sign_len
		end := ((row-1)*uss_sign.Sign_counts + column) * uss_sign.Sign_len
		sign := uss_sign.Sign[start:end]
		// 比较签名值与验签值
		if bytes.Equal(verify_sign, sign) {
			j++
		}
	}
	// 3.判断验签结果
	var verifysign_result bool
	var delta float32 = 0.75
	if float32(j) >= (float32(0.5)+2*(1-delta))*float32(uss_sign.Sign_counts) {
		verifysign_result = true
	} else {
		verifysign_result = false
	}
	return verifysign_result
}

// GenSignTaskSN，产生指定字节长度的随机数，主要可做签名序列号（一般为16字节）
// 参数：随机数长度uint32
// 返回值：特定长度的随机数[]byte
func GenSignTaskSN(length uint32) []byte {
	sign_task_sn := make([]byte, length)
	io.ReadFull(cryptorand.Reader, sign_task_sn)
	return sign_task_sn
}

// generateToeplitz，生成toeplitz矩阵
// 参数：签名索引QKDSignMatrixIndex，矩阵行数uint32；矩阵列数uint32
// 返回值：USSToeplitzMatrix，每次签名使用同一个toeplitz矩阵
func generateToeplitz(signindex qkdserv.QKDSignMatrixIndex, m, n uint32) USSToeplitzMatrixMsg {
	toeplitz_matrix := USSToeplitzMatrixMsg{}
	toeplitz_matrix.Row_counts = m    // 矩阵行数
	toeplitz_matrix.Column_counts = n // 矩阵列数

	// 生成长为m+n-1字节的随机数，用于生成矩阵
	_, s := qbtools.GenRandomWithPRF([]byte(TOEPLITZ_KEY),
		signindex.Sign_dev_id, signindex.Sign_task_sn, 1, m+n-1)

	// 用前面生成的随机数生成矩阵
	for i := 0; i < int(m); i++ {
		for j := 0; j < int(n); j++ {
			toeplitz_matrix.Toeplitz_matrix[i][j] = uint(s[int(m)-1-i+j])
		}
	}
	return toeplitz_matrix
}

// ussToeplitzHashSign，签名
// 参数：toeplitz矩阵USSToeplitzMatrixMsg，密钥[]byte，待签名消息+每行签名个数+签名单位长度USSToeplitzHashSignMsg
// 返回值：签名结果[]byte
func ussToeplitzHashSign(toeplitz_matrix USSToeplitzMatrixMsg,
	r []byte, m [1024]byte, counts, len uint32) []byte {
	uss_sign := USSToeplitzHashSignMsg{}
	//uss_sign.Sign_counts = counts
	//uss_sign.Sign_len = len
	//uss_sign.Message = m

	// 签名个数
	sign_number := int(counts * counts)

	// 逐次签名
	for i := 0; i < sign_number; i++ {
		start := i * int(len)
		end := (i + 1) * int(len)
		topelitz_m := toeplitzMatrixMultiply(toeplitz_matrix, m)
		random := r[start:end]
		s := toeplitzMatrixAnd(topelitz_m, random)
		uss_sign.Sign = append(uss_sign.Sign, s...)
	}
	return uss_sign.Sign
}

// toeplitzMatrixMultiply，toeplitz矩阵乘法
// 参数：toeplitz矩阵USSToeplitzMatrix，消息[1024]byte
// 返回值：运算结果[16]byte
func toeplitzMatrixMultiply(toeplitz_matrix USSToeplitzMatrixMsg, m [1024]byte) [16]byte {
	var row_result uint
	var result [16]byte
	for i := 0; i < int(toeplitz_matrix.Row_counts); i++ {
		row_result = 0
		for j := 0; j < int(toeplitz_matrix.Column_counts); j++ {
			row_result = (uint(m[j])*toeplitz_matrix.Toeplitz_matrix[i][j] + row_result) % 0xFF
		}
		result[i] = byte(row_result)
	}
	return result
}

// toeplitzMatrixAnd，异或
// 参数：矩阵乘法结果[16]byte，密钥[]byte
// 返回值：运算结果[]byte
func toeplitzMatrixAnd(toeplitz_m [16]byte, random []byte) []byte {
	var result [16]byte
	var r [16]byte
	for i := 0; i < len(random); i++ {
		r[i] = random[i]
	}
	for i := 0; i < 16; i++ {
		result[i] = byte(uint(toeplitz_m[i]) ^ uint(r[i]))
	}
	return result[:]
}
