package main

import (
	peer "./src/peer-to-peer"
	//"fmt"
)

func main() {
	p := peer.CreatePeer(8090)
	p.AddPeer("0.0.0.0:9000")
}
