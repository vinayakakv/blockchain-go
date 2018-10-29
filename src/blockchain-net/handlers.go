package blockchain_net

import (
	blockchain "../blockchain-core"
	peer "../peer-to-peer"
	"net"
)

type BlockChainNode struct {
	peer       peer.Peer
	blockchain blockchain.BlockChain
}

func CreateBlockChainNode(listenPort uint16) *BlockChainNode {
	bcn := new(BlockChainNode)
	bcn.blockchain = blockchain.BlockChain{}
	bcn.blockchain.InitBlockChain()
	bcn.peer = *peer.CreatePeer(listenPort)
	return bcn
}

func HandleGETBLOCKCHAIN(p *peer.Peer, conn net.Conn, arg interface{}) {

}

func (bcn *BlockChainNode) StartBlockChainNode() {
	bcn.peer.AddHandler("PING", peer.HandlePING)
	bcn.peer.Broadcast(peer.Message{"GETBLOCKCHAIN", nil})
	bcn.peer.Start()
}
