package main

import (
	"fmt"
	"github.com/edwintcloud/gochain/blockchain"
	"strconv"
)

func main() {
	// create new blockchain
	bc := blockchain.InitBlockChain("Genesis")

	// add a few blocks
	bc.AddBlock("Second")
	bc.AddBlock("Third")
	bc.AddBlock("Fourth")

	// print out block info
	for _, block := range bc.Blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
	}
}
