package chain

import "github.com/fletaio/fleta_v1/common"

// AccountTransaction defines common functions of account model based transactions
type AccountTransaction interface {
	Seq() uint64
	From() common.Address
}
