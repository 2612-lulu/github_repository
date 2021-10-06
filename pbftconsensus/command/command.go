package command

import (
	"flag"
	"fmt"
	"log"
	"os"
	"pbftconsensus/network"
	"qkdserv"
)

// CLI responsible for processing command line arguments
type COMM struct{}

// 命令行帮助函数
func (command *COMM) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  startPBFT -Start a node with ID specified in NODE_ID env.") // 开启联盟节点
}

func (command *COMM) validateArgs() {
	// 判断参数准确与否
	if len(os.Args) < 1 {
		command.printUsage()
		os.Exit(1)
	}
}

// 命令行代码封装调用
func (command *COMM) Run() {
	command.validateArgs()

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		fmt.Printf("NODE_Name env var is not set!")
		os.Exit(1)
	}
	qkdserv.Node_name = nodeName // 调用此程序的当前节点或客户端名称
	// 初始化签名密钥池
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)

	// 1.利用NewFlagSet函数立flag。
	// name参数的种类："getbalance"，对应命令行参数os.Args[1]，代表要做什么事情
	// errorHandling错误的处理方式：继续ContineOnError，退出ExitOnError，抛出恐慌PanicOnError
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError) // 创建节点

	// 2.设定参数接收变量，如果有多个参数值要获取，需要设置多个变量

	switch os.Args[1] {
	// 3.利用FlagSet解析命令行参数，解析是从os.Args[2]开始
	case "startPBFT":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		command.printUsage()
		os.Exit(1)
	}

	// 4.确认FlagSet参数解析。
	if startNodeCmd.Parsed() {
		network.NewNodeConsensus(nodeName)
	}

}
