package types

import (
	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/common/hash"
)

// Provider defines functions that loads chain data from the chain
type Provider interface {
	ChainID() uint8
	Symbol() string
	Usage() string
	Version() uint16
	Height() uint32
	LastStatus() (uint32, hash.Hash256)
	LastHash() hash.Hash256
	LastTimestamp() uint64
	Hash(height uint32) (hash.Hash256, error)
	Header(height uint32) (*Header, error)
	Block(height uint32) (*Block, error)
	Seq(addr common.Address) uint64
	Events(From uint32, To uint32) ([]Event, error)
	NewLoaderWrapper(pid uint8) LoaderWrapper
	NewAddress(height uint32, index uint16) common.Address
}
