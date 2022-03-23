package qkdserv

// QKD种子密钥，用于生成安全随机数，可更改
const QKD_KEY = "QKD simulation"

// 签名密钥池： key=签名密钥索引，value=签名密钥矩阵
var QKD_sign_random_matrix_pool map[QKDSignMatrixIndex]QKDSignRandomsMatrix

// 调用该程序的参与者名称，主要有两类参与者：联盟节点(P1、P2...),客户端(C1、C2...)
var Node_name string

// 签名密钥矩阵的一行
type QKDSignRandomsMatrixRow struct {
	Counts     uint32 // 该随机数的个数（是方形矩阵，等于总行数/列数）
	Row_num    uint32 // 随机数所在行的行号
	Column_num uint32 // 随机数所在列的列号，随机数个数counts>1时，列号无用
	Unit_len   uint32 // 每个随机数的长度，字节为单位
	Randoms    []byte // 随机数
}

// 签名密钥矩阵
type QKDSignRandomsMatrix struct {
	Main_row_num byte                      // 主行号，即有多列的行，为0，表示是全矩阵，否则为只有一行为主的矩阵
	Row_counts   uint32                    // 随机数行数
	Sign_randoms []QKDSignRandomsMatrixRow // 每行随机数
}

// 签名矩阵索引，由id和SN可以唯一标识该次签名，该索引也是签名密钥索引
type QKDSignMatrixIndex struct {
	Sign_dev_id  [16]byte // 签名者id
	Sign_task_sn [16]byte // 本次签名的序列号
}

// 签名密钥主行号信息
type QKDSignRandomMainRowNum struct {
	Sign_node_name    string // 签名者的节点号，用于确定主行号
	Main_row_num      uint32 // 主行号，可以自定义，也可由节点号生成
	Random_row_counts uint32 // 每行随机数的个数，可选参数
	Random_unit_len   uint32 // 每个随机数的字节长度，可选参数
}
