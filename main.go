package main

import (
	"flag"
	"fmt"
	"github.com/edwintcloud/gochain/blockchain"
	_ "github.com/joho/godotenv/autoload" // load .env
	"log"
	"os"
	"runtime"
	"strconv"
)

// CLI is a command line interface structure.
type CLI struct {
	bc *blockchain.BlockChain
}

// Initialize function which runs before main
func init() {

	// ensure DB_PATH is created
	os.MkdirAll(os.Getenv("DB_PATH"), os.ModePerm)
}

// MAIN FUNCTION
func main() {

	// defer os exit as last deferment to ensure db is properly closed
	// (probably unneccesary)
	defer os.Exit(0)

	// create new blockchain
	bc := blockchain.InitBlockChain()

	// defer db to close once main function exits
	defer bc.DB.Close()

	// create new cli and run CLI
	cli := CLI{bc}
	cli.run()
}

// run runs command line interface.
func (cli *CLI) run() {

	// validate command line arguments are entered
	// or print instructions and exit gracefully
	if len(os.Args) < 2 {

		// no command was entered, print usage and
		// safely shutdown go routines (so we don't
		// corrupt our database)
		cli.printUsage()
		runtime.Goexit()
	}

	// initialize command line flags
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printBlocksCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	// parse first command line argument
	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("Unable to parse add command: %s", err.Error())
		}
	case "print":
		err := printBlocksCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("Unable to parse print command: %s", err.Error())
		} else {
			cli.printBlocks()
		}
	default:
		// print usage instructions and exit gracefully
		cli.printUsage()
		runtime.Goexit()
	}

	// continue parsing addBlockCmd
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			// print usage instructions and exit gracefully
			cli.printUsage()
			runtime.Goexit()
		}
		// add block to blockchain
		cli.addBlock(*addBlockData)
	}
}

// printUsage prints usage instructions for the cli.
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Printf(" add -block BLOCK_DATA\t Adds a block to the blockchain.\n")
	fmt.Printf(" print\t Prints current blocks in the blockchain.\n")
}

// addBlock adds a block to the cli blockchain
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Added block to blockchain!")
}

// printBlocks iterates over each block in the blockchain,
// printing them out one-by-one
func (cli *CLI) printBlocks() {
	iter := cli.bc.NewIterator()

	// iterate over blocks
	for {
		block := iter.Next()

		fmt.Printf("\nPrevious Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		// break once PrevHash is empty (Genesis block has been reached)
		if len(block.PrevHash) == 0 {
			break
		}
	}
}
