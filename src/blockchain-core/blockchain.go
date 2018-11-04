package blockchain_core

import (
	"sync"
	"time"
	log "github.com/sirupsen/logrus"
	"strings"
	"github.com/davecgh/go-spew/spew"
	"fmt"
)

var difficulty uint64 = 3

type BlockChain struct {
	mutex                sync.RWMutex
	Chain                []*Block
	CumulativeDifficulty uint64
	dirty                bool
}

func SetDifficulty(n uint64) {
	difficulty = n
}

func (bc *BlockChain) InitBlockChain() {
	bc.Lock()
	defer bc.Unlock()
	if len(bc.Chain) == 0 {
		bc.CumulativeDifficulty = difficulty
		bc.Chain = append(bc.Chain, &Block{
			Index:        0,
			PreviousHash: "",
			Timestamp:    time.Now().String(),
			Data:         "Genesis Block",
			Difficulty:   difficulty,
		})
		bc.Chain[0].Mine()
	}
}

func (bc *BlockChain) Add(data string) {
	bc.Lock()
	defer bc.Unlock()
	b := CreateBlock(bc.Chain[len(bc.Chain)-1], data)
	bc.dirty = true
	bc.Chain = append(bc.Chain, b)
	bc.CumulativeDifficulty += b.Difficulty
}

func (bc *BlockChain) Replace(other BlockChain) {
	bc.Lock()
	defer bc.Unlock()
	if !bc.dirty && other.IsValid() && other.CumulativeDifficulty > bc.CumulativeDifficulty {
		bc.Chain = other.Chain
		bc.CumulativeDifficulty = other.CumulativeDifficulty
		log.Printf("Blockchain replaced")
	}
}

func (bc *BlockChain) IsValid() bool {
	bc.RLock()
	defer bc.RUnlock()
	for i := 0; i < len(bc.Chain)-1; i++ {
		current := bc.Chain[i]
		next := bc.Chain[i+1]
		if next.Index != current.Index+1 ||
			next.PreviousHash != current.Hash ||
			current.CalculateHash() != current.Hash ||
			!strings.HasPrefix(current.Hash, strings.Repeat("0", int(current.Difficulty))) {
			return false
		}
	}
	last := bc.Chain[len(bc.Chain)-1]
	if last.CalculateHash() != last.Hash ||
		!strings.HasPrefix(last.Hash, strings.Repeat("0", int(last.Difficulty))) {
		return false
	}
	return true
}

func (bc *BlockChain) ClearDirty() {
	bc.Lock()
	defer bc.Unlock()
	bc.dirty = false
}

func (bc *BlockChain) Dump() {
	bc.RLock()
	defer bc.RUnlock()
	for _, block := range bc.Chain {
		spew.Dump(block)
	}
}

func (bc *BlockChain) Print() {
	bc.RLock()
	defer bc.RUnlock()
	str := "["
	for _, block := range bc.Chain {
		str += block.Data + ","
	}
	str += "]\n"
	fmt.Print(str)
}

func (bc *BlockChain) Lock() {
	bc.mutex.Lock()
}

func (bc *BlockChain) RLock() {
	bc.mutex.RLock()
}

func (bc *BlockChain) Unlock() {
	bc.mutex.Unlock()
}

func (bc *BlockChain) RUnlock() {
	bc.mutex.RUnlock()
}
