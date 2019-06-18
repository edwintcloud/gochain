package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

// CreateWallets makes a map of wallets and populates it with
// data from the wallets file if it exists.
func CreateWallets() (map[string]*Wallet, error) {
	wallets := make(map[string]*Wallet)

	// try to load wallets from file
	err := LoadWalletsFile(&wallets)

	// return wallets and err
	return wallets, err
}

// LoadWalletsFile loads wallets from a file into a map.
func LoadWalletsFile(wallets *map[string]*Wallet) error {

	// try to read file or return error
	fileBytes, err := ioutil.ReadFile(os.Getenv("WALLETS_FILE"))
	if err != nil {
		return err
	}

	// register gob encoder to read file format and create a
	// new decoder
	gob.Register(elliptic.P256())
	gobDecoder := gob.NewDecoder(bytes.NewReader(fileBytes))

	// attempt to decode file into wallets or return err
	return gobDecoder.Decode(wallets)

}

// SaveWalletsFile saves wallets to a file as bytes to the
// specified wallets file.
func SaveWalletsFile(wallets *map[string]*Wallet) {
	var buffer bytes.Buffer

	// register gob encoder and create a new encoder
	gob.Register(elliptic.P256())
	gobEncoder := gob.NewEncoder(&buffer)

	// attempt to encode wallets into bytes
	err := gobEncoder.Encode(wallets)
	if err != nil {
		log.Panicln("Unable to encode wallets using gob encoder: ", err.Error())
	}

	// write the bytes from the buffer into the specified file
	err = ioutil.WriteFile(os.Getenv("WALLETS_FILE"), buffer.Bytes(), 0644)
	if err != nil {
		log.Panicln("Unable to write wallets bytes buffer to a file: ", err.Error())
	}
}
