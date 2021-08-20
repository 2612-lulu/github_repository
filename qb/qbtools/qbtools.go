//qbtools包，包含量子安全区块连原型系统中常用的一些函数
//创建人：zhanglu
//创建时间：2021/08/04
package qbtools

import (
	"crypto/hmac"
	"crypto/sha256"
)

// GenRandomWithPRF,生成随机数（密钥）：根据种子密钥和签名索引产生符合要求的随机数
// 参数：种子密钥[]byte，签名索引QKDSignMatrixIndex，每行随机数个数uint32，随机数的单位字节长度uint32
// 返回值：随机数字节长度uint32，随机数[]byte
func GenRandomWithPRF(key []byte, sign_dev_id, sign_task_sn [16]byte, random_counts uint32, unit_len uint32) (uint32, []byte) {
	// 计算轮次数，random_len字节一个随机数，需要random_counts*unit_len字节，一轮sha256，产生32字节，+1就足够
	randoms_len := random_counts * unit_len
	rounds := (randoms_len)/32 + 1

	// 签名索引连接成一个[]byte
	data := append(sign_dev_id[:], sign_task_sn[:]...)

	hmac_sha256 := hmac.New(sha256.New, key)
	hmac_sha256.Write(data)
	hmac_r := hmac_sha256.Sum(nil)

	var randoms []byte

	// 多轮计算
	for i := 0; i < int(rounds); i++ {
		hmac_sha256 = hmac.New(sha256.New, key)
		hmac_sha256.Write(hmac_r)
		hmac_r = hmac_sha256.Sum(nil)
		randoms = append(randoms, hmac_r...) // 多轮随机数连接
	}
	signrandoms := randoms[0:randoms_len]
	return randoms_len, signrandoms
}

func GetNodeIDTable(nodeName [2]byte) [16]byte {
	NodeIDTable := make(map[[2]byte][16]byte)
	NodeTable := map[[2]byte]string{
		{'P', '1'}: "FHDG3489EYUWHBHD",
		{'P', '2'}: "JGOJFOIJHRNB2346",
		{'P', '3'}: "FIHFUIB376486821",
		{'P', '4'}: "3748HFIYYHFIGFI3",
		{'C', '1'}: "CHIKGHBUIGHB3468",
		{'C', '2'}: "HDFUIY68687NKFJH",
	}
	id, ok := NodeTable[nodeName]
	if ok {
		var NodeID [16]byte
		ID := []byte(id)
		for i := 0; i < 16; i++ {
			NodeID[i] = ID[i]
		}
		NodeIDTable[nodeName] = NodeID
	}
	return NodeIDTable[nodeName]
}
