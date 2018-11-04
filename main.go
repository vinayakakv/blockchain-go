package main

import (
	blockchain "./src/blockchain-core"
	peer "./src/peer-to-peer"
	"math/rand"
	"time"
	"fmt"
	"strconv"
	"strings"
	"github.com/c-bata/go-prompt"
	"os"
)

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

func AnalyzeMining(maxDifficulty uint64, insertCount uint64) []uint64 {
	bc := blockchain.BlockChain{}
	bc.InitBlockChain()
	avgTime := make([]uint64, maxDifficulty)
	for i := uint64(1); i <= maxDifficulty; i++ {
		bc.SetDifficulty(i)
		fmt.Printf("Current Difficulty is %d\n", i)
		for j := uint64(0); j < insertCount; j++ {
			start := time.Now()
			bc.Add(RandomString(10))
			avgTime[i-1] += uint64(time.Since(start) / time.Microsecond)
		}
		avgTime[i-1] /= insertCount
	}
	return avgTime
}

var p = &peer.Peer{}

var suggestions = []prompt.Suggest{
	{Text: "init", Description: "Creates the Peer listening on specified port"},
	{Text: "add", Description: "Adds Peer specified by address to neighbour list"},
	{Text: "exit", Description: "Quits the program"},
	{Text: "insert", Description: "Inserts a block into Blockchain"},
	{Text: "print", Description: "Prints the Blockchain"},
}

func executor(input string) {
	input = strings.TrimSpace(input)
	parts := strings.Split(input, " ")
	switch parts[0] {
	case "init":
		port, _ := strconv.Atoi(parts[1])
		p = peer.CreatePeer(uint16(port))
		p.AddHandler("PING", peer.HandlePING)
		p.AddHandler("BLOCKCHAINBCAST", peer.HandleBLOCKCHAINBCAST)
		go p.Start()
	case "add":
		p.AddPeer(parts[1])
	case "exit":
		os.Exit(0)
	case "insert":
		p.GetBlockChain().Add(parts[1])
	case "print":
		p.GetBlockChain().Print()
	default:
		fmt.Printf("Unknown command %s\n", parts[0])
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if len(strings.Split(in.TextBeforeCursor(), " ")) > 1 || w == "" {
		return []prompt.Suggest{}
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}

func RunDevelTerminal() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>"),
		prompt.OptionTitle("bc-prompt"),
	)
	p.Run()
}

func main() {
	RunDevelTerminal()
}
