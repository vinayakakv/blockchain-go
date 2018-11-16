//Implements basic P2P functionality over TCP Network
package peer_to_peer

import (
	blockchain "../blockchain-core"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"sync"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

const (
	DIALTIMEOUT = time.Second * 2 //Timeout for Dialing connections
	RWTIMEOUT   = time.Second * 2 //Timeout for Read/Write with a Peer
)

//Callback is a handler for various messages received by a Peer
//p is the Peer who received the message
//conn is the connection from where the message was received
//arg is the data supplied in the message
type callback func(p *Peer, conn net.Conn, arg interface{})

//Represents a Node in P2P Network
type Peer struct {
	listenPort  uint16                 //TCP port where Current Peer is listening
	addr        string                 //Address of the Current Peer in the Network
	handlers    map[string]callback    //Handlers for each Message Type
	connections chan net.Conn          //Incoming connections to a Peer
	neighbours  sync.Map               //ip:port -> valid mapping
	blockchain  *blockchain.BlockChain //Blockchain of Current Peer
	log         *log.Entry             //Log Entry of Current Peer
}

//Returns the Network Address of current Peer
func (p *Peer) Addr() string {
	return p.addr
}

//Represents the Messages communicated between Peers
type Message struct {
	Action string      //Describes the intention of the message
	Data   interface{} //Contains the data required perform the intent
}

// Creates a peer listening at specified port
func CreatePeer(listenPort uint16) *Peer {
	return &Peer{
		listenPort: listenPort,
		handlers:   make(map[string]callback),
		blockchain: new(blockchain.BlockChain),
		log:        log.WithFields(log.Fields{"peer": listenPort}),
	}
}

//Returns the Blockchain of Current Peer
func (p *Peer) GetBlockChain() (*blockchain.BlockChain) {
	return p.blockchain
}

// Adds handler associated with particular message
func (p *Peer) AddHandler(key string, action callback) {
	p.handlers[key] = action
}

// Initiates the peer. Peer starts listening for incoming connections
func (p *Peer) init() {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", p.listenPort))
	p.blockchain.InitBlockChain()
	if err != nil {
		p.log.WithFields(log.Fields{
			"port": p.listenPort,
			"what": err,
		}).Panic("Error while attempting to listen")
	}
	p.addr = ln.Addr().String()
	p.log.WithFields(log.Fields{
		"address": fmt.Sprintf("tcp://%s", ln.Addr()),
	}).Info("Started Listening")
	p.connections = make(chan net.Conn)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				p.log.WithFields(log.Fields{
					"what": err,
				}).Errorf("Error while attempting to accept connection")
			}
			p.log.WithFields(log.Fields{
				"from": conn.RemoteAddr(),
				"to":   conn.LocalAddr(),
			}).Trace("New Connection")
			p.connections <- conn
		}
	}()
}

// Starts the peer node, Incoming connections are handled
func (p *Peer) Start() {
	p.init()
	go p.BroadcastBlockChain()
	go p.BroadcastAddition()
	for {
		go p.handleConn(<-p.connections)
	}
}

//Sends a Message to connection
//If wantsReply is true, wait for the reply
//Otherwise returns immediately after sending
//In case of timeout error, returns the error
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

//Adds a peer to neighbours after performing a handshake
//addr is the ip:port combination of another peer
// self -> Peer : PING
// Peer -> self : PONG
// self -> Peer : ACK
// at the end, self is in the neighbour list of Peer and vice versa
func (p *Peer) AddPeer(addr string) (e error) {
	p.log.WithFields(log.Fields{
		"addr": addr,
	}).Trace("Trying to add Peer")
	conn, err := net.DialTimeout("tcp", addr, DIALTIMEOUT)
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error in AddPeer")
		return
	}
	err = conn.SetDeadline(time.Now().Add(RWTIMEOUT))
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error in AddPeer")
		return
	}
	reply, e := Send(Message{"PING", map[string]uint16{"port": p.listenPort}}, conn, true)
	if e != nil {
		p.log.WithFields(log.Fields{
			"addr": addr,
			"what": e,
		}).Debugf("Unable to PING")
		return
	}
	if reply.Action == "PONG" {
		_, e = Send(Message{"ACK", nil}, conn, false)
		if e != nil {
			p.log.WithFields(log.Fields{
				"addr": addr,
				"what": e,
			}).Debugf("Unable to PONG")
			return
		}
		p.neighbours.Store(addr, true)
		p.log.WithFields(log.Fields{
			"addr": addr,
		}).Info("Successfully added peer")
	}
	p.log.Info("GETBLOCKCHAIN triggered")
	p.Broadcast(Message{Action: "GETBLOCKCHAIN"})
	return
}

//Incoming connection handler
//Decodes the JSON message and Executes associated handler
func (p *Peer) handleConn(client net.Conn) {
	defer client.Close()
	message := new(Message)
	err := json.NewDecoder(client).Decode(message)
	if err != nil {
		p.log.WithFields(log.Fields{
			"from": client.RemoteAddr(),
			"what": err,
		}).Error("Error while parsing JSON")
		return
	}
	p.log.WithFields(log.Fields{
		"action": message.Action,
		"from":   client.RemoteAddr(),
	}).Trace("Got a message")
	handler, exist := p.handlers[message.Action]
	if !exist {
		p.log.WithFields(log.Fields{
			"for":     message.Action,
			"request": message,
		}).Warn("No handler defined")
		return
	}
	handler(p, client, message.Data)
}

//Sends the Message to all the neighbours
func (p *Peer) Broadcast(m Message) {
	p.neighbours.Range(func(addr, valid interface{}) bool {
		addrStr := addr.(string)
		isValid := valid.(bool)
		if isValid {
			conn, err := net.DialTimeout("tcp4", addrStr, DIALTIMEOUT)
			//defer conn.Close()
			if err != nil {
				p.neighbours.Store(addr, false)
				p.log.WithFields(log.Fields{
					"to":   addrStr,
					"what": err,
				}).Error("Timeout error while Broadcasting")
				return true
			}
			conn.SetDeadline(time.Now().Add(RWTIMEOUT))
			_, err = Send(m, conn, false)
			if err != nil {
				p.neighbours.Store(addr, false)
				p.log.WithFields(log.Fields{
					"to":   addrStr,
					"what": err,
				}).Error("Send error while Broadcasting")
				return true
			}
		}
		return true
	})
}

//Broadcasts the current blockchain to all neighbours
func (p *Peer) BroadcastBlockChain() {
	var data bytes.Buffer
	p.blockchain.RLock()
	err := gob.NewEncoder(&data).Encode(p.blockchain)
	p.blockchain.RUnlock()
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error while gobbing")
		return
	}
	m := Message{Action: "BLOCKCHAINBCAST", Data: data.Bytes()}
	p.log.Info("BLOCKCHAINBCAST Triggered")
	p.Broadcast(m)
}

//Broadcasts new insertions to all neighbours
func (p *Peer) BroadcastAddition() {
	for {
		newBlock := p.blockchain.GetNewBlock()
		var data bytes.Buffer
		err := gob.NewEncoder(&data).Encode(*newBlock)
		if err != nil {
			p.log.WithFields(log.Fields{
				"what": err,
			}).Error("Error while gobbing")
			return
		}
		m := Message{Action: "NEWBLOCK", Data: data.Bytes()}
		p.log.WithFields(log.Fields{
			"index": newBlock.Index,
		}).Info("NEWBLOCK Triggered")
		p.Broadcast(m)
	}
}
