package blockchain_core

import (
	"time"
	"fmt"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}

type Block struct {
	Index        uint64
	Hash         string
	PreviousHash string
	Timestamp    int64
	Data         string
	Difficulty   uint64
	Nonce        string
}

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

func (b *Block) ToString() string {
	return fmt.Sprintf("%d%s%d%s%d%s", b.Index, b.PreviousHash, b.Timestamp, b.Data, b.Difficulty, b.Nonce)
}

func (b *Block) Mine() {
	prefix := strings.Repeat("0", int(b.Difficulty))
	for i := uint64(0); ; i++ {
		b.Nonce = fmt.Sprintf("%x", i)
		b.Hash = b.CalculateHash()
		if strings.HasPrefix(b.Hash, prefix) {
			log.WithFields(log.Fields{
				"index" : b.Index,
				"hash" : b.Hash,
				"difficulty" : b.Difficulty,
			}).Trace("Mined block")
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
