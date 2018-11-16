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
	BlockGenerationInterval      = 10 * time.Second //Average time between insertion of blocks
	DifficultyAdjustmentInterval = 5                //Interval for reviewing difficulty
)

//Implements in-memory Blockchain structure
type BlockChain struct {
	mutex                sync.RWMutex //Lock to handle concurrent r/w
	insertions           chan *Block  //Chanel for newly inserted blocks
	Chain                []*Block     //Core Blockchain
	CumulativeDifficulty uint64       //Sum(2^difficulty_b) for all Blocks b
	Difficulty           uint64       //Current difficulty of blockchain
}

//Returns the difficulty for new block
//Adjusted difficulty is returned if Last Block Index is a multiple of Difficulty Adjustment Interval
func (bc *BlockChain) GetDifficulty() uint64 {
	latestBlock := bc.Chain[len(bc.Chain)-1]
	if latestBlock.Index%DifficultyAdjustmentInterval == 0 && latestBlock.Index != 0 {
		return bc.GetAdjustedDifficulty()
	} else {
		return latestBlock.Difficulty
	}
}

//Returns a block form Insert Chanel of Blockchain
func (bc *BlockChain) GetNewBlock() *Block {
	return <-bc.insertions
}

//Returns adjusted difficulty for the blockchain
//Difficulty is increased if block insertions are too fast
//Difficulty is decreased if block insertions are too slow
//Difficulty is kept constant if block insertions are at normal rate
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

//Initiates a Blockchain
//Initial Difficulty is set to 1
//Creates and mines Genesis Block
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

//Adds a new block containing the data to the Blockchain
//New block is also mined
//Mined block is also sent to broadcast list
func (bc *BlockChain) Add(data string) {
	bc.Lock()
	defer bc.Unlock()
	b := CreateBlock(bc.Chain[len(bc.Chain)-1], data, bc.GetDifficulty())
	bc.Chain = append(bc.Chain, b)
	bc.CumulativeDifficulty += 1 << b.Difficulty
	bc.insertions <- b
}

//Adds a new block to the Blockchain
//New block is verified before adding it to the blockchain
func (bc *BlockChain) AddBlock(b *Block) bool {
	bc.Lock()
	defer bc.Unlock()
	if b.CalculateHash() == b.Hash && bc.Chain[len(bc.Chain)-1].Hash == b.PreviousHash && bc.Chain[len(bc.Chain)-1].Index == b.Index-1 {
		bc.Chain = append(bc.Chain, b)
		bc.CumulativeDifficulty += 1 << b.Difficulty
		bc.insertions <- b
		return true
	}
	return false
}

//Replaces the current blockchain with other blockchain
//Replacement happens iff other blockchain is valid and has a cumulative difficulty greater than that of current block
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

//Returns weather current blockchain is valid
//Verifies block integrity and hash linkages
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

//Dumps the entire blockchain in spew format
func (bc *BlockChain) Dump() {
	bc.RLock()
	defer bc.RUnlock()
	for _, block := range bc.Chain {
		spew.Dump(block)
	}
}

//Prints the blockchain as list to stdout
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

//Locks the blockchain for writing
func (bc *BlockChain) Lock() {
	bc.mutex.Lock()
}

//Locks the blockchain for reading
func (bc *BlockChain) RLock() {
	bc.mutex.RLock()
}

//Unlocks the blockchain which was locked for writing
func (bc *BlockChain) Unlock() {
	bc.mutex.Unlock()
}

//Unlocks the blockchain which was locked for reading
func (bc *BlockChain) RUnlock() {
	bc.mutex.RUnlock()
}
