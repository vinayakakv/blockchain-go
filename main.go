package main

import (
	blockchain "./src/blockchain-core"
	peer "./src/peer-to-peer"
	"math/rand"
	"time"
	"fmt"
	"github.com/manifoldco/promptui"
	"strconv"
	"os"
	"strings"
	"bufio"
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
		blockchain.SetDifficulty(i)
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

func RunTerminal(p *peer.Peer) {
	bc := p.GetBlockChain()
	for ; ; {
		list := promptui.Select{
			Label: "Select Operation",
			Items: []string{
				"Set Difficulty",
				"Insert Block",
				"View Blockchain",
				"Analyze Mining",
				"Exit",
			},
			//Templates: &templates,
		}
		handlers := []func(){
			func() {
				prompt := promptui.Prompt{Label: "Difficulty"}
				data, err := prompt.Run()
				if err != nil {
					panic(err)
				}
				difficulty, err := strconv.Atoi(data)
				if err != nil {
					fmt.Printf("%s\n", err)
					return
				}
				blockchain.SetDifficulty(uint64(difficulty))
			},
			func() {
				prompt := promptui.Prompt{Label: "String Data"}
				data, err := prompt.Run()
				if err != nil {
					panic(err)
				}
				bc.Add(data)
				fmt.Printf("Data Successfully Inserted\n")
			},
			func() {
				bc.Print()
			},
			func() {
				prompt := promptui.Prompt{Label: "maxDifficulty"}
				data, err := prompt.Run()
				if err != nil {
					panic(err)
				}
				maxDifficulty, err := strconv.Atoi(data)
				if err != nil {
					fmt.Printf("%s\n", err)
					return
				}
				prompt = promptui.Prompt{Label: "maxRuns"}
				data, err = prompt.Run()
				if err != nil {
					panic(err)
				}
				maxRuns, err := strconv.Atoi(data)
				if err != nil {
					fmt.Printf("%s\n", err)
					return
				}
				res := AnalyzeMining(uint64(maxDifficulty), uint64(maxRuns))
				fmt.Printf("%v\n", res)
			},
			func() {
				os.Exit(0)
			},
		}
		idx, _, err := list.Run()
		if err != nil {
			panic(err)
		}
		handlers[idx]()
	}
}

func RunDevelTerminal(p *peer.Peer) {
	for {
		fmt.Printf(">>")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		parts := strings.Split(input, " ")
		for i := 0; i < len(parts); i++ {
			parts[i] = strings.TrimSpace(parts[i])
		}
		switch parts[0] {
		case "createPeer":
			port, _ := strconv.Atoi(parts[1])
			fmt.Printf("%v\n", uint16(port))
			p = peer.CreatePeer(uint16(port))
			p.AddHandler("PING", peer.HandlePING)
			p.AddHandler("BLOCKCHAINBCAST", peer.HandleBLOCKCHAINBCAST)
			go p.Start()
		case "addPeer":
			p.AddPeer(parts[1])
		case "exit":
			return
		case "addBlock":
			p.GetBlockChain().Add(parts[1])
		case "checkValid":
			fmt.Printf("%v", p.GetBlockChain().IsValid())
		default:
			fmt.Printf("Unknown command %s", parts[0])
		}
	}
}

func main() {
	p := peer.Peer{}
	RunDevelTerminal(&p)
}
