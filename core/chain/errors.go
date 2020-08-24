package chain

import "errors"

// errors
var (
	ErrExistServiceName             = errors.New("exist service name")
	ErrExistServiceID               = errors.New("exist service id")
	ErrNotExistService              = errors.New("not exist service")
	ErrInvalidChainID               = errors.New("invalid chain id")
	ErrInvalidVersion               = errors.New("invalid version")
	ErrInvalidHeight                = errors.New("invalid height")
	ErrInvalidPrevHash              = errors.New("invalid prev hash")
	ErrInvalidContextHash           = errors.New("invalid context hash")
	ErrInvalidLevelRootHash         = errors.New("invalid level root hash")
	ErrInvalidTimestamp             = errors.New("invalid timestamp")
	ErrInvalidGenerator             = errors.New("invalid generator")
	ErrExceedHashCount              = errors.New("exceed hash count")
	ErrInvalidHashCount             = errors.New("invalid hash count")
	ErrInvalidGenesisHash           = errors.New("invalid genesis hash")
	ErrInvalidTxInKey               = errors.New("invalid txin key")
	ErrInvalidResult                = errors.New("invalid result")
	ErrChainClosed                  = errors.New("chain closed")
	ErrStoreClosed                  = errors.New("store closed")
	ErrAlreadyGenesised             = errors.New("already genesised")
	ErrDirtyContext                 = errors.New("dirty context")
	ErrReservedID                   = errors.New("reserved id")
	ErrAddBeforeChainInit           = errors.New("add before chain init")
	ErrApplicationIDMustBe255       = errors.New("application id must be 255")
	ErrFoundForkedBlock             = errors.New("found forked block")
	ErrCannotDeleteGeneratorAccount = errors.New("cannot delete generator account")
	ErrInvalidAccountName           = errors.New("invalid account name")
)
