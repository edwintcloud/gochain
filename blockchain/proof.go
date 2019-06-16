package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

// Difficulty is the mining difficulty.
const Difficulty = 18

// ProofOfWork represents a proof of work.
type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// NewProof creates a new proof of work and returns a
// reference to the new proof of work.
func NewProof(b *Block) *ProofOfWork {

	// cast 1 to big int
	target := big.NewInt(1)

	// left shift bytes in target by 256 - difficulty
	// target << 256 - Difficulty
	target.Lsh(target, uint(256-Difficulty))

	// return new proof of work
	return &ProofOfWork{b, target}
}

// InitData initializes a proof of work with provided
// nonce.
func (pow *ProofOfWork) InitData(nonce int) []byte {

	// create new byte slice from prev hash, data, nonce, and difficulty
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.Data,
			ToBytes(int64(nonce)),
			ToBytes(int64(Difficulty)),
		}, []byte{})

	// return byte slice
	return data
}

// Run executes a proof of work.
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte
	nonce := 0

	for nonce < math.MaxInt64 {
		// get a byte slice proof of work with nonce
		data := pow.InitData(nonce)

		// hash the proof of work data
		hash = sha256.Sum256(data)

		// print current hash
		fmt.Printf("\r%x", hash)

		// convert hash into big int
		intHash.SetBytes(hash[:])

		// compare proof of work target and intHash
		if intHash.Cmp(pow.Target) == -1 {
			// block has been signed, break
			break
		} else {
			// increment nonce
			nonce++
		}
	}

	// print some space
	fmt.Println()

	// return nonce and hash
	return nonce, hash[:]
}

// Validate verifies that a completed proof of work is valid.
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	// get a byte slice proof of work with nonce
	data := pow.InitData(pow.Block.Nonce)

	// hash the proof of work data
	hash := sha256.Sum256(data)

	// convert hash into big int
	intHash.SetBytes(hash[:])

	// compare proof of work target and intHash
	// return true if match
	return intHash.Cmp(pow.Target) == -1
}

// ToBytes decodes a int64 into bytes.
func ToBytes(num int64) []byte {

	// create bytes buffer
	buffer := bytes.Buffer{}

	// decode num into bytes
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		log.Fatalf("Unable to decode %d into bytes: %s", num, err.Error())
	}

	// return bytes from buffer
	return buffer.Bytes()
}
