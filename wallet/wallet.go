package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
	"github.com/btcsuite/btcutil/base58"
)

const (
	checksumLen = 4
	version     = byte(0x00)
)

// Wallet represents a token wallet for an address.
type Wallet struct {
	// eliptical curve digital signing algorithm private key
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// CreateWallet creates a new Wallet.
func CreateWallet() *Wallet {

	// generate a new key pair
	privKey, pubKey := GenerateKeyPair()

	// return new wallet
	return &Wallet{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
}

// Address returns the generated Wallet address which is a base58 formed
// from the public key hash, version, and checksum.
func (w *Wallet) Address() []byte {

	// generate public key hash
	pubHash := GeneratePublicKeyHash(w.PublicKey)

	// concatenate the version to the begining of pubHash
	vHash := append([]byte{version}, pubHash...)

	// concatenate the checksum to the end of vHash
	finalHash := append(vHash, GenerateChecksum(vHash)...)

	// return the byte slice representation of the base58
	// encoding of finalHash
	return []byte(base58.Encode(finalHash))
}

// GenerateKeyPair generates a new ecdsa private and public key pair.
// As a note, this algorithm can generate 10^77 unique keys which is
// more than the number of known atoms in the universe O_O
func GenerateKeyPair() (ecdsa.PrivateKey, []byte) {

	// define curve type as p256 (outputs will be 256 bytes)
	curve := elliptic.P256()

	// generate key using curve and random number generator
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panicln("Unable to generate ecdsa key pair: ", err.Error())
	}

	// concatenate ecdsa pubKey x and y to make a public key
	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)

	// return key pair
	return *privKey, pubKey
}

// GeneratePublicKeyHash generates a hash for a public key using sha256 and ripemd160.
func GeneratePublicKeyHash(pubKey []byte) []byte {

	// hash using sha256
	pubHash := sha256.Sum256(pubKey)

	// write pubHash into a ripemd160 hash
	rmdHash := ripemd160.New()
	_, err := rmdHash.Write(pubHash[:])
	if err != nil {
		log.Panicln("Unable to write pubHash into ripemd160 hash: ", err.Error())
	}

	// generate and return final hash
	return rmdHash.Sum(nil)
}

// GenerateChecksum generates a checksum for a public key hash.
func GenerateChecksum(payload []byte) []byte {

	// generate a sha256 hash from payload
	hash := sha256.Sum256(payload)

	// generate a sha256 hash from hash
	rehash := sha256.Sum256(hash[:])

	// return checksum of checksumLen bytes
	return rehash[:checksumLen]
}
