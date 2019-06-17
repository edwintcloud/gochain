package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
)

// Block represents a block in the blockchain.
type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

// HashTransactions hashes transactions into a byte slice.
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	// add each transaction from block into txHashes
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	// join txHashes together and hash them into txHash
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	// return hash of transactions
	return txHash[:]
}

// CreateBlock creates a new block with a hash and returns a referrence
// to the created block.
func CreateBlock(txs []*Transaction, prevHash []byte) *Block {

	// create new block from data and prev block hash
	block := Block{
		Hash:         []byte{},
		Transactions: txs,
		PrevHash:     prevHash,
		Nonce:        0,
	}

	// create proof of work for block
	pow := NewProof(&block)

	// run proof of work on data
	nonce, hash := pow.Run()

	// update block with hash and nonce
	block.Hash = hash[:]
	block.Nonce = nonce

	// return a reference to the new block
	return &block
}

// Serialize serializes a block into a byte slice so it can be stored in the db.
func (b *Block) Serialize() []byte {
	var buffer bytes.Buffer

	// create encoder on res bytes buffer
	encoder := gob.NewEncoder(&buffer)

	// use encoder to encode block into byte slice
	err := encoder.Encode(b)
	if err != nil {
		log.Panicf("Unable to encode block structure into byte slice: %s", err.Error())
	}

	// return bytes from buffer
	return buffer.Bytes()
}

// Deserialize deserializes a byte slice into a new Block and returns a
// reference to the created Block.
func Deserialize(data []byte) *Block {
	var block Block

	// create decoder on a bytes reader of the data byte slice
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// use decoder to decode bytes reader into created block
	err := decoder.Decode(&block)
	if err != nil {
		log.Panicf("Unable to decode byte slice into a new Block struct: %s", err.Error())
	}

	// return reference to decoded block
	return &block
}

// FindUnspentTransactions determines how many tokens an address has by
// finding transactions that have outputs which are not referenced
// by other inputs.
func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction
	spentTxOutputs := make(map[string][]int)

	iter := bc.NewIterator()

	// iterate over blocks
	for {
		block := iter.Next()

		// iterate over transactions for current block
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs: // a label to continue from
			// iterate over outputs for current Transaction
			for outIdx, out := range tx.Outputs {
				// check if current output inside spentTxOutputs
				if spentTxOutputs[txID] != nil {
					for _, spentOut := range spentTxOutputs[txID] {
						if spentOut == outIdx {
							// current output is spent, continue
							continue Outputs
						}
					}
				}
				// if transaction is unspent and can be unlocked by
				// address, add it to unspentTxs
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			// LABEL END - Outputs

			// if transaction is not coinbase tx, find other
			// transactions that are referenced by inputs
			// that can be unlocked by the address
			if tx.IsCoinbase() == false {
				// iterate over inputs
				for _, in := range tx.Inputs {
					// if address can unlock the output referenced
					// by the input, add the tx to spentTXOutputs
					inTxID := hex.EncodeToString(in.ID)
					spentTxOutputs[inTxID] = append(spentTxOutputs[inTxID], in.Out)
				}
			}
		}

		// break once PrevHash is empty (Genesis block has been reached)
		if len(block.PrevHash) == 0 {
			break
		}
	}

	// return unspent transactions
	return unspentTxs
}

// FindUnspentTxOutputs finds all unspent transaction outputs that
// correspond to an address.
func (bc *BlockChain) FindUnspentTxOutputs(address string) []TxOutput {
	var unspentTxOutputs []TxOutput

	// get unspent transactions
	unspentTxs := bc.FindUnspentTransactions(address)

	// iterate over unspent transactions
	for _, tx := range unspentTxs {
		// iterate over outputs for current tx
		for _, out := range tx.Outputs {
			// if the output can be unlocked by the address,
			// add it to unspentTxOutputs
			if out.CanBeUnlocked(address) {
				unspentTxOutputs = append(unspentTxOutputs, out)
			}
		}
	}

	// return unspent transaction outputs
	return unspentTxOutputs
}

// FindSpendableOutputs ensures enough tokens exists in unspent transaction
// outputs to cover the amount.
func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	spendableOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work: // a label to continue from
	// iterate over unspent transactions
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		// iterate over outputs for current tx
		for outIdx, out := range tx.Outputs {
			// if output can be unlocked by address and accumulated is less
			// than amount, increment accumulated by out value and add
			// tx to spendableOutputs
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				spendableOutputs[txID] = append(spendableOutputs[txID], outIdx)

				// once accumulated reaches or exceeds the amount, we have found
				// enough spendable outputs and can break
				if accumulated >= amount {
					break Work
				}
			}

		}
	}

	// return accumulated amount and spendable outputs
	return accumulated, spendableOutputs
}
