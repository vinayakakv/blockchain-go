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
	Index        uint64
	Hash         string
	PreviousHash string
	Timestamp    string
	Data         string
	Difficulty   uint64
	Nonce        string
}

func CreateBlock(oldBlock *Block, data string) *Block {
	b := &Block{
		Index:        oldBlock.Index + 1,
		PreviousHash: oldBlock.Hash,
		Timestamp:    time.Now().String(),
		Data:         data,
		Difficulty:   difficulty,
	}
	b.Mine()
	return b
}

func (b *Block) ToString() string {
	return fmt.Sprintf("%d%s%d%s%d%s", b.Index, b.PreviousHash, b.Timestamp, b.Data, b.Difficulty, b.Nonce)
}

func (b *Block) Mine() {
	prefix := strings.Repeat("0", int(b.Difficulty))
	for i := uint64(0); ; i++ {
		b.Nonce = fmt.Sprintf("%x", i)
		b.Hash = b.CalculateHash()
		if strings.HasPrefix(b.Hash, prefix) {
			log.Printf("Mined block %d : Hash %s\n", b.Index, b.Hash)
			break
		}
	}
}

func (b *Block) CalculateHash() string {
	record := b.ToString()
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
