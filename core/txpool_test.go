package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/KyrinCode/Mitosis/config"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewConfig2(nodeId uint32) *config.Config {
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

	bootnode := "/ip4/127.0.0.1/tcp/50467/p2p/12D3KooWJLHcagKsPPf6nWtXizRTmAgsZJGrDAWP1xFFmUCT7bvx"

	return config.NewConfig("test", topo, 4, nodeId, accounts, bootnode)
}

func NewTxPoolWithId(nodeId uint32) *TxPool {
	conf := NewConfig2(nodeId)
	txPool := NewTxPool(conf)
	return txPool
}

func TestTxPool(t *testing.T) {
	txPool := NewTxPoolWithId(12)
	for _, testTx := range txPool.PendingTxs {
		fmt.Println(testTx)
	}
}