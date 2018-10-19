package blockchain_core

import (
	"time"
	"github.com/davecgh/go-spew/spew"
	"strings"
)

const difficulty = 3

type BlockChain struct {
	chain []Block
}

func (bc *BlockChain) InitBlockChain() {
	//bc.chain = make([]Block,0)
	bc.chain = append(bc.chain, Block{
		index:        0,
		previousHash: "",
		timestamp:    time.Now().Unix(),
		data:         "Genesis Block",
		difficulty:   difficulty,
	})
	bc.chain[0].calculateHash()
}

func (bc *BlockChain) Add(data string) {
	b := CreateBlock(bc.chain[len(bc.chain)-1], data)
	bc.chain = append(bc.chain, b)
}

func (bc *BlockChain) IsValid() bool {
	for i := 0; i < len(bc.chain)-1; i++ {
		current := bc.chain[i]
		next := bc.chain[i+1]
		if next.index != current.index+1 || next.previousHash != current.hash || !strings.HasPrefix(current.hash, strings.Repeat("0", current.difficulty)) {
			return false
		}
	}
	last := bc.chain[len(bc.chain)-1]
	if !strings.HasPrefix(last.hash, strings.Repeat("0", last.difficulty)) {
		return false
	}
	return true
}

func (bc *BlockChain) Print() {
	for _, block := range bc.chain {
		spew.Dump(block)
	}
}
