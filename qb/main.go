package main

import (
	"os"
	"qb/qbcli"
)

func main() {
	os.Setenv("NODE_NAME", "C1")
	cli := qbcli.CLI{}
	cli.Run()
}
