package blockchain_core

import (
	"time"
	"fmt"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"log"
)

type Block struct {
	index        uint64
	hash         string
	previousHash string
	timestamp    string
	data         string
	difficulty   uint64
	nonce        string
}

func CreateBlock(oldBlock Block, data string) Block {
	b := Block{
		index:        oldBlock.index + 1,
		previousHash: oldBlock.hash,
		timestamp:    time.Now().String(),
		data:         data,
		difficulty:   difficulty,
	}
	b.calculateHash()
	return b
}

func (b Block) ToString() string {
	return fmt.Sprintf("%d%s%s%d%s%d%s", b.index, b.hash, b.previousHash, b.timestamp, b.data, b.difficulty, b.nonce)
}

func (b *Block) calculateHash() {
	prefix := strings.Repeat("0", int(b.difficulty))
	for i := uint64(0); ; i++ {
		b.nonce = fmt.Sprintf("%x", i)
		record := b.ToString()
		h := sha256.New()
		h.Write([]byte(record))
		hashed := h.Sum(nil)
		b.hash = hex.EncodeToString(hashed)
		if strings.HasPrefix(b.hash, prefix) {
			log.Printf("Mined block %d : Hash %s\n", b.index, b.hash)
			break
		}
	}
}
