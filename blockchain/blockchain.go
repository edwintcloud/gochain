package blockchain

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
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

// InitBlockChain initializes a new BlockChain with an initial Genesis block.
func InitBlockChain() *BlockChain {
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

			// create Genesis block
			genesis := CreateBlock("Genesis", []byte{})
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
func (bc *BlockChain) AddBlock(data string) {
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
	newBlock := CreateBlock(data, prevHash)

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
