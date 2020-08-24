package chain

import (
	"bytes"

	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

const hashPerLevel = 16
const levelHashAppender = "fletablockchain"

// hash16 returns Hash(x1,'f',x2,'l',...,x16)
func hash16(hashes []hash.Hash256) (hash.Hash256, error) {
	if len(hashes) > hashPerLevel {
		return hash.Hash256{}, ErrExceedHashCount
	}

	var buffer bytes.Buffer
	var EmptyHash hash.Hash256
	for i := 0; i < hashPerLevel; i++ {
		if i < len(hashes) {
			buffer.Write(hashes[i][:])
		} else {
			buffer.Write(EmptyHash[:])
		}
		if i < len(levelHashAppender) {
			if err := buffer.WriteByte(byte(levelHashAppender[i])); err != nil {
				return hash.Hash256{}, err
			}
		}
	}
	return hash.DoubleHash(buffer.Bytes()), nil
}

func buildLevel(hashes []hash.Hash256) ([]hash.Hash256, error) {
	LvCnt := len(hashes) / hashPerLevel
	if len(hashes)%hashPerLevel != 0 {
		LvCnt++
	}

	LvHashes := make([]hash.Hash256, LvCnt)
	for i := 0; i < LvCnt; i++ {
		last := (i + 1) * hashPerLevel
		if last > len(hashes) {
			last = len(hashes)
		}
		h, err := hash16(hashes[i*hashPerLevel : last])
		if err != nil {
			return nil, err
		}
		LvHashes[i] = h
	}
	return LvHashes, nil
}

// BuildLevelRoot returns the level root hash
func BuildLevelRoot(hashes []hash.Hash256) (hash.Hash256, error) {
	if len(hashes) > 65536 {
		return hash.Hash256{}, ErrExceedHashCount
	}
	if len(hashes) == 0 {
		return hash.Hash256{}, ErrInvalidHashCount
	}

	lv3, err := buildLevel(hashes)
	if err != nil {
		return hash.Hash256{}, err
	}
	lv2, err := buildLevel(lv3)
	if err != nil {
		return hash.Hash256{}, err
	}
	lv1, err := buildLevel(lv2)
	if err != nil {
		return hash.Hash256{}, err
	}
	h, err := hash16(lv1)
	if err != nil {
		return hash.Hash256{}, err
	}
	return h, nil
}

// HashTransaction returns the hash of the transaction
func HashTransaction(ChainID uint8, tx types.Transaction) hash.Hash256 {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		panic(err)
	}
	return HashTransactionByType(ChainID, t, tx)
}

// HashTransactionByType returns the hash of the transaction using the type
func HashTransactionByType(ChainID uint8, t uint16, tx types.Transaction) hash.Hash256 {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint8(ChainID); err != nil {
		panic(err)
	}
	if err := enc.EncodeUint16(t); err != nil {
		panic(err)
	}
	if err := enc.Encode(tx); err != nil {
		panic(err)
	}
	return hash.Hash(buffer.Bytes())
}

func isCapitalAndNumber(Name string) bool {
	for i := 0; i < len(Name); i++ {
		c := Name[i]
		if (c < '0' || '9' < c) && (c < 'A' || 'Z' < c) {
			return false
		}
	}
	return true
}

func isAlphabetAndNumber(Name string) bool {
	for i := 0; i < len(Name); i++ {
		c := Name[i]
		if (c < '0' || '9' < c) && (c < 'a' || 'z' < c) && (c < 'A' || 'Z' < c) {
			return false
		}
	}
	return true
}
