package common

import (
	"bytes"
	"strings"

	"github.com/fletaio/fleta/common/binutil"
	"github.com/mr-tron/base58/base58"
)

// AddressSize is 14 bytes
const AddressSize = 14

// Address is the [AddressSize]byte with methods
type Address [AddressSize]byte

// NewAddress returns a Address by the AccountCoordinate and the magic
func NewAddress(height uint32, index uint16, magic uint64) Address {
	var addr Address
	binutil.BigEndian.PutUint32(addr[:], height)
	binutil.BigEndian.PutUint16(addr[4:], index)
	binutil.BigEndian.PutUint64(addr[6:], magic)
	return addr
}

// MarshalJSON is a marshaler function
func (addr Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + addr.String() + `"`), nil
}

// UnmarshalJSON is a unmarshaler function
func (addr *Address) UnmarshalJSON(bs []byte) error {
	if len(bs) < 3 {
		return ErrInvalidAddressFormat
	}
	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		return ErrInvalidAddressFormat
	}
	v, err := ParseAddress(string(bs[1 : len(bs)-1]))
	if err != nil {
		return err
	}
	copy(addr[:], v[:])
	return nil
}

// String returns a base58 value of the address
func (addr Address) String() string {
	var bs []byte
	checksum := addr.Checksum()
	result := bytes.TrimRight(addr[:], string([]byte{0}))
	if len(result) < 7 {
		bs = make([]byte, 7)
		copy(bs[1:], result[:])
	} else if len(result) < 15 {
		bs = make([]byte, 15)
		copy(bs[1:], result[:])
	}
	bs[0] = checksum

	base := addr[6:8]
	if base[0] == 0 && base[1] == 0 {
		return base58.Encode(bs)
	} else {
		tbs := addr[8:]
		for i := 0; i < len(tbs); i += 2 {
			tbs[i] = tbs[i] ^ base[0]
			tbs[i+1] = tbs[i+1] ^ base[1]
		}
		for i := 0; i < len(tbs); i++ {
			if tbs[i] == 0 {
				return string(tbs[:i]) + "_" + base58.Encode(bs[:9])
			}
		}
		return string(tbs) + "_" + base58.Encode(bs[:9])
	}
}

// Clone returns the clonend value of it
func (addr Address) Clone() Address {
	var cp Address
	copy(cp[:], addr[:])
	return cp
}

// Checksum returns the checksum byte
func (addr Address) Checksum() byte {
	var cs byte
	for _, c := range addr {
		cs = cs ^ c
	}
	return cs
}

// Height returns the height of the address created
func (addr Address) Height() uint32 {
	return binutil.BigEndian.Uint32(addr[:])
}

// Index returns the index of the address created
func (addr Address) Index() uint16 {
	return binutil.BigEndian.Uint16(addr[4:])
}

// Nonce returns the nonce of the address created
func (addr Address) Nonce() uint64 {
	return binutil.BigEndian.Uint64(addr[6:])
}

// ParseAddress parse the address from the string
func ParseAddress(str string) (Address, error) {
	ls := strings.SplitN(str, "_", 2)
	var symbol string
	if len(ls) > 1 {
		symbol = ls[0]
		if len(str) < len(symbol) {
			return Address{}, ErrInvalidAddressFormat
		}
		str = str[len(symbol)+1:]
	}
	bs, err := base58.Decode(str)
	if err != nil {
		return Address{}, err
	}
	var base []byte
	if len(symbol) > 0 {
		if len(bs) != 9 {
			return Address{}, ErrInvalidAddressFormat
		}
		base = bs[7:]
	} else {
		if len(bs) != 7 {
			return Address{}, ErrInvalidAddressFormat
		}
	}
	cs := bs[0]
	var addr Address
	copy(addr[:], bs[1:])
	if len(symbol) > 0 {
		copy(addr[6:], base)
		tbs := make([]byte, 6)
		if base[0] != 0 || base[1] != 0 {
			copy(tbs, []byte(symbol))
		}
		for i := 0; i < len(tbs); i += 2 {
			tbs[i] = tbs[i] ^ base[0]
			tbs[i+1] = tbs[i+1] ^ base[1]
		}
		copy(addr[8:], tbs)
	}
	if cs != addr.Checksum() {
		return Address{}, ErrInvalidAddressCheckSum
	}
	return addr, nil
}

// MustParseAddress panic when error occurred
func MustParseAddress(str string) Address {
	addr, err := ParseAddress(str)
	if err != nil {
		panic(err)
	}
	return addr
}
