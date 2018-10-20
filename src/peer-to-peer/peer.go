//P2P without peer discovery!
package peer_to_peer

import (
	"net"
	"fmt"
	"log"
	"encoding/json"
	"errors"
)

type callback func(conn net.Conn, arg interface{})

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
	log.Printf("Listening at tcp://%s", ln.Addr())
	p.connections = make(chan net.Conn)
	go func() {
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
	for {
		go p.handleConn(<-p.connections)
	}
}

func (p *Peer) AddPeer(addr string) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		log.Printf("%s",err)
		return
	}
	reply, _ := p.Send(Message{"PING", nil}, conn)
	log.Printf("%s", reply)
}

func (p *Peer) handleConn(client net.Conn) {
	message := new(Message)
	err := json.NewDecoder(client).Decode(message)
	if err != nil {
		log.Printf("Error while parsing JSON from %s. Error %s\n", client.RemoteAddr(), err)
		return
	}
	log.Printf("Got message %s from %s\n", message.Action, client.RemoteAddr())
	handler, exist := p.handlers[message.Action]
	if !exist {
		log.Printf("No handler for %s defined. Ignoring request %v", message.Action, message)
		return
	}
	handler(client, message.Data)
}

func (p *Peer) Send(message Message, conn net.Conn) (reply Message, e error) {
	if conn == nil {
		e = errors.New("attempting to write to nil connection")
		return
	}
	json.NewEncoder(conn).Encode(message)
	json.NewDecoder(conn).Decode(reply)
	return
}
