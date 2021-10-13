package qbtx

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"qkdserv"
	"uss"
	"utils"
)

// 定义准备金金额
const RESERVE = 20

// 定义验签者数量
var N uint32

// Transaction，交易结构，多入多处：
// 有一些输出并没有被关联到某个输入上；一笔交易的输入可以引用之前多笔交易的输出；一个输入必须引用一个输出
type Transaction struct {
	TX_id   []byte     `json:"TXid"`   // 交易ID，非常重要的Hash值，是在input签名之前计算出来的，作为UTXOSet.map的key存在
	TX_vin  []TXInput  `json:"TXvin"`  // 交易输入项
	TX_vout []TXOutput `json:"TXvout"` // 交易输出项
}

// SetID，根据交易输入与输出项生成交易ID。
func (tx Transaction) SetID() []byte {
	var hash [32]byte
	tx_copy := tx
	tx_copy.TX_id = []byte{}
	hash = sha256.Sum256(tx_copy.SerializeTX())
	return hash[:]
}

// Serialize，交易序列化
func (tx Transaction) SerializeTX() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

// DeserializeTX，交易反序列化
func DeserializeTX(data []byte) Transaction {
	var tx Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		log.Panic(err)
	}
	return tx
}

// USSTransactionSign，对交易输入项签名
func (tx Transaction) USSTransactionSign(node_name string) {
	tx_copy := tx.TrimmedCopyTX() // 复制并修剪交易以得到待签名数据

	for in_id, input := range tx_copy.TX_vin { // 循环向输入项签名
		tx_copy.TX_vin[in_id].TX_uss_sign = uss.USSToeplitzHashSignMsg{} // 置空
		data_to_sign := input.SerializeInput()                           // 待签名数据
		signature := uss.USSToeplitzHashSignMsg{
			Sign_index: qkdserv.QKDSignMatrixIndex{
				Sign_dev_id:  utils.GetNodeID(node_name),
				Sign_task_sn: uss.GenSignTaskSN(16),
			},
			Main_row_num: qkdserv.QKDSignRandomMainRowNum{
				Sign_node_name:    node_name,
				Main_row_num:      0, // 签名主行号，签名时默认为0
				Random_row_counts: N,
				Random_unit_len:   16,
			},
			USS_counts:   N,
			USS_unit_len: 16,
			USS_message:  data_to_sign,
		}
		signature = uss.UnconditionallySecureSign(signature.Sign_index, signature.USS_counts, signature.USS_unit_len, signature.USS_message)
		tx.TX_vin[in_id].TX_uss_sign = signature
		//tx_copy.TX_vin[in_id].TX_src = "" // 将From置空
	}
}

// VerifyUSSTransactionSign,交易输入项验签
func (tx Transaction) VerifyUSSTransactionSign() bool {
	txCopy := tx

	for inID, _ := range tx.TX_vin {
		if !uss.UnconditionallySecureVerifySign(txCopy.TX_vin[inID].TX_uss_sign) {
			fmt.Println("verify of tx wrong")
		}
	}

	return true
}

// TrimmedCopyTX，交易修剪以得到待签名消息
func (tx Transaction) TrimmedCopyTX() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.TX_vin { // 将原交易内的签名和公钥都置空
		inputs = append(inputs, TXInput{vin.Refer_tx_id, vin.Refer_tx_id_index, uss.USSToeplitzHashSignMsg{}, ""})
	}

	for _, vout := range tx.TX_vout { // 复制原输入项
		outputs = append(outputs, TXOutput{vout.TX_value, vout.TX_dst})
	}

	txCopy := Transaction{tx.TX_id, inputs, outputs} // 复制一份交易
	return txCopy
}

// NewReserveTX，发放准备金：只有输出，没有输入，输出来自于准备金
func NewReserveTX(to []string, data string) Transaction {
	if data == "" { // 如果输入data为0，则生成一串随机数作data
		randData := make([]byte, 20)  // 初始化一个长度为20的字节数组
		_, err := rand.Read(randData) // 取伪随机数
		if err != nil {
			log.Panic(err)
		}
		data = string(randData) // 格式化输出：[]byte转string
	}
	// 创建一个输入项：空
	tx_in := TXInput{[]byte{}, -1, uss.USSToeplitzHashSignMsg{}, data}
	// 创建输出项
	tx_out := make([]TXOutput, 0)
	for _, addr := range to {
		out := NewTXOutput(RESERVE, addr) // 交易金额=RESERVE，接收方地址=to
		tx_out = append(tx_out, out)
	}
	tx := Transaction{nil, []TXInput{tx_in}, tx_out}
	tx.TX_id = tx.SetID()

	return tx
}

// IsReserveTX,检查交易是否是发放准备金
func (tx Transaction) IsReserveTX() bool {
	// 判断依据：1.输入项只有一条；2.引用的交易输出编号为-1；3.引用的交易ID为空
	if len(tx.TX_vin) == 1 && len(tx.TX_vin[0].Refer_tx_id) == 0 && tx.TX_vin[0].Refer_tx_id_index == -1 {
		return true
	}
	return false
}

func (tx Transaction) PrintTransaction() {
	//fmt.Printf("\tID:%x\n", tx.ID)
	for _, vin := range tx.TX_vin {
		//fmt.Printf("\tVin:%d\n", i+1)
		//fmt.Printf("\t--Txid:%x\n", vin.Txid)
		//fmt.Printf("\t  VoutIndex:%d", vin.Vout)
		//fmt.Printf("\t--Sign:%x\n", vin.Signature.Sign)
		fmt.Printf("\tFrom:%s\n", vin.TX_src)
	}
	for _, vout := range tx.TX_vout {
		//fmt.Printf("\tVout:%d\n", j+1)
		fmt.Printf("\tValue:%d\n", vout.TX_value)
		fmt.Printf("\tTo:%s\n", vout.TX_dst)
	}
	fmt.Printf("\n")
}
