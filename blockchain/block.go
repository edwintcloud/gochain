package blockchain

// Block represents a block in the blockchain.
type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

// BlockChain is the representation of our blockchain.
type BlockChain struct {
	Blocks []*Block
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

// AddBlock adds a block to the receiver BlockChain.
func (bc *BlockChain) AddBlock(data string) {

	// get previous block
	prevBlock := bc.Blocks[len(bc.Blocks)-1]

	// create new block from data and prev block hash
	newBlock := CreateBlock(data, prevBlock.Hash)

	// append the new block to receiver blockchain
	bc.Blocks = append(bc.Blocks, newBlock)
}

// InitBlockChain initializes a new BlockChain with provided first element.
func InitBlockChain(data string) *BlockChain {
	return &BlockChain{[]*Block{CreateBlock(data, []byte{})}}
}
