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

	// create new cli and run CLI
	cli := CLI{}
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
	getBalanceCmd := flag.NewFlagSet("getbal", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("create", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printBlocksCmd := flag.NewFlagSet("print", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	// parse first command line argument
	switch os.Args[1] {
	case "print":
		err := printBlocksCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("Unable to parse print command: %s", err.Error())
		} else {
			cli.printBlocks()
		}
	case "getbal":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("Unable to parse %s command: %s", os.Args[1], err.Error())
		}
	case "create":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("Unable to parse %s command: %s", os.Args[1], err.Error())
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("Unable to parse %s command: %s", os.Args[1], err.Error())
		}
	default:
		// print usage instructions and exit gracefully
		cli.printUsage()
		runtime.Goexit()
	}

	// continue parsing getBalanceCmd
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	// continue parsing createBlockchainCmd
	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	// continue parsing sendCmd
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func (cli *CLI) createBlockChain(address string) {
	bc := blockchain.InitBlockChain(address)
	bc.DB.Close()
	fmt.Println("Finished!")
}

func (cli *CLI) getBalance(address string) {
	bc := blockchain.InitBlockChain(address)
	defer bc.DB.Close()

	balance := 0
	unspentTxOutputs := bc.FindUnspentTxOutputs(address)

	for _, out := range unspentTxOutputs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := blockchain.InitBlockChain(from)
	defer bc.DB.Close()

	tx := bc.NewTransaction(from, to, amount)
	bc.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success!")
}

// printUsage prints usage instructions for the cli.
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Printf(" getbal -address ADDRESS\t Gets the balance for an address.\n")
	fmt.Printf(" create -address ADDRESS\t Creates a blockchain and sends genesis reward to address.\n")
	fmt.Printf(" print\t Prints the blocks in the chain.\n")
	fmt.Printf(" send -from FROM -to TO -amount AMOUNT\t Sends amount of coins from one address to another.\n")
}

// printBlocks iterates over each block in the blockchain,
// printing them out one-by-one
func (cli *CLI) printBlocks() {
	bc := blockchain.InitBlockChain("")
	defer bc.DB.Close()
	iter := bc.NewIterator()

	// iterate over blocks
	for {
		block := iter.Next()

		fmt.Printf("\nPrevious Hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		// break once PrevHash is empty (Genesis block has been reached)
		if len(block.PrevHash) == 0 {
			break
		}
	}
}
