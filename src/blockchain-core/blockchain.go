package blockchain_core

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const (
	BlockGenerationInterval      = 10 * time.Second
	DifficultyAdjustmentInterval = 5
)

type BlockChain struct {
	mutex                sync.RWMutex
	insertions           chan *Block
	Chain                []*Block
	CumulativeDifficulty uint64
	dirty                bool
	Difficulty           uint64
}

func (bc *BlockChain) GetDifficulty() uint64 {
	latestBlock := bc.Chain[len(bc.Chain)-1]
	if latestBlock.Index%DifficultyAdjustmentInterval == 0 && latestBlock.Index != 0 {
		return bc.GetAdjustedDifficulty()
	} else {
		return latestBlock.Difficulty
	}
}

func (bc *BlockChain) GetNewBlock() *Block {
	return <-bc.insertions
}

func (bc *BlockChain) GetAdjustedDifficulty() uint64 {
	prevAdjustmentBlock := bc.Chain[len(bc.Chain)-DifficultyAdjustmentInterval]
	latestBlock := bc.Chain[len(bc.Chain)-1]
	timeExpected := BlockGenerationInterval * DifficultyAdjustmentInterval
	t1 := time.Unix(prevAdjustmentBlock.Timestamp, 0)
	t2 := time.Unix(latestBlock.Timestamp, 0)
	timeTaken := t2.Sub(t1)
	var newDifficulty uint64
	if timeTaken < timeExpected/2 {
		newDifficulty = prevAdjustmentBlock.Difficulty + 1
		log.WithFields(log.Fields{"difficulty": newDifficulty}).Infof("Increased Difficulty")
	} else if timeTaken > timeExpected*2 {
		newDifficulty = prevAdjustmentBlock.Difficulty - 1
		log.WithFields(log.Fields{"difficulty": newDifficulty}).Infof("Decreased Difficulty")
	} else {
		newDifficulty = prevAdjustmentBlock.Difficulty
	}
	bc.Difficulty = newDifficulty
	return newDifficulty
}

func (bc *BlockChain) InitBlockChain() {
	bc.Lock()
	defer bc.Unlock()
	if len(bc.Chain) == 0 {
		difficulty := uint64(1)
		bc.CumulativeDifficulty = difficulty
		bc.insertions = make(chan *Block)
		bc.Chain = append(bc.Chain, &Block{
			Index:        0,
			PreviousHash: "",
			Timestamp:    time.Now().Unix(),
			Data:         "Genesis Block",
			Difficulty:   difficulty,
		})
		bc.Chain[0].Mine()
	}
}

func (bc *BlockChain) Add(data string) {
	bc.Lock()
	defer bc.Unlock()
	b := CreateBlock(bc.Chain[len(bc.Chain)-1], data, bc.GetDifficulty())
	bc.dirty = true
	bc.Chain = append(bc.Chain, b)
	bc.CumulativeDifficulty += 1 << b.Difficulty
	bc.insertions <- b
}

func (bc *BlockChain) AddBlock(b *Block) {
	bc.Lock()
	defer bc.Unlock()
	bc.dirty = true
	if bc.Chain[len(bc.Chain)-1].Hash == b.PreviousHash {
		bc.Chain = append(bc.Chain, b)
		bc.CumulativeDifficulty += 1 << b.Difficulty
		bc.insertions <- b
	}
}

func (bc *BlockChain) Replace(other BlockChain) bool {
	bc.Lock()
	defer bc.Unlock()
	if other.IsValid() && other.CumulativeDifficulty > bc.CumulativeDifficulty {
		bc.Chain = other.Chain
		bc.CumulativeDifficulty = other.CumulativeDifficulty
		return true
	}
	return false
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
