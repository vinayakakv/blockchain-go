//P2P without peer discovery!
package peer_to_peer

import (
	"net"
	"fmt"
	"log"
	"encoding/json"
	"errors"
	"time"
	"sync"
)

const (
	DIALTIMEOUT = time.Second * 5
	RWTIMEOUT   = time.Second * 80
)

type callback func(p *Peer, conn net.Conn, arg interface{})

//Represents a Node in P2P Network
type Peer struct {
	listenPort  uint16
	handlers    map[string]callback
	connections chan net.Conn
	neighbours  sync.Map //ip:port -> valid mapping
}

type Message struct {
	Action string
	Data   interface{}
}

// Creates a peer listening at specified port
func CreatePeer(listenPort uint16) *Peer {
	return &Peer{
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

//Adds a peer to neighbours after performing a handshake
//addr is the ip:port combination of another peer
// self -> Peer : PING
// Peer -> self : PONG
// self -> Peer : ACK
// at the end, self is in the neighbour list of Peer and vice versa
func (p *Peer) AddPeer(addr string) (e error) {
	log.Printf("Trying to add Peer %s", addr)
	conn, err := net.DialTimeout("tcp4", addr, DIALTIMEOUT)
	if err != nil {
		log.Printf("Error in AddPeer %s", err)
		return
	}
	conn.SetDeadline(time.Now().Add(RWTIMEOUT))
	if err != nil {
		log.Printf("Error in AddPeer %s", err)
		return
	}
	reply, e := Send(Message{"PING", map[string]uint16{"port": p.listenPort}}, conn, true)
	if e != nil {
		log.Printf("Unable to PING %s : %s", addr, e)
		return
	}
	if reply.Action == "PONG" {
		_, e = Send(Message{"ACK", nil}, conn, false)
		if e != nil {
			log.Printf("Unable to PONG %s : %s", addr, e)
			return
		}
		p.neighbours.Store(addr, true)
		log.Printf("Successfully added peer %s", addr)
	}
	return
}

func (p *Peer) handleConn(client net.Conn) {
	defer client.Close()
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
	handler(p, client, message.Data)
}

func (p *Peer) Broadcast(m Message) {
	p.neighbours.Range(func(addr, valid interface{}) bool {
		addrStr := addr.(string)
		isValid := valid.(bool)
		if isValid {
			conn, err := net.DialTimeout("tcp4", addrStr, DIALTIMEOUT)
			if err != nil {
				p.neighbours.Store(addr, false)
				return true
			}
			conn.SetDeadline(time.Now().Add(RWTIMEOUT))
			_, err = Send(m, conn, false)
			if err != nil {
				p.neighbours.Store(addr, false)
				return true
			}
		}
		return true
	})
}
func Send(message Message, conn net.Conn, wantsReply bool) (reply Message, e error) {
	if conn == nil {
		e = errors.New("attempting to write to nil connection")
		return
	}
	e = json.NewEncoder(conn).Encode(message)
	if e != nil {
		return
	}
	if wantsReply {
		e = json.NewDecoder(conn).Decode(&reply)
		if e != nil {
			return
		}
	}
	return
}
