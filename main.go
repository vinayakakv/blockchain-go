package main

import (
	peer "./src/peer-to-peer"
	"fmt"
	"github.com/c-bata/go-prompt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
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
		p.AddHandler("NEWBLOCK", peer.HandleNEWBLOCK)
		p.AddHandler("GETBLOCKCHAIN", peer.HandleGETBLOCKCHAIN)
		go p.Start()
	case "add", "a":
		p.AddPeer(parts[1])
	case "exit", "e":
		os.Exit(0)
	case "insert", "i":
		p.GetBlockChain().Add(parts[1])
	case "print", "p":
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

func Simulate(peerCount int, basePort uint16, insertCount int, networkType string) {
	// Phase 1 : Initialization
	peers := make([]*peer.Peer, peerCount)
	done := make(chan bool)
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

	//Phase 2: Connectivity
	switch networkType {
	case "fc":
		for i := 0; i < peerCount; i++ {
			for j := i + 1; j < peerCount; j++ {
				peers[i].AddPeer(peers[j].Addr())
			}
		}
	case "lin":
		for i := 0; i < peerCount-1; i++ {
			peers[i].AddPeer(peers[i+1].Addr())
		}
	case "cir":
		for i := 0; i < peerCount-1; i++ {
			peers[i].AddPeer(peers[i+1].Addr())
		}
		peers[peerCount-1].AddPeer(peers[0].Addr())
	case "ran":
		for i := 0; i < peerCount; i++ {
			for j := i + 1; j < peerCount; j++ {
				add := rand.Intn(2) == 1
				if add {
					peers[i].AddPeer(peers[j].Addr())
				}
			}
		}
	}

	//Phase 3 : Random insertions with delay
	for i := 0; i < insertCount; i++ {
		p := rand.Intn(peerCount)
		data := RandomString(10)
		sleep := time.Duration(rand.Intn(5))
		log.WithFields(log.Fields{
			"count": i,
			"peer":  peers[p].Addr(),
		}).Info("Insert Triggered")
		peers[p].GetBlockChain().Add(data)
		time.Sleep(sleep * time.Second)
		//done <- true
	}
	<-done
}

func main() {
	//Simulate(10, 10000, 70, "ran")
	RunDevelTerminal()
}
