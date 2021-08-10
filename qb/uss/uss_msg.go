package uss

import "qb/qkdserv"

// 用于生成toeplitz字符串
const TOEPLITZ_KEY = "Toeplitz Matrix"

// toeplitz矩阵，每次签名、验签用同一个toeplitz矩阵
var Toeplitz_Matrix USSToeplitzMatrixMsg

// USSToeplitzMatrixMsg，包含生成toeplitz矩阵需要的参数
type USSToeplitzMatrixMsg struct {
	Row_counts      uint32         // 矩阵行数，=单位签名长度=单位密钥长度
	Column_counts   uint32         // 矩阵列数，>=签名消息长度
	Toeplitz_matrix [16][1024]uint // 存放toeplitz矩阵，行数默认为<=16，列数<=1024
}

// USSToeplitzHashSignMsg,用于存放签名、验签所需的参数
type USSToeplitzHashSignMsg struct {
	Sign_index   qkdserv.QKDSignMatrixIndex      // 签名索引
	Main_row_num qkdserv.QKDSignRandomMainRowNum // 主行号
	Sign_counts  uint32                          // 每行签名个数，=验签者数量
	Sign_len     uint32                          // 签名单位长度，=密钥单位长度（以字节为单位）
	Message      [1024]byte                      // 签名消息，默认<=1024字节
	Sign         []byte                          // 签名消息，默认<=16字节
}
