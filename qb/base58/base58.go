// base58包，用于进行base58编码，计算方式为辗转相除法
// 创建人：zhanglu
// 创建时间：2021/09/14
package base58

import (
	"bytes"
	"math/big"
	"qb/qbtools"
)

// base58编码基础数组，去掉了l、I、0、O这几个书写时容易混淆的字母
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode，计算base58编码
func Base58Encode(input []byte) []byte {
	var result []byte
	x := big.NewInt(0).SetBytes(input)
	// 计算除数
	base := big.NewInt(int64(len(b58Alphabet)))
	// 获取big.Int类型0
	zero := big.NewInt(0)
	// 用于存储余数，big.Int类型可以支持更大的整数
	mod := &big.Int{}
	// 只要被除数不为0，就继续计算
	for x.Cmp(zero) != 0 {
		// 求余运算，x=商值/被除数，base=除数，mod=余数
		x.DivMod(x, base, mod)
		// 取出编码，存储到result中
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}
	// 将结果逆序
	qbtools.ReverseBytes(result)

	return result
}

// Base58Decode,计算base58解码
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}
