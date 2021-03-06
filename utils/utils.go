//utils包，包含量子安全区块连原型系统中常用的一些函数
//创建人：zhanglu
//创建时间：2021/08/04
package utils

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// 配置文件路径
const INIT_PATH = "../config/"

const FLOW_PATH = "../log/qbc_system_flow_log/"
const SIGN_PATH = "../log/uss_log/"
const VERIFY_PATH = "../log/verify_log/"

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

// GetNodeID，获取节点设备号
// 参数：节点名称string
// 返回值：节点设备号[16]byte
func GetNodeID(node_name string) [16]byte {
	node_id_table := make(map[string][16]byte)
	node_table := InitConfig(INIT_PATH + "id_table.txt")
	id, ok := node_table[node_name]
	if ok {
		var node_id [16]byte
		ID := []byte(id)
		for i := 0; i < 16; i++ {
			node_id[i] = ID[i]
		}
		node_id_table[node_name] = node_id
	}
	return node_id_table[node_name]
}

// Digest，摘要函数
// 参数：消息[]byte
// 返回值：摘要值[]byte
func Digest(m []byte) []byte {
	h := sha256.New()
	h.Write(m)
	digest_m := h.Sum(nil)
	return digest_m
}

// InitConfig,读取key=value类型的配置文件
// 参数：配置文件存放路径string
// 返回值：节点/客户端配置信息map[string]string
func InitConfig(path string) map[string]string {
	config := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}

		config[key] = value
	}
	return config
}

// Init_log,初始化log日志存放文件
// 参数：日志存放路径string
// 返回值：初始化处理错误error，初始化成功返回nil
func Init_log(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644) //【如果已经存在，则在尾部添加写】
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return nil, err
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	return logFile, nil
}

// LogStage，在终端打印共识处理过程
// 参数：共识过程string，该共识过程处理情况bool
// 返回值：无，终端输出共识流程
func LogStage(stage string, isDone bool) {
	if isDone {
		fmt.Printf("[STAGE-DONE] %s\n", stage)
	} else {
		fmt.Printf("[STAGE-BEGIN] %s\n", stage)
	}
}

// Send，http信息发送
// 参数：目的地值，待发送消息
// 返回值：无
func Send(url string, msg []byte) {
	buff := bytes.NewBuffer(msg)
	http.Post("http://"+url, "application/json", buff)
}

// ReverseBytes，将字符串逆序
// 参数：目标数据
// 返回值：无
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// IntToHex，将int转换为[]byte
// 参数：待转换数据int64
// 返回值：目标类型数据[]byte
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
