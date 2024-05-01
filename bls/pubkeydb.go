package mitosisbls

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/KyrinCode/Mitosis/config"
	"github.com/KyrinCode/Mitosis/types"
	logger "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/log"
	"github.com/herumi/bls-eth-go-binary/bls"
)

type PubKeyDB struct {
	// shardID uint32
	// Size    uint32
	DB map[uint32]*bls.PublicKey // key: nodeId
	// conf    *config.Config
	mutex sync.Mutex

	// ShardPubMsg []types.NodePub
}

func NewPubKeyDB() *PubKeyDB {
	return &PubKeyDB{
		// shardID:     shardID,
		// Size:        0,
		DB: make(map[uint32]*bls.PublicKey),
		// conf:        conf,
		// ShardPubMsg: make([]types.NodePub, 0),
	}
}

func (p *PubKeyDB) Add(nodeId uint32, pubByte []byte) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var pubKey bls.PublicKey
	addNewKey := false
	err := pubKey.Deserialize(pubByte)
	if err != nil {
		return false
	}
	if _, ok := p.DB[nodeId]; !ok {
		addNewKey = true
	}
	p.DB[nodeId] = &pubKey

	return addNewKey
}

func (p *PubKeyDB) GetAggregatePubKey(bitmap types.Bitmap, offset uint32) *bls.PublicKey {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	elements := bitmap.GetElement()
	var pubKey bls.PublicKey
	for _, i := range elements {
		if pbk, ok := p.DB[i+offset]; ok {
			pubKey.Add(pbk)
		} else {
			log.Error("PubKey %d is missing", i)
		}
	}
	return &pubKey
}

func (p *PubKeyDB) GetPubKey(nodeId uint32) *bls.PublicKey {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	ps, ok := p.DB[nodeId]
	if !ok {
		return nil
	}
	return ps
}

func (p *PubKeyDB) Reset(nodeIds []uint32) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	cert, success := LoadCert("../bls/keygen/cert/cert.json")
	if !success {
		log.Error("cert.json is missing")
		return
	}

	bls.Init(bls.BLS12_381)
	bls.SetETHmode(bls.EthModeDraft07)
	for _, nodeId := range nodeIds {
		var pubKey bls.PublicKey
		// fmt.Println(cert[nodeId])

		err := pubKey.DeserializeHexStr(cert[nodeId])
		if err != nil {
			log.Error("Deserialize public key error")
		}
		// 从cert中得到nodeId到pubKey的映射关系，赋值给pubKey
		p.DB[nodeId] = &pubKey
	}
}

func LoadCert(filename string) (map[uint32]string, bool) {
	cert := make(map[uint32]string)

	data, err := os.ReadFile(filename)
	if err != nil {
		logger.Error("Read cert.json file error")
		return cert, false
	}
	certJson := []byte(data)
	err = json.Unmarshal(certJson, &cert)
	if err != nil {
		logger.Error("unmarshal json data error")
		return cert, false
	}

	return cert, true
}

type PubKeyStore struct {
	PubKeyDBs map[uint32]*PubKeyDB // key: shardId
	// conf     *config.Config
	mutex sync.Mutex
}

func NewPubKeyStore(topo config.Topology) *PubKeyStore {
	pubKeyDBs := make(map[uint32]*PubKeyDB)
	// RShards
	for shardId := 0; shardId <= len(topo.RShardIds); shardId++ {
		pubKeyDBs[uint32(shardId)] = NewPubKeyDB()
	}
	// PShards
	for _, shardId := range topo.RShardIds {
		for _, pShardId := range topo.PShardIds[topo.RShardIds[shardId-1]] {
			pubKeyDBs[pShardId] = NewPubKeyDB()
		}
	}
	return &PubKeyStore{
		PubKeyDBs: pubKeyDBs,
		// conf:     conf,
	}

}

func (p *PubKeyStore) AddPubKey(shardId, nodeId uint32, pubByte []byte) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	pubKeyDB, ok := p.PubKeyDBs[shardId]
	if !ok {
		pubKeyDB = NewPubKeyDB()
	}
	addNewKey := pubKeyDB.Add(nodeId, pubByte)
	p.PubKeyDBs[shardId] = pubKeyDB
	return addNewKey
}

func (p *PubKeyStore) GetAggregatePubKey(shardId uint32, bitmap types.Bitmap, offset uint32) *bls.PublicKey {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	pubKeyDB, ok := p.PubKeyDBs[shardId]
	if !ok {
		return nil
	}
	return pubKeyDB.GetAggregatePubKey(bitmap, offset)
}

func (p *PubKeyStore) GetPubKey(shardId, nodeId uint32) *bls.PublicKey {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	pubKeyDB, ok := p.PubKeyDBs[shardId]
	if !ok {
		return nil
	}
	return pubKeyDB.GetPubKey(nodeId)
}

func (p *PubKeyStore) Reset(nodes map[uint32][]uint32) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// 直接遍历nodes map
	for shardId, nodeIds := range nodes {
		pubKeyDB, ok := p.PubKeyDBs[shardId]
		if !ok {
			pubKeyDB = NewPubKeyDB()

		}
		pubKeyDB.Reset(nodeIds)
		p.PubKeyDBs[shardId] = pubKeyDB
	}
}
