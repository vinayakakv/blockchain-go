package peer_to_peer

import (
	"net"
	"fmt"
	"log"
	"encoding/json"
	"errors"
	"strings"
)

type callback func(arg interface{})

//Represents a Node in P2P Network
type Peer struct {
	neighbours  []net.IP
	listenPort  uint16
	handlers    map[string]callback
	connections chan net.Conn
	bacstSoc    net.PacketConn
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
	pc, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", p.listenPort))
	p.bacstSoc = pc
	if err != nil {
		log.Panicf("Error while attempting to listen at Port %d : %s\n", p.listenPort, err)
	}
	log.Printf("Listening at tcp://%s\tudp://%s\n", ln.Addr(), pc.LocalAddr())
	p.connections = make(chan net.Conn)
	go handleBcast(pc)
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

func handleBcast(client net.PacketConn) {
	//log.Printf("Handling broadcasts at %s",client.LocalAddr())
	for {
		b := make([]byte, 1024)
		n, addr, err := client.ReadFrom(b)
		if err != nil {
			log.Panicf("%s", err)
		}
		log.Printf("Recieved bcast from %s. Message is %s", addr, b[:n])
		extip, _ := externalIP()
		log.Printf("External IP is %s", extip)
		if !strings.Contains(addr.String(), extip) {
			client.WriteTo([]byte("PONG"), addr)
		}
	}
}

func (p *Peer) Broadcast(message Message) {
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", p.listenPort))
	//laddr,_ := net.ResolveUDPAddr("udp4",fmt.Sprintf(":%d",p.listenPort))
	_, err := p.bacstSoc.WriteTo([]byte("PING"), addr)
	if err != nil {
		log.Printf("Error while attempting to broadcast %s\n", err)
		return
	}
	log.Printf("Sent broadcast from %s", p.bacstSoc.LocalAddr())
	buffer := make([]byte, 1024)
	n, a, _ := p.bacstSoc.ReadFrom(buffer)
	log.Printf("Reply is %s from %s", buffer[:n], a)
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("not connected to network")
}
