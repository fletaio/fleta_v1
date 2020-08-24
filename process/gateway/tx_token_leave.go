package gateway

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/common/amount"
	"github.com/fletaio/fleta_v1/common/hash"
	"github.com/fletaio/fleta_v1/core/types"
	"github.com/fletaio/fleta_v1/process/admin"
)

// TokenLeave is a TokenLeave
type TokenLeave struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	CoinTXID   string
	CoinFrom   common.Address
	ERC20TXID  hash.Hash256
	ERC20To    ERC20Address
	Amount     *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *TokenLeave) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *TokenLeave) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *TokenLeave) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *TokenLeave) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Gateway)

	if tx.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}
	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return types.ErrDustAmount
	}
	if _, _, err := types.ParseTransactionID(tx.CoinTXID); err != nil {
		return err
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if sp.HasOutTXID(loader, tx.CoinTXID) {
		return ErrProcessedOutTXID
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *TokenLeave) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Gateway)
	sp.setOutTXID(ctw, tx.CoinTXID)
	return nil
}

// MarshalJSON is a marshaler function
func (tx *TokenLeave) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"coin_txid":`)
	if bs, err := json.Marshal(tx.CoinTXID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"coin_from":`)
	if bs, err := tx.CoinFrom.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"erc20_txid":`)
	if bs, err := tx.ERC20TXID.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"erc20_to":`)
	if bs, err := tx.ERC20To.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
