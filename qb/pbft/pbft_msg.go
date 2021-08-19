package pbft

import "qb/uss"

// Request消息，由客户端发往主节点
type RequestMsg struct {
	Time_stamp       int64                      // 时间戳，发出请求时客户端本地时钟的值
	Client_id        [16]byte                   // 客户端标识
	RequestOperation                            // 请求的具体操作
	Sign_client      uss.USSToeplitzHashSignMsg // 客户端对request消息的签名
}

// Request消息中Operation的具体参数
type RequestOperation struct {
	Operation_type []byte // 操作类型，本系统中主要是转账
	M              []byte // 消息
	Digest_m       []byte // 消息m的摘要值
}

// Reply消息，由各节点发往客户端
type ReplyMsg struct {
	View       int64                      // 当前视图编号
	Time_stamp int64                      // 相应request的时间戳
	Client_id  [16]byte                   // 客户端标识
	Node_i     int64                      // 当前节点编号
	Result     bool                       // 执行request操作的结果
	Sign_i     uss.USSToeplitzHashSignMsg // 当前节点对Commit消息中v,t,c,i,r的签名
}

// PrePrepare消息，由主节点发往从节点
type PrePrepareMsg struct {
	View            int64                      // 当前视图编号
	Sequence_number int64                      // 主节点分配的序列号n,在[h, H]之间，用于对request进行排序？
	Digest_m        []byte                     // 客户端请求消息中m的摘要
	Sign_p          uss.USSToeplitzHashSignMsg // 主节点P对PrePrepare消息中v,n,d的签名
	Request         RequestMsg
}

// Prepare消息，由从节点发往其他所有节点
type PrepareMsg struct {
	View            int64                      // 当前视图编号
	Sequence_number int64                      // 主节点分配的序列号n,在[h, H]之间，用于对request进行排序？
	Digest_m        []byte                     // 客户端请求消息中m的摘要
	Node_i          int64                      // 当前节点编号
	Sign_i          uss.USSToeplitzHashSignMsg // 当前从节点i对Prepare消息的签名
}

// Commit消息，由任一节点发往其他各节点
type CommitMsg struct {
	View            int64                      // 当前视图编号
	Sequence_number int64                      // 主节点分配的序列号n,在[h, H]之间，用于对request进行排序？
	Digest_m        []byte                     // 客户端请求消息中m的摘要
	Node_i          int64                      // 当前节点编号
	Sign_i          uss.USSToeplitzHashSignMsg // 当前从节点i对Commit消息的签名
}
