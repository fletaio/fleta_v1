package vault

import (
	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/common/amount"
	"github.com/fletaio/fleta_v1/core/types"
)

type FeeTransaction interface {
	From() common.Address
	Fee(p types.Process, lw types.LoaderWrapper) *amount.Amount
}
