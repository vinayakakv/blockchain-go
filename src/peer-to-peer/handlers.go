package peer_to_peer

import (
	"net"
	"log"
	"strings"
	"fmt"
	"time"
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
		p.neighbours[ip] = time.Now()
		log.Printf("Added %s to peer list", ip)
	}
}
