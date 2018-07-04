package main

import (
	"github.com/orbs-network/orbs-network-go/bootstrap"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	nodeId := os.Getenv("NODE_ID")

	bootstrap.NewNode(":" +port, nodeId, true, 1)
}
