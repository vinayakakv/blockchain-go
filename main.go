package main

import (
	blockchain "./src/blockchain-core"
	"math/rand"
	"time"
	"fmt"
	"github.com/manifoldco/promptui"
	"strconv"
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

func RunTerminal(){
	bc := blockchain.BlockChain{}
	for ; ; {
		list := promptui.Select{
			Label: "Select Operation",
			Items: []string{
				"Create Blockchain",
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
				fmt.Printf("Creating Blockchain\n", )
				bc.InitBlockChain()
			},
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

func main() {
	RunTerminal()
}
