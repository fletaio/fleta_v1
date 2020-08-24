package payment

import (
	"github.com/fletaio/fleta_v1/common"
	"github.com/fletaio/fleta_v1/common/amount"
	"github.com/fletaio/fleta_v1/core/types"
	"github.com/fletaio/fleta_v1/encoding"
)

func (p *Payment) getRequestPayment(lw types.LoaderWrapper, TXID string) (*RequestPayment, error) {
	if bs := lw.ProcessData(toRequestPaymentKey(TXID)); len(bs) > 0 {
		tx := &RequestPayment{}
		if err := encoding.Unmarshal(bs, &tx); err != nil {
			return nil, err
		}
		return tx, nil
	} else {
		return nil, ErrNotExistRequestPayment
	}
}

func (p *Payment) addRequestPayment(ctw *types.ContextWrapper, TXID string, tx *RequestPayment) error {
	body, err := encoding.Marshal(tx)
	if err != nil {
		return err
	}
	ctw.SetProcessData(toRequestPaymentKey(TXID), body)
	return nil
}

func (p *Payment) removeRequestPayment(ctw *types.ContextWrapper, TXID string) {
	ctw.SetProcessData(toRequestPaymentKey(TXID), nil)
}

// GetTopicName returns the topic name of the topic
func (p *Payment) GetTopicName(loader types.Loader, topic uint64) (string, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.ProcessData(toTopicKey(topic)); len(bs) > 0 {
		return string(bs), nil
	} else {
		return "", ErrNotExistTopic
	}
}

func (p *Payment) addTopic(ctw *types.ContextWrapper, topic uint64, Name string) error {
	if bs := ctw.ProcessData(toTopicKey(topic)); len(bs) > 0 {
		return ErrExistTopic
	}
	ctw.SetProcessData(toTopicKey(topic), []byte(Name))
	return nil
}

func (p *Payment) removeTopic(ctw *types.ContextWrapper, topic uint64) {
	ctw.SetProcessData(toTopicKey(topic), nil)
}

func (p *Payment) getSubscribe(lw types.LoaderWrapper, topic uint64, addr common.Address) (*amount.Amount, error) {
	if bs := lw.AccountData(addr, toTopicKey(topic)); len(bs) > 0 {
		am := amount.NewAmountFromBytes(bs)
		return am, nil
	} else {
		return nil, ErrNotExistSubscribe
	}
}

func (p *Payment) addSubscribe(ctw *types.ContextWrapper, topic uint64, addr common.Address, am *amount.Amount) error {
	if bs := ctw.AccountData(addr, toTopicKey(topic)); len(bs) > 0 {
		return ErrExistSubscribe
	}
	ctw.SetAccountData(addr, toTopicKey(topic), am.Bytes())
	return nil
}

func (p *Payment) removeSubscribe(ctw *types.ContextWrapper, topic uint64, addr common.Address) {
	ctw.SetAccountData(addr, toTopicKey(topic), nil)
}
