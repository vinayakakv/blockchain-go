package peer_to_peer

import (
	"net"
	"log"
	"strings"
	"fmt"
	"bytes"
	"encoding/gob"
	blockchain "../blockchain-core"
	"encoding/base64"
)

func HandlePING(p *Peer, conn net.Conn, data interface{}) {
	body, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid data recieved in PING")
		return
	}
	reply, err := Send(Message{"PONG", nil}, conn, true)
	if err != nil {
		log.Printf("Error while waiting for ACK : %s", err)
		return
	}
	if reply.Action != "ACK" {
		log.Printf("Invalid reply recieved")
		return
	} else {
		ip := conn.RemoteAddr().String()
		ip = strings.Split(ip, ":")[0]
		ip = ip + fmt.Sprintf(":%v", body["port"])
		p.neighbours.Store(ip, true)
		log.Printf("Added %s to peer list", ip)
	}
}

func HandleBLOCKCHAINBCAST(p *Peer, conn net.Conn, data interface{}) {
	buf := new(bytes.Buffer)
	byteData, err := base64.StdEncoding.DecodeString(data.(string))
	if err != nil {
		log.Printf("Error while base64 decode : %s", err)
		return
	}
	buf.Write(byteData)
	bc := blockchain.BlockChain{}
	err = gob.NewDecoder(buf).Decode(&bc)
	if err != nil {
		log.Printf("Error while ungobbing %s", err)
		return
	}
	p.blockchain.Replace(bc)
	p.blockchain.Print()
}
