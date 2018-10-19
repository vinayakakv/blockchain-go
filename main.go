package main

import (
	//"fmt"
	blockchain "./src/blockchain-core"
	"fmt"
)

func main() {
	bc := blockchain.BlockChain{}
	bc.InitBlockChain()
	for i := 0; i < 100; i++ {
		bc.Add(fmt.Sprintf("%x", i))
	}
	bc.Print()
	fmt.Println(bc.IsValid())
}
