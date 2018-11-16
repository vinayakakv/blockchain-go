package peer_to_peer

import (
	blockchain "../blockchain-core"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"
)

//Handler for the PING message
//Replies for a PONG and waits for ACK
//If ACK comes, adds other peer to neighbour list
func HandlePING(p *Peer, conn net.Conn, data interface{}) {
	body, ok := data.(map[string]interface{})
	if !ok {
		p.log.Error("Invalid data recieved in PING")
		return
	}
	reply, err := Send(Message{"PONG", nil}, conn, true)
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error while waiting for ACK")
		return
	}
	if reply.Action != "ACK" {
		log.Error("Invalid reply recieved")
		return
	} else {
		ip := conn.RemoteAddr().String()
		ip = strings.Split(ip, ":")[0]
		ip = ip + fmt.Sprintf(":%v", body["port"])
		p.neighbours.Store(ip, true)
		p.log.WithFields(log.Fields{
			"ip": ip,
		}).Info("Added to peer list")
	}
	p.log.Info("GETBLOCKCHAIN triggered")
	p.Broadcast(Message{Action: "GETBLOCKCHAIN"})
}

//Handler for NEWBLOCK
//Validates the incoming block for hash and timestamp validity
//If valid, tries to insert it into the blockchain. In case of failure, triggers GETBLOCKCHAIN
func HandleNEWBLOCK(p *Peer, conn net.Conn, data interface{}) {
	buf := new(bytes.Buffer)
	byteData, err := base64.StdEncoding.DecodeString(data.(string))
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error while base64 decode")
		return
	}
	buf.Write(byteData)
	b := &blockchain.Block{}
	err = gob.NewDecoder(buf).Decode(&b)
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error while ungobbing")
		return
	}
	lastBlock := p.blockchain.Chain[len(p.blockchain.Chain)-1]
	if b.Index > lastBlock.Index {
		t1 := time.Unix(lastBlock.Timestamp, 0)
		now := time.Now()
		diff := t1.Sub(now)
		if lastBlock.Hash != lastBlock.CalculateHash() && diff > 2*time.Second && diff < -2*time.Second {
			p.log.WithFields(log.Fields{
				"index": b.Index,
				"data":  b.Data,
			}).Info("Invalid block received")
		} else if lastBlock.Hash == b.PreviousHash {
			p.blockchain.AddBlock(b)
			p.log.WithFields(log.Fields{
				"index": b.Index,
				"data":  b.Data,
			}).Info("Inserted Block")
		} else {
			p.log.Info("GETBLOCKCHAIN triggered")
			p.Broadcast(Message{Action: "GETBLOCKCHAIN"})
		}
	}
}

//Handler for GETBLOCKCHAIN
//Broadcasts the blockchain to all neighbours
func HandleGETBLOCKCHAIN(p *Peer, conn net.Conn, data interface{}) {
	p.BroadcastBlockChain()
}

//Handler for GETBLOCKCHAIN
//Tries to replace the blockchain with received one
func HandleBLOCKCHAINBCAST(p *Peer, conn net.Conn, data interface{}) {
	buf := new(bytes.Buffer)
	byteData, err := base64.StdEncoding.DecodeString(data.(string))
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error while base64 decode")
		return
	}
	buf.Write(byteData)
	bc := blockchain.BlockChain{}
	err = gob.NewDecoder(buf).Decode(&bc)
	if err != nil {
		p.log.WithFields(log.Fields{
			"what": err,
		}).Error("Error while ungobbing")
		return
	}
	replaced := p.blockchain.Replace(bc)
	if replaced {
		p.log.Info("Blockchain replaced")
	}
}
