package main

import (
	"os"

	"github.com/edwintcloud/gochain/cli"
	_ "github.com/joho/godotenv/autoload" // load .env
)

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
	cli := cli.CLI{}
	cli.Run()

	// w := wallet.CreateWallet()

	// fmt.Printf("pub key: %x\n", w.PublicKey)
	// fmt.Printf("pub hash: %x\n", wallet.GeneratePublicKeyHash(w.PublicKey))
	// fmt.Printf("address: %s\n", w.Address())
}
