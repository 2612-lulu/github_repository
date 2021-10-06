package utils

import (
	"fmt"
	"testing"
)

func TestInitConfig(t *testing.T) {
	fmt.Println("----------【qbtools】----------------------------------------------------------------------")

	client_local_host_map := InitConfig(INIT_PATH + "client_localhost.txt")
	for Ci, local_host := range client_local_host_map {
		fmt.Print(string(Ci[:]))
		fmt.Print("=", local_host)
		fmt.Print("\n")
	}
	fmt.Print("\n\n")

	node_local_host_map := InitConfig(INIT_PATH + "node_localhost.txt")
	for Ci, local_host := range node_local_host_map {
		fmt.Print(string(Ci[:]))
		fmt.Print("=", local_host)
		fmt.Print("\n")
	}
	fmt.Print("\n\n")

	id := InitConfig(INIT_PATH + "id_table.txt")
	for Ci, local_host := range id {
		fmt.Print(string(Ci[:]))
		fmt.Print("=", local_host)
		fmt.Print("\n")
	}
}
