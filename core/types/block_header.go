package types

import (
	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/common/hash"
)

// Header is validation informations
type Header struct {
	ChainID       uint8
	Version       uint16
	Height        uint32
	PrevHash      hash.Hash256
	LevelRootHash hash.Hash256
	ContextHash   hash.Hash256
	Timestamp     uint64
	Generator     common.Address
	ConsensusData []byte
}
