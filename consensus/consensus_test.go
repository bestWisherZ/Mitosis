package consensus

import (
	"bytes"
	"encoding/binary"

	// "fmt"
	"testing"
	"time"

	"github.com/KyrinCode/Mitosis/config"
	"github.com/KyrinCode/Mitosis/eventbus"
	"github.com/KyrinCode/Mitosis/p2p"

	// "github.com/KyrinCode/Mitosis/types"
	"github.com/KyrinCode/Mitosis/core"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewConfig(nodeId uint32) *config.Config {
	topo := config.Topology{
		RShardIds: []uint32{1, 2},
		PShardIds: map[uint32][]uint32{
			1: {1001, 1002},
			2: {1003, 1004},
		},
	}

	accounts := make([]common.Address, 10)
	for i := 0; i < len(accounts); i++ {
		data := int64(i)
		bytebuf := bytes.NewBuffer([]byte{})
		binary.Write(bytebuf, binary.BigEndian, data)
		a := crypto.Keccak256Hash(bytebuf.Bytes()).String()
		accounts[i] = common.HexToAddress(a)
	}

	bootnode := "/ip4/127.0.0.1/tcp/54321/p2p/12D3KooWSScaw1zaPERB3mT2R3hmwavHKMiTNPGtR8dNdeZYrwNV"

	return config.NewConfig("test", topo, 4, nodeId, accounts, bootnode)
}

func NewBFTProtocolWithId(nodeId uint32) *BFTProtocol {
	conf := NewConfig(nodeId)
	eb := eventbus.New()
	p2pNode := p2p.NewProtocol(eb, conf)
	bc := core.NewBlockchain(p2pNode, conf, eb)
	bft := NewBFTProtocol(p2pNode, eb, bc, conf)
	p2pNode.Start()
	bc.Server()
	bft.Server()
	return bft
}

func NewBlockChainWithId(nodeId uint32) *core.Blockchain {
	conf := NewConfig(nodeId)
	eb := eventbus.New()
	p2pNode := p2p.NewProtocol(eb, conf)
	bc := core.NewBlockchain(p2pNode, conf, eb)
	p2pNode.Start()
	return bc
}

func TestNewBFTProtocol(t *testing.T) {
	blockchains := []*core.Blockchain{}
	for nodeId := uint32(1); nodeId <= 8; nodeId++ {
		bc := NewBlockChainWithId(nodeId)
		bc.Server()
		blockchains = append(blockchains, bc)
	}

	var bfts []*BFTProtocol
	for nodeId := uint32(9); nodeId <= 24; nodeId++ {
		bft := NewBFTProtocolWithId(nodeId)
		bft.Start()
		bfts = append(bfts, bft)
	}

	for {
		time.Sleep(time.Minute)
	}
}
