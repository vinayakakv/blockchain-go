//Contains the Data Structures and Algorithms to implement a Simple Blockchain
package blockchain_core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

// Basic Block Structure
type Block struct {
	Index        uint64 //Position of the Block in Blockchain
	Hash         string //SHA-256 Hash of Current Block Data
	PreviousHash string //SHA-256 Hash of Previous Block Data
	Timestamp    int64  //Creation time of Current Block in Unix Format
	Data         string //Block Data
	Difficulty   uint64 //Number of zeros should be present as prefix in hash
	Nonce        string //Number in hex format found in order to satisfy difficulty
}

//Creates a new block and mines it
//Requires previous block,current block data and difficulty
//Returns a pointer to newly created block
func CreateBlock(oldBlock *Block, data string, difficulty uint64) *Block {
	b := &Block{
		Index:        oldBlock.Index + 1,
		PreviousHash: oldBlock.Hash,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		Difficulty:   difficulty,
	}
	b.Mine()
	return b
}

//Returns string representation of current block.
//Contains Index,PreviousHash,Timestamp,Data,Difficulty and Nonce
func (b *Block) ToString() string {
	return fmt.Sprintf("%d%s%d%s%d%s", b.Index, b.PreviousHash, b.Timestamp, b.Data, b.Difficulty, b.Nonce)
}

//Mines current block
//Mining is the task of finding Nonce in order to match Difficulty
func (b *Block) Mine() {
	prefix := strings.Repeat("0", int(b.Difficulty))
	for i := uint64(0); ; i++ {
		b.Nonce = fmt.Sprintf("%x", i)
		b.Hash = b.CalculateHash()
		if strings.HasPrefix(b.Hash, prefix) {
			log.WithFields(log.Fields{
				"index":      b.Index,
				"hash":       b.Hash,
				"difficulty": b.Difficulty,
			}).Trace("Mined block")
			break
		}
	}
}

//Helper method to calculate SHA-256 Hash of Block
func (b *Block) CalculateHash() string {
	record := b.ToString()
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
