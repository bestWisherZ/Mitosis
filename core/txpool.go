// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	crytporand "crypto/rand"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/KyrinCode/Mitosis/config"
	"github.com/KyrinCode/Mitosis/types"
)

type TxPool struct {
	PendingTxs []types.Transaction
	txMutex    sync.Mutex

	PendingInboundChunks []types.OutboundChunk
	inboundChunkMutex    sync.Mutex

	config *config.Config
}

func NewTxPool(config *config.Config) *TxPool {
	txPool := &TxPool{
		config: config,
	}
	if config.ShardId > 1000 && config.IsLeader {
		txPool.loadTestTxs()
	}
	return txPool
}

// func (pool *TxPool) takeTxs(txLen int) ([]types.Transaction, []types.OutboundChunk) {
// 	pool.inboundChunkMutex.Lock()
// 	var txList []types.Transaction
// 	var inboundChunkList []types.OutboundChunk
// 	// cross-shard txs have higher priority
// 	for _, inboundChunk := range pool.PendingInboundChunks {
// 		if txLen > 0 {
// 			txLen -= len(inboundChunk.Txs)
// 			inboundChunkList = append(inboundChunkList, inboundChunk)
// 		} else {
// 			break
// 		}
// 	}
// 	pool.PendingInboundChunks = pool.PendingInboundChunks[len(inboundChunkList):]
// 	pool.inboundChunkMutex.Unlock()
// 	if txLen <= 0 {
// 		return txList, inboundChunkList
// 	}

// 	pool.txMutex.Lock()
// 	if txLen < len(pool.PendingTxs) {
// 		txList = pool.PendingTxs[:txLen]
// 		pool.PendingTxs = pool.PendingTxs[txLen:]
// 	} else {
// 		txList = pool.PendingTxs
// 		pool.PendingTxs = []types.Transaction{}
// 	}
// 	pool.txMutex.Unlock()

// 	if len(pool.PendingTxs) < 50 {
// 		pool.AddTxs()
// 	}

// 	return txList, inboundChunkList
// }

func (pool *TxPool) takeTxs(txLen int) ([]types.Transaction, []types.OutboundChunk) {
	pool.inboundChunkMutex.Lock()
	var txList []types.Transaction
	var inboundChunkList []types.OutboundChunk
	// cross-shard txs have higher priority
	for _, inboundChunk := range pool.PendingInboundChunks {
		if txLen > 0 {
			txLen -= len(inboundChunk.Txs)
			inboundChunkList = append(inboundChunkList, inboundChunk)
		} else {
			break
		}
	}
	pool.PendingInboundChunks = pool.PendingInboundChunks[len(inboundChunkList):]
	pool.inboundChunkMutex.Unlock()
	if txLen <= 0 {
		return txList, inboundChunkList
	}

	pool.txMutex.Lock()
	if txLen < len(pool.PendingTxs) {
		txList = pool.PendingTxs[:txLen]
		pool.PendingTxs = pool.PendingTxs[txLen:]
	} else {
		txList = pool.PendingTxs
		pool.PendingTxs = []types.Transaction{}
	}
	pool.txMutex.Unlock()

	// if len(pool.PendingTxs) < 50 {
	// 	pool.AddTxs()
	// }

	return txList, inboundChunkList
}

func (pool *TxPool) AddInboundChunk(inboundChunk *types.OutboundChunk) {
	newInboundChunk := inboundChunk.Copy().(types.OutboundChunk)
	pool.inboundChunkMutex.Lock()
	pool.PendingInboundChunks = append(pool.PendingInboundChunks, newInboundChunk)
	pool.inboundChunkMutex.Unlock()
}

// func (pool *TxPool) AddTxs() {
// 	pool.txMutex.Lock()
// 	for i := 0; i < 100; i++ {
// 		pool.PendingTxs = append(pool.PendingTxs, *pool.RandTx())
// 	}
// 	pool.txMutex.Unlock()
// }

// func (pool *TxPool) RandTx() *types.Transaction {
// 	fromShard := pool.config.ShardId
// 	toShard := rand.Uint32()%pool.config.PShardNum + 1001
// 	fromIdx := rand.Uint32() % uint32(len(pool.config.Accounts))
// 	toIdx := rand.Uint32() % uint32(len(pool.config.Accounts))
// 	fromAddress := pool.config.Accounts[fromIdx]
// 	toAddress := pool.config.Accounts[toIdx]
// 	value := uint64(1)
// 	// generate random data bytes with four bytes length
// 	data := make([]byte, 4)
// 	crytporand.Read(data)
// 	return types.NewTransaction(fromShard, toShard, fromAddress, toAddress, value, data, "tx_hash_logic")
// }

// func (pool *TxPool) TestTxs() []*types.Transaction {
// 	txs := []*types.Transaction{}
// 	// 90个片内给另9个账户
// 	for i := 0; i < 100; i++ {
// 		txs = append(txs, )
// 	}
// 	return txs
// }

type TestTx struct {
	TxHashLogic string `json:"tx_hash_logic"`
	FromShard uint32 `json:"from_shard"`
	FromAddr  string `json:"from_address"`
	ToShard   uint32 `json:"to_shard"`
	// // option1: estuary
	// ToAddr    map[string]uint64 `json:"to_address"`
	// // option2: monoxide
	ToAddr string `json:"to_address"`
}

func (pool *TxPool) loadTestTxs() {
	// // option1: estuary
	// filename := "../test/estuary-dataset/estuary_shard_" + strconv.Itoa(int(pool.config.ShardId)) + "_txs.json"
	// // option2: monoxide
	filename := "../test/monoxide-dataset/monoxide_shard_" + strconv.Itoa(int(pool.config.ShardId)) + "_txs.json"
	file, err:= os.Open(filename)
	if err != nil {
		logChain.Errorf("[Node-%d-%d] error loading test transactions", pool.config.ShardId, pool.config.NodeId)
		return
	}
	defer file.Close()

	var testTxs []TestTx
	err = json.NewDecoder(file).Decode(&testTxs)
	if err != nil {
		logChain.Errorf("[Node-%d-%d] error decoding json", pool.config.ShardId, pool.config.NodeId)
		return
	}

	pool.txMutex.Lock()
	data := make([]byte, 4)
	for _, testTx := range testTxs {
		crytporand.Read(data)
		from_address := common.BytesToAddress([]byte(testTx.FromAddr))

		// // option1: estuary
		// var to_address common.Address
		// for k, _ := range testTx.ToAddr {
		// 	to_address = common.BytesToAddress([]byte(k))
		// 	break
		// }

		// option2: monoxide
		to_address := common.BytesToAddress([]byte(testTx.ToAddr))

		tx := types.NewTransaction(testTx.FromShard, testTx.ToShard, from_address, to_address, 0, data, testTx.TxHashLogic)
		pool.PendingTxs = append(pool.PendingTxs, *tx)
	}
	pool.txMutex.Unlock()
}