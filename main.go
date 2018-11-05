package main

import (
	peer "./src/peer-to-peer"
	"math/rand"
	"fmt"
	"strconv"
	"strings"
	"github.com/c-bata/go-prompt"
	"os"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
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

func Simulate(peerCount int, basePort uint16, insertCount int) {
	// Phase 1 : Initialization
	peers := make([]*peer.Peer, peerCount)
	var wg sync.WaitGroup
	wg.Add(peerCount)
	for i := 0; i < peerCount; i++ {
		go func(i int) {
			defer wg.Done()
			peers[i] = peer.CreatePeer(basePort + uint16(i))
			peers[i].AddHandler("PING", peer.HandlePING)
			peers[i].AddHandler("BLOCKCHAINBCAST", peer.HandleBLOCKCHAINBCAST)
			go peers[i].Start()
		}(i)
	}
	wg.Wait()
	for i := 0; i < peerCount; i++ {
		for j := i + 1; j < peerCount; j++ {
			peers[i].AddPeer(peers[j].Addr())
		}
	}
	//Phase 2 : Random insertions with delay
	wg.Add(insertCount)
	for i := 0; i < insertCount; i++ {
		go func(i int) {
			defer wg.Done()
			time.Sleep(1 * time.Second)
			log.WithFields(log.Fields{"count": i}).Info("Insert Triggered")
			peer := rand.Intn(peerCount)
			data := RandomString(10)
			//sleep := time.Duration(rand.Intn(5))
			peers[peer].GetBlockChain().Add(data)
		}(i)
		//done <- true
	}
	wg.Wait()
}

func main() {
	//Simulate(10, 10000, 100)
	RunDevelTerminal()
}
