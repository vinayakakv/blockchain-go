package peer_to_peer

import (
	"net"
	"fmt"
	"log"
	"encoding/json"
)

type callback func(arg interface{})

//Represents a Node in P2P Network
type Peer struct {
	neighbours  []net.IP
	listenPort  uint16
	handlers    map[string]callback
	connections chan net.Conn
}

type Message struct {
	Action string
	Data   interface{}
}

// Creates a peer listening at specified port
func CreatePeer(listenPort uint16) *Peer {
	return &Peer{
		neighbours: make([]net.IP, 0),
		listenPort: listenPort,
		handlers:   make(map[string]callback),
	}
}

// Adds handler associated with particular message
func (p *Peer) AddHandler(key string, action callback) {
	p.handlers[key] = action
}

// Initiates the peer. Peer starts listening for incoming connections
func (p *Peer) init() {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", p.listenPort))
	if err != nil {
		log.Panicf("Error while attempting to listen at Port %d : %s\n", p.listenPort, err)
	}
	log.Printf("Listening at %s\n", ln.Addr())
	p.connections = make(chan net.Conn)
	go func() {
		log.Printf("Hi")
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("Error while attempting to accept connection %s\n", err)
			}
			log.Printf("Connection: %s <- %s \n", conn.LocalAddr(), conn.RemoteAddr())
			p.connections <- conn
		}
	}()
}

// Starts the peer node, Incoming connections are handled
func (p *Peer) Start() {
	p.init()
	p.Broadcast(Message{
		"PING",
		nil,
	})
	for {
		go handleConn(<-p.connections)
	}
}

func handleConn(client net.Conn) {
	message := new(Message)
	err := json.NewDecoder(client).Decode(message)
	if err != nil {
		log.Printf("Error while parsing JSON from %s. Error %s\n", client.RemoteAddr(), err)
		return
	}
	log.Printf("Got message %s from %s\n", message.Action, client.RemoteAddr())
}

func (p *Peer) Broadcast(message Message) {
	bcast_addr,err := net.ResolveUDPAddr("udp4","255.255.255.255:8084")
	if err !=nil {
		log.Printf("Error while resolving UDP adress. %v",err)
	}
	con, err := net.DialUDP("udp4", nil,bcast_addr)
	if err != nil {
		log.Printf("Error while attempting to broadcast %s\n", err)
		return
	}
	defer con.Close()
	json.NewEncoder(con).Encode(message)
}
