package blockchain

import (
	"bytes"
	"log"
	"os"
	"strconv"

	"github.com/btcsuite/btcutil/base58"
	"github.com/edwintcloud/gochain/wallet"
)

// TxInput represents an input transaction.
type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

// TxOutput represents an output transaction.
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// CreateTxOutput creates a new TxOutput.
func CreateTxOutput(value int, address string) *TxOutput {

	// create new TxOutput
	out := TxOutput{
		Value:      value,
		PubKeyHash: nil,
	}

	// lock TxOutput by populating PubKeyHash
	out.Lock([]byte(address))

	// return reference to new TxOutput
	return &out
}

// UsesKey verifies that a TxInput has a valid public key.
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	return bytes.Compare(wallet.GeneratePublicKeyHash(in.PubKey), pubKeyHash) == 0
}

// Lock locks TxOutput.
func (out *TxOutput) Lock(address []byte) {
	checksumLen, err := strconv.Atoi(os.Getenv("CHECKSUM_LENGTH"))
	if err != nil {
		log.Panicln("Unable to convert env var CHECKSUM_LENGTH to int for method (TxOutput) Lock: ", err.Error())
	}

	// decode address from base58 back to sha256 hash
	pubKeyHash := base58.Decode(string(address[:]))

	// set TxOutput public key hash to decoded hash
	// without the version or checksum
	out.PubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLen]
}

// IsLockedWithKey checks to see if output has public key hash equal to given
// public key hash.
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput creates a new output Transaction.
func NewTXOutput(value int, address string) *TxOutput {
	txOut := &TxOutput{value, nil}
	txOut.Lock([]byte(address))

	return txOut
}
