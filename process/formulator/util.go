package formulator

import (
	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/core/types"
)

func toHyperFormulator(loader types.LoaderWrapper, addr common.Address) (*FormulatorAccount, error) {
	acc, err := loader.Account(addr)
	if err != nil {
		return nil, err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return nil, types.ErrInvalidAccountType
	}
	if frAcc.FormulatorType != HyperFormulatorType {
		return nil, types.ErrInvalidAccountType
	}
	return frAcc, nil
}
