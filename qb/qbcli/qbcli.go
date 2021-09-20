package qbcli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"qb/qbnode"
)

// CLI responsible for processing command line arguments
type CLI struct{}

var client *Client
var node *qbnode.Node

// 命令行帮助函数
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  transaction -from FROM -to TO -amount AMOUNT -Send AMOUNT of coins from FROM address to TO.")
}

func (cli *CLI) validateArgs() {
	// 判断参数准确与否
	if len(os.Args) < 1 {
		cli.printUsage()
		os.Exit(1)
	}
}

// 命令行代码封装调用
func (cli *CLI) Run() {
	cli.validateArgs()

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		fmt.Printf("NODE_Name env var is not set!")
		os.Exit(1)
	}
	CreateServer(nodeName) // 创建节点

	// 1.利用NewFlagSet函数立flag。
	// name参数的种类："getbalance"，对应命令行参数os.Args[1]，代表要做什么事情
	// errorHandling错误的处理方式：继续ContineOnError，退出ExitOnError，抛出恐慌PanicOnError
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError) // 查询余额
	txCmd := flag.NewFlagSet("transaction", flag.ExitOnError)        // 交易

	// 2.设定参数接收变量，如果有多个参数值要获取，需要设置多个变量
	// name参数名称：如"address"
	// value默认值：如 ""，0
	// usage对应的元素：如"The address to get balance for"
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	txFrom := txCmd.String("from", "", "Source wallet address")
	txTo := txCmd.String("to", "", "Destination wallet address")
	txAmount := txCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	// 3.利用FlagSet解析命令行参数，解析是从os.Args[2]开始
	case "getbalance": // 查询账户余额
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "transaction": // 发起交易
		err := txCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	// 4.确认FlagSet参数解析。
	if getBalanceCmd.Parsed() { // 确认flag参数getBalanceCmd出现
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		//cli.getBalance(*getBalanceAddress, nodeID)
	}
	if txCmd.Parsed() {
		if *txFrom == "" || *txTo == "" || *txAmount <= 0 {
			txCmd.Usage()
			os.Exit(1)
		}
		cli.sendTX(*txFrom, *txTo, nodeName, *txAmount)
	}

}

// 节点初始化
func CreateServer(Name string) {
	node_type := Name[:1]
	switch node_type {
	case "P": // 如果是联盟节点
		node = qbnode.NewNode(Name)
	case "C":
		client = NewClient(Name) // 启动client节点
		//client.Httplisten()      // 开启http
	default:
		fmt.Println("NODE_Name env var is not correct")
	}
}
