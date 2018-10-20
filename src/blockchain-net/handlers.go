package blockchain_net

import (
	blockchain "../blockchain-core"
	peer "../peer-to-peer"
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


