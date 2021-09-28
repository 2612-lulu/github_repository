package qbtx

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"qb/qbtools"
	"qb/qkdserv"
	"qb/uss"
)

// 定义准备金金额
const RESERVE = 10

// 定义验签者数量
const N = 4

// Transaction，交易结构，多入多处：
// 有一些输出并没有被关联到某个输入上；一笔交易的输入可以引用之前多笔交易的输出；一个输入必须引用一个输出
type Transaction struct {
	ID   []byte     // 交易ID，非常重要的Hash值，是在input签名之前计算出来的，作为UTXOSet.map的key存在
	Vin  []TXInput  // 交易输入项
	Vout []TXOutput // 交易输出项
}

type ToGenTx struct {
	From  string
	To    string
	Value int
}

// SetID，根据交易输入与输出项生成交易ID。
func (tx *Transaction) SetID() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.SerializeTX())

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
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}
	return transaction
}

// SignTX，对交易输入项签名
func (tx *Transaction) SignTX(nodeName string) {
	txCopy := tx.TrimmedCopyTX() // 复制并修剪交易以得到待签名数据

	for inID, _ := range txCopy.Vin { // 循环向输入项签名
		txCopy.Vin[inID].Signature = uss.USSToeplitzHashSignMsg{} // 置空
		dataToSign := txCopy.Vin[inID].SerializeInput()           // 待签名数据
		signature := uss.USSToeplitzHashSignMsg{
			Sign_index: qkdserv.QKDSignMatrixIndex{
				Sign_dev_id:  qbtools.GetNodeIDTable(nodeName),
				Sign_task_sn: uss.GenSignTaskSN(16),
			},
			Main_row_num: qkdserv.QKDSignRandomMainRowNum{
				Sign_Node_Name: nodeName,
				Main_Row_Num:   0, // 签名主行号，签名时默认为0
			},
			Sign_counts: N,
			Sign_len:    16,
			Message:     dataToSign,
		}
		signature = uss.Sign(signature.Sign_index, signature.Sign_counts, signature.Sign_len, signature.Message)
		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].From = "" // 将From置空
	}
}

// VerifyTX,交易输入项验签
func (tx *Transaction) VerifyTX() bool {
	txCopy := tx

	for inID, _ := range tx.Vin {
		if uss.VerifySign(txCopy.Vin[inID].Signature) {
			log.Println("verify of tx success")
		} else {
			log.Println("verify of tx wrong")
		}
	}

	return true
}

// TrimmedCopyTX，交易修剪以得到待签名消息
func (tx *Transaction) TrimmedCopyTX() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin { // 将原交易内的签名和公钥都置空
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, uss.USSToeplitzHashSignMsg{}, ""})
	}

	for _, vout := range tx.Vout { // 复制原输入项
		outputs = append(outputs, TXOutput{vout.Value, vout.To})
	}

	txCopy := Transaction{tx.ID, inputs, outputs} // 复制一份交易
	return txCopy
}

// NewReserveTX，发放准备金：只有输出，没有输入，输出来自于准备金
func NewReserveTX(to []string, data string) *Transaction {
	if data == "" { // 如果输入data为0，则生成一串随机数作data
		randData := make([]byte, 20)  // 初始化一个长度为20的字节数组
		_, err := rand.Read(randData) // 取伪随机数
		if err != nil {
			log.Panic(err)
		}
		data = string(randData) // 格式化输出：[]byte转string
	}
	// 创建一个输入项：空
	txin := TXInput{[]byte{}, -1, uss.USSToeplitzHashSignMsg{}, data}
	// 创建输出项
	txout := make([]TXOutput, 0)
	for _, addr := range to {
		out := NewTXOutput(RESERVE, addr) // 交易金额=RESERVE，接收方地址=to
		txout = append(txout, *out)
	}
	tx := Transaction{nil, []TXInput{txin}, txout}
	tx.ID = tx.SetID()
	/*if tx.IsReserveTX() {
		log.Println("create a new reserve tx")
	}*/
	return &tx
}

// IsReserveTX,检查交易是否是发放准备金
func (tx Transaction) IsReserveTX() bool {
	// 判断依据：1.输入项只有一条；2.引用的交易输出编号为-1；3.引用的交易ID为空
	if len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1 {
		return true
	}
	return false
}

func (tx Transaction) PrintTransaction() {
	fmt.Printf("ID:%x\n", tx.ID)

	for i, vin := range tx.Vin {
		fmt.Printf("Vin:%d\n", i)
		fmt.Printf("\tTxid.:%x\n", vin.Txid)
		fmt.Printf("\tVoutIndex.:%d\n", vin.Vout)
		fmt.Printf("\tSign:%x\n", vin.Signature)
		fmt.Printf("\tFrom:%s\n", vin.From)
	}
	for j, vout := range tx.Vout {
		fmt.Printf("Vout:%d\n", j)
		fmt.Printf("\tValue:%d\n", vout.Value)
		fmt.Printf("\tTo:%s\n", vout.To)
	}
	fmt.Printf("\n")
}
