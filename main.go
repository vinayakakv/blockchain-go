package main

import (
	peer "./src/peer-to-peer"
	//"fmt"
	"fmt"
)

func main() {
	p := peer.CreatePeer(8090)
	//p.AddPeer("0.0.0.0:9000")
	p.AddHandler("PING", peer.HandlePING)
	p.Start()
	addr := ""
	fmt.Printf("Enter the adress of the client to connect to\n")
	fmt.Scanf("%s",addr)
	p.AddPeer(addr)
}
