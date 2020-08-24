package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/core/types"
)

// UpdateValidatorPolicy is used to update validator policy of the hyper formulator
type UpdateValidatorPolicy struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Policy     *ValidatorPolicy
}

// Timestamp returns the timestamp of the transaction
func (tx *UpdateValidatorPolicy) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *UpdateValidatorPolicy) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *UpdateValidatorPolicy) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *UpdateValidatorPolicy) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Policy == nil {
		return ErrInvalidValidatorPolicy
	}
	if tx.Policy.CommissionRatio1000 >= 1000 {
		return ErrInvalidValidatorPolicy
	}
	if tx.Policy.PayOutInterval <= 0 {
		return ErrInvalidValidatorPolicy
	}
	if tx.Policy.PayOutInterval > 30 {
		return ErrInvalidValidatorPolicy
	}

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	acc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if frAcc.FormulatorType != HyperFormulatorType {
		return ErrNotHyperFormulator
	}
	if err := frAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *UpdateValidatorPolicy) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	acc, err := ctw.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc := acc.(*FormulatorAccount)
	frAcc.Policy = tx.Policy
	return nil
}

// MarshalJSON is a marshaler function
func (tx *UpdateValidatorPolicy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"seq":`)
	if bs, err := json.Marshal(tx.Seq_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"from":`)
	if bs, err := tx.From_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"policy":`)
	if bs, err := tx.Policy.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
