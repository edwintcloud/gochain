package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Block represents a block in the blockchain.
type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

// CreateBlock creates a new block with a hash and returns a referrence
// to the created block.
func CreateBlock(data string, prevHash []byte) *Block {

	// create new block from data and prev block hash
	block := Block{
		Hash:     []byte{},
		Data:     []byte(data),
		PrevHash: prevHash,
		Nonce:    0,
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
