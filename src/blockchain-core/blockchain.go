package blockchain_core

import (
	"time"
	"fmt"
)

type Block struct {
	index uint64
	hash string
	previousHash string
	timestamp time.Time
	data string
}

func (b Block) getHash() string{
	return fmt.Sprintf("%ld%s%s%s",b.index,b.previousHash,b.timestamp.String(),b.data)
}