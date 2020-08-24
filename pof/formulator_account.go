package pof

import (
	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/core/types"
)

type FormulatorAccount interface {
	types.Account
	IsFormulator() bool
	GeneratorHash() common.PublicHash
	IsActivated() bool
}
