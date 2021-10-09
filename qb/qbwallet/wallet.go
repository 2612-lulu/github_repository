// 定义了钱包结构及区块链帐户地址的生成方法
// 参考P213图5-11

package qbwallet

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"qb/base58"
	"qkdserv"
	"utils"

	"golang.org/x/crypto/ripemd160"
)

// 定义前缀版本
const version = byte(0x00)

// 校验码长度
const addressChecksumLen = 4

// 存储钱包地址的文件
const walletFile = "../qb/qbwallet/wallets/wallet_%s.dat"

//钱包结构
type Wallet struct {
	Node_id [16]byte // 设备id
	Addr    []byte   // 钱包地址
}

// NewWallet，创建钱包
func NewWallet(nodeID string) *Wallet {
	var wallet Wallet
	wallet.Node_id = utils.GetNodeIDTable(nodeID) // 获取设备ID用于生成钱包地址
	wallet.Addr = wallet.getAddress()             // 生成钱包地址
	wallet.saveToFile(qkdserv.Node_name)          // 将钱包地址存入文件
	return &wallet
}

// getAddress，生成地址
func (w Wallet) getAddress() []byte {
	// 1.计算公钥hash
	addr_hash := hashID(w.Node_id[:])

	// 2.计算校验和
	versionedPayload := append([]byte{version}, addr_hash...) // 加入前缀version
	checksum := checksum(versionedPayload)

	// 3.计算base58编码
	fullPayload := append(versionedPayload, checksum...)
	address := base58.Base58Encode(fullPayload)

	return address
}

// hashID，计算QKD设备ID'hash
func hashID(node_id []byte) []byte {
	// 1.先hash一次
	publicSHA256 := sha256.Sum256(node_id)

	// 2.计算ripemd160。需要事先导入ripemd160包
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// checksum，计算校验和。需要计算两次hash，输入的内容是已经加了前缀0x00的公钥hash
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen] // 前4节校验码
}

// SaveToFile，保存方法,把新建的wallet添加进去
func (w Wallet) saveToFile(nodeName string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeName)

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// ValidateAddress，检验地址合法有效性。反解析
func ValidateAddress(address string) bool {
	addr_hash := base58.Base58Decode([]byte(address))
	actualChecksum := addr_hash[len(addr_hash)-addressChecksumLen:]
	version := addr_hash[0]
	addr_hash = addr_hash[1 : len(addr_hash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, addr_hash...))
	if bytes.Equal(actualChecksum, targetChecksum) { // 如果校验成功
		return true
	} else {
		return false
	}
}

// LoadFromFile，把所有的wallet读出来
func (w *Wallet) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	// 在读取之前，要先确认文件是否在，如果不存在，直接退出
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		log.Println("wallet file is not exit")
		return err
	}

	// 读取内容
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	// 解码
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&w)
	if err != nil {
		log.Panic(err)
	}

	return nil
}
