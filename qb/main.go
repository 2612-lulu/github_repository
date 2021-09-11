package main

import (
	"os"
	"qb/network"
)

func main() {
	nodeID := os.Args[1]
	network.CreateServer(nodeID)
}
