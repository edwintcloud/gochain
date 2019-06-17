package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// Transaction represents a blockchain transaction.
type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

// TxInput represents an input transaction.
type TxInput struct {
	ID  []byte
	Out int
	Sig string
}

// TxOutput represents an output transaction.
type TxOutput struct {
	Value  int
	PubKey string
}

// SetID generates a hash id for a transaction.
func (tx *Transaction) SetID() {
	var buffer bytes.Buffer
	var hash [32]byte

	// create encoder on res bytes buffer
	encoder := gob.NewEncoder(&buffer)

	// use encoder to encode transaction into byte slice
	err := encoder.Encode(tx)
	if err != nil {
		log.Panicf("Unable to encode block structure into byte slice: %s", err.Error())
	}

	// convert byte slice from buffer into a hash
	hash = sha256.Sum256(buffer.Bytes())

	// update transaction with generated hash
	tx.ID = hash[:]
}

// CoinbaseTx is a transfer for rewarding an account for mining a block.
func CoinbaseTx(to, data string) *Transaction {

	// ensure data string is not empty
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	// create transaction structures
	txIn := TxInput{
		ID:  []byte{},
		Out: -1,
		Sig: data,
	}
	txOut := TxOutput{
		Value:  100,
		PubKey: to,
	}
	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}

	// generate hash id for transaction
	tx.SetID()

	// return a reference to transaction
	return &tx
}

// NewTransaction initiates a new blockchain transaction.
func (bc *BlockChain) NewTransaction(from, to string, amount int) *Transaction {
	var txInputs []TxInput
	var txOutputs []TxOutput

	// find spendable outputs for address and amount
	acc, spendableOutputs := bc.FindSpendableOutputs(from, amount)

	// quit program if not enough funds to cover amount
	if acc < amount {
		log.Panic("Error: not enough funds to complete transaction")
	}

	// iterate over spendable outputs
	for id, outs := range spendableOutputs {
		txID, err := hex.DecodeString(id)
		if err != nil {
			log.Panicf("Unable to decode id  %v to string: %s", id, err.Error())
		}

		// iterate over current spendable outputs slice of out id's
		for _, out := range outs {
			// add a TxInput to txInputs for from address
			txInputs = append(txInputs, TxInput{
				ID:  txID,
				Out: out,
				Sig: from,
			})
		}
	}

	// add a TxOutput to txOutputs for to address
	txOutputs = append(txOutputs, TxOutput{
		Value:  acc - amount,
		PubKey: to,
	})

	// credit excess back to sender
	if acc > amount {
		txOutputs = append(txOutputs, TxOutput{
			Value:  acc - amount,
			PubKey: from,
		})
	}

	// create transaction with txInputs and txOutputs
	tx := Transaction{
		ID:      nil,
		Inputs:  txInputs,
		Outputs: txOutputs,
	}

	// generate hash id for transaction
	tx.SetID()

	// return a reference to the transaction
	return &tx

}

// IsCoinbase verifies if transaction is a Coinbase transaction.
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 &&
		tx.Inputs[0].Out == -1
}

// CanUnlock verifies that data can be unlocked by the output
// that is referenced inside the TxInput.
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// CanBeUnlocked verifies that account (data) owns the
// information inside the TxOutput.
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
