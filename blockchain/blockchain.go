package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/badger"
)

// BlockChain is the representation of our blockchain.
type BlockChain struct {
	PrevHash []byte
	DB       *badger.DB
}

// Iterator is a structure used to iterate over
// blocks in the database.
type Iterator struct {
	CurrentHash []byte
	DB          *badger.DB
}

// InitBlockChain initializes a new BlockChain with an initial Genesis block or
// if the blockchain already exists, loads the prevHash.
func InitBlockChain(address string) *BlockChain {
	var prevHash []byte
	dbPath := os.Getenv("DB_PATH")

	// configure badgerDB
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	// open database
	db, err := badger.Open(opts)
	if err != nil {
		log.Panicf("Unable to open database at path %s: %s", dbPath, err.Error())
	}

	// initiate update on the database by passing in closure
	// update allows read and write (view allows read only)
	err = db.Update(func(txn *badger.Txn) error {

		// check if blockchain in database
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {

			// blockchain was not found in db
			fmt.Println("No existing blockchain found in database.")

			// create Coinbase transaction with address
			cbTx := CoinbaseTx(address, "Genesis Block")

			// create Genesis block
			genesis := CreateBlock([]*Transaction{cbTx}, []byte{})
			fmt.Println("Genesis block created")

			// put genesis in db with the hash as key
			// and byte slice of block as value
			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				// return from closure with error
				return errors.New("unable to set genesis hash - " + err.Error())
			}

			// put genesis in db as previous hash (Hash is a byte slice)
			// and set prevHash
			err = txn.Set([]byte("lh"), genesis.Hash)
			prevHash = genesis.Hash

			// return from closure
			return err
		}

		// blockchain was found in db
		fmt.Println("Blockchain found in database.")

		// get previous hash item from db
		prevHashItem, err := txn.Get([]byte("lh"))
		if err != nil {
			// return from closure with error
			return errors.New("unable to get previous hash item - " + err.Error())
		}

		// set prevHash to value of prevHashItem
		prevHash, err = prevHashItem.Value()

		// return from closure
		return err
	})
	if err != nil {
		log.Panicf("Unable to update database: %s", err.Error())
	}

	// create blockchain with db reference and prevHash from db
	// and return it's reference
	return &BlockChain{
		PrevHash: prevHash,
		DB:       db,
	}
}

// AddBlock adds a block to the receiver BlockChain.
func (bc *BlockChain) AddBlock(transactions []*Transaction) {
	var prevHash []byte

	// initiate read-only transaction on db to get previous hash from db
	err := bc.DB.View(func(txn *badger.Txn) error {

		// get previous hash item from db
		prevHashItem, err := txn.Get([]byte("lh"))
		if err != nil {
			// return from closure with error
			return errors.New("unable to get previous hash item - " + err.Error())
		}

		// set prevHash to value of prevHashItem
		prevHash, err = prevHashItem.Value()

		// return from closure
		return err
	})
	if err != nil {
		log.Panicf("Unable to read previous hash from database: %s", err.Error())
	}

	// create new block with previous hash and data
	newBlock := CreateBlock(transactions, prevHash)

	// initiate rw transaction on db to insert newBlock
	err = bc.DB.Update(func(txn *badger.Txn) error {

		// put newBlock in db with the hash as key
		// and byte slice of block as value
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			// return from closure with error
			return errors.New("unable to set newBlock hash - " + err.Error())
		}

		// put newBlock in db as previous hash (Hash is a byte slice)
		// and set blockchain PrevHash
		err = txn.Set([]byte("lh"), newBlock.Hash)
		bc.PrevHash = newBlock.Hash

		// return from closure
		return err
	})
	if err != nil {
		log.Panicf("Unable to update database with new block: %s", err.Error())
	}
}

// NewIterator initializes and returns a reference to a
// new blockchain Iterator from a BlockChain.
func (bc *BlockChain) NewIterator() *Iterator {
	return &Iterator{bc.PrevHash, bc.DB}
}

// Next returns the next Block in a blockchain Iterator (order is reversed).
func (iter *Iterator) Next() *Block {
	var block *Block

	// initiate read only transaction on db to get next block
	err := iter.DB.View(func(txn *badger.Txn) error {

		// get current hash item from db
		nextItem, err := txn.Get(iter.CurrentHash)
		if err != nil {
			// return from closure with error
			return errors.New("unable to get next hash item - " + err.Error())
		}

		// get byte slice value from nextItem
		encodedBlock, err := nextItem.Value()
		if err != nil {
			// return from closure with error
			return errors.New("unable to get value from nextItem - " + err.Error())
		}

		// deserialize encodedBlock into a new Block
		block = Deserialize(encodedBlock)

		// return from closure
		return err
	})
	if err != nil {
		log.Panicf("Unable to get next block from database: %s", err.Error())
	}

	// update iterator to move to next block
	iter.CurrentHash = block.PrevHash

	// return reference to new block
	return block
}

// FindTransaction finds a transaction in the Blockchain by ID.
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.NewIterator()

	// iterate over blocks
	for {
		block := iter.Next()

		// iterate through transactions for current block
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	// return empty transaction and error if transaction was not found
	return Transaction{}, errors.New("transaction does not exist")
}

// SignTransaction signs a blockchain Transaction.
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	// iterate over TxInputs in Transaction and populate
	// prevTXs
	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			log.Panicln("Unable to sign blockchain transaction: ", err.Error())
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	// sign Transaction using Transaction method
	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies a Transaction on the blockchain.
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	// iterate over TxInputs in Transaction and populate
	// prevTXs
	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			log.Panicln("Unable to verify blockchain transaction: ", err.Error())
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	// verify Transaction using Transaction method
	// and return result
	return tx.Verify(prevTXs)
}

// FindUnspentTransactions determines how many tokens an address has by
// finding transactions that have outputs which are not referenced
// by other inputs.
func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
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
				if out.IsLockedWithKey(pubKeyHash) {
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
					if in.UsesKey(pubKeyHash) {
						// if address can unlock the output referenced
						// by the input, add the tx to spentTXOutputs
						inTxID := hex.EncodeToString(in.ID)
						spentTxOutputs[inTxID] = append(spentTxOutputs[inTxID], in.Out)
					}
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
func (bc *BlockChain) FindUnspentTxOutputs(pubKeyHash []byte) []TxOutput {
	var unspentTxOutputs []TxOutput

	// get unspent transactions
	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)

	// iterate over unspent transactions
	for _, tx := range unspentTxs {
		// iterate over outputs for current tx
		for _, out := range tx.Outputs {
			// if the output can be unlocked by the address,
			// add it to unspentTxOutputs
			if out.IsLockedWithKey(pubKeyHash) {
				unspentTxOutputs = append(unspentTxOutputs, out)
			}
		}
	}

	// return unspent transaction outputs
	return unspentTxOutputs
}

// FindSpendableOutputs ensures enough tokens exists in unspent transaction
// outputs to cover the amount.
func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	spendableOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)
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
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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
