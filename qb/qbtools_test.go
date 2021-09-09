package main

import (
	"fmt"
	"qb/qbtools"
	"testing"
)

func TestInitConfig(t *testing.T) {
	fmt.Println("----------【qbtools】----------------------------------------------------------------------")
	client_local_host_map := qbtools.InitConfig_localhost("./qbtools/config/client_localhost.txt")
	for Ci, local_host := range client_local_host_map {
		fmt.Print(string(Ci[:]))
		fmt.Print("=", local_host)
		fmt.Print("\n")
	}
}
