package main

import (
	//"fmt"
	p2p "./src/peer-to-peer"
)

func main() {
	x := p2p.CreatePeer(8085)
	x.Start()
}
