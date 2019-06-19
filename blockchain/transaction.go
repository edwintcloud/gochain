package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/edwintcloud/gochain/wallet"
)

// Transaction represents a blockchain transaction.
type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

// Serialize serializes a Transaction into bytes.
func (tx *Transaction) Serialize() []byte {
	var buffer bytes.Buffer

	// create encoder on res bytes buffer
	encoder := gob.NewEncoder(&buffer)

	// use encoder to encode Transaction into byte slice
	err := encoder.Encode(tx)
	if err != nil {
		log.Panicf("Unable to encode Transaction structure into byte slice: %s", err.Error())
	}

	// return bytes from buffer
	return buffer.Bytes()
}

// GenerateHash generates a sha256 hash from the bytes of a Transaction
// structure. It is important we do not use a pointer receiver here so
// that the original Transaction is not modified.
func (tx Transaction) GenerateHash() []byte {
	var hash [32]byte

	// clear the ID field
	tx.ID = []byte{}

	// generate hash
	hash = sha256.Sum256(tx.Serialize())

	// return hash
	return hash[:]
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
		ID:        []byte{},
		Out:       -1,
		Signature: nil,
		PubKey:    []byte(data),
	}
	txOut := NewTXOutput(
		100,
		to,
	)
	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{*txOut},
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

	// create wallets and generate public key for from addressed wallet
	wallets, err := wallet.CreateWallets()
	if err != nil {
		log.Panicln("Unable to load wallets while creating new blockchain transaction: ", err.Error())
	}
	w := wallets[from]
	pubKeyHash := wallet.GeneratePublicKeyHash(w.PublicKey)

	// find spendable outputs for address and amount
	acc, spendableOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

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
				ID:        txID,
				Out:       out,
				Signature: nil,
				PubKey:    w.PublicKey,
			})
		}
	}

	// add a TxOutput to txOutputs for to address
	txOutputs = append(txOutputs, *NewTXOutput(
		amount,
		to,
	))

	// credit excess back to sender
	if acc > amount {
		txOutputs = append(txOutputs, *NewTXOutput(
			acc-amount,
			to,
		))
	}

	// create transaction with txInputs and txOutputs
	tx := Transaction{
		ID:      nil,
		Inputs:  txInputs,
		Outputs: txOutputs,
	}

	// generate hash and sign transaction
	tx.ID = tx.GenerateHash()
	bc.SignTransaction(&tx, w.PrivateKey)

	// return a reference to the transaction
	return &tx

}

// IsCoinbase verifies if transaction is a Coinbase transaction.
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 &&
		tx.Inputs[0].Out == -1
}

// Sign signs a Transaction.
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	// verify Transaction is not a Coinbase Transaction
	if tx.IsCoinbase() {
		return
	}

	// iterate over Transaction Inputs
	for _, in := range tx.Inputs {
		// verify that the transaction referenced in prevTXs
		// by the current TxInput does not have a nil ID
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panicln("Unable to sign Transaction: the previous transaction does not exist")
		}
	}

	// create a trimmed copy of the Transaction so we don't modify
	// the original while signing
	txCopy := tx.TrimmedCopy()

	// iterate over txCopy inputs
	for inID, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTX.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.GenerateHash()
		txCopy.Inputs[inID].PubKey = nil

		// sign ID using privKey
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panicln("Unable to sign Transaction: ", err.Error())
		}

		// add signature (concatenaton of signing outputs) to original Transaction input
		tx.Inputs[inID].Signature = append(r.Bytes(), s.Bytes()...)

	}

}

// Verify verifies a Transaction.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {

	// return true for a Coinbase Transaction
	if tx.IsCoinbase() {
		return true
	}

	// iterate over Transaction Inputs
	for _, in := range tx.Inputs {
		// verify that the transaction referenced in prevTXs
		// by the current TxInput does not have a nil ID
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panicln("Unable to verify Transaction: the previous transaction does not exist")
		}
	}

	// create a trimmed copy of the Transaction so we don't modify
	// the original while signing
	txCopy := tx.TrimmedCopy()

	// define the curve for checking the signature of each input
	curve := elliptic.P256()

	// iterate over txCopy inputs
	for inID, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTX.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.GenerateHash()
		txCopy.Inputs[inID].PubKey = nil

		// unpack r and s from signature
		r := big.Int{}
		s := big.Int{}
		sigMedian := len(in.Signature) / 2
		r.SetBytes(in.Signature[:sigMedian])
		s.SetBytes(in.Signature[sigMedian:])

		// unpack x and y from public key
		x := big.Int{}
		y := big.Int{}
		keyMedian := len(in.PubKey) / 2
		x.SetBytes(in.PubKey[:keyMedian])
		y.SetBytes(in.PubKey[keyMedian:])

		// create ecdsa public key using curve, x, and y
		pubKey := ecdsa.PublicKey{curve, &x, &y}

		// verify the private key with the public key
		if !ecdsa.Verify(&pubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	// return true if all inputs were verified
	return true
}

// TrimmedCopy makes a deep copy of a Transaction excluding the signature and
// public key for each TxInput.
func (tx *Transaction) TrimmedCopy() Transaction {
	newTx := Transaction{}

	for _, in := range tx.Inputs {
		newTx.Inputs = append(newTx.Inputs, TxInput{
			ID:        in.ID,
			Out:       in.Out,
			Signature: nil,
			PubKey:    nil,
		})
	}

	copy(newTx.Outputs, tx.Outputs)

	return newTx
}

// ToString returns a string representation of a Transaction.
func (tx *Transaction) String() string {
	result := []string{
		fmt.Sprintf("--- Transaction %x:", tx.ID),
	}

	// iterate over inputs
	for inID, in := range tx.Inputs {
		result = append(result,
			fmt.Sprintf("\tInput %d:", inID),
			fmt.Sprintf("\t\tTXID:\t%x", in.ID),
			fmt.Sprintf("\t\tOut:\t%d", in.Out),
			fmt.Sprintf("\t\tSignature:\t%x", in.Signature),
			fmt.Sprintf("\t\tPubKey:\t%x", in.PubKey),
		)
	}

	// iterate over outputs
	for outID, out := range tx.Outputs {
		result = append(result,
			fmt.Sprintf("\tOutput %d:", outID),
			fmt.Sprintf("\t\tValue:\t%d", out.Value),
			fmt.Sprintf("\t\tScript:\t%x", out.PubKeyHash),
		)
	}

	// return the result as a joined string
	return strings.Join(result, "\n")
}
