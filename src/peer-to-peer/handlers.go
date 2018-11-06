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
)


func HandlePING(p *Peer, conn net.Conn, data interface{}) {
	body, ok := data.(map[string]interface{})
	if !ok {
		p.log.Error("Invalid data recieved in PING")
		return
	}
	reply, err := Send(Message{"PONG", nil}, conn, true)
	if err != nil {
		p.log.WithFields(log.Fields{
			"what" : err,
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
			"ip" : ip,
		}).Info("Added to peer list")
	}
}

func HandleBLOCKCHAINBCAST(p *Peer, conn net.Conn, data interface{}) {
	buf := new(bytes.Buffer)
	byteData, err := base64.StdEncoding.DecodeString(data.(string))
	if err != nil {
		p.log.WithFields(log.Fields{
			"what" : err,
		}).Error("Error while base64 decode")
		return
	}
	buf.Write(byteData)
	bc := blockchain.BlockChain{}
	err = gob.NewDecoder(buf).Decode(&bc)
	if err != nil {
		p.log.WithFields(log.Fields{
			"what" : err,
		}).Error("Error while ungobbing")
		return
	}
	replaced := p.blockchain.Replace(bc)
	if replaced {
		p.log.Info("Blockchain replaced")
	}
}
