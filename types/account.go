package types

import (
	"encoding/hex"
	"errors"
	"fmt"

	crypto "github.com/tepleton/go-crypto"
	"github.com/tepleton/tmlibs/bech32"
	cmn "github.com/tepleton/tmlibs/common"
)

//Address is a go crypto-style Address
type Address = cmn.HexBytes

// Bech32 prefixes
const (
	Bech32PrefixAccAddr = "tepletonaccaddr"
	Bech32PrefixAccPub  = "tepletonaccpub"
	Bech32PrefixValAddr = "tepletonvaladdr"
	Bech32PrefixValPub  = "tepletonvalpub"
)

// Bech32ifyAcc takes Address and returns the bech32 encoded string
func Bech32ifyAcc(addr Address) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixAccAddr, addr.Bytes())
}

// MustBech32ifyAcc panics on bech32-encoding failure
func MustBech32ifyAcc(addr Address) string {
	enc, err := Bech32ifyAcc(addr)
	if err != nil {
		panic(err)
	}
	return enc
}

// Bech32ifyAccPub takes AccountPubKey and returns the bech32 encoded string
func Bech32ifyAccPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixAccPub, pub.Bytes())
}

// MustBech32ifyAccPub panics on bech32-encoding failure
func MustBech32ifyAccPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyAccPub(pub)
	if err != nil {
		panic(err)
	}
	return enc
}

// Bech32ifyVal returns the bech32 encoded string for a validator address
func Bech32ifyVal(addr Address) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixValAddr, addr.Bytes())
}

// MustBech32ifyVal panics on bech32-encoding failure
func MustBech32ifyVal(addr Address) string {
	enc, err := Bech32ifyVal(addr)
	if err != nil {
		panic(err)
	}
	return enc
}

// Bech32ifyValPub returns the bech32 encoded string for a validator pubkey
func Bech32ifyValPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixValPub, pub.Bytes())
}

// MustBech32ifyValPub pancis on bech32-encoding failure
func MustBech32ifyValPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyValPub(pub)
	if err != nil {
		panic(err)
	}
	return enc
}

// create an Address from a string
func GetAccAddressHex(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a string
func GetAccAddressBech32(address string) (addr Address, err error) {
	bz, err := getFromBech32(address, Bech32PrefixAccAddr)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// must create an Address from a string
func MustGetAccAddressBech32(address string) Address {
	addr, err := GetAccAddressBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}

// create a Pubkey from a string
func GetAccPubKeyBech32(address string) (pk crypto.PubKey, err error) {
	bz, err := getFromBech32(address, Bech32PrefixAccPub)
	if err != nil {
		return nil, err
	}

	pk, err = crypto.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// must create a Pubkey from a string
func MustGetAccPubkeyBec32(address string) crypto.PubKey {
	pk, err := GetAccPubKeyBech32(address)
	if err != nil {
		panic(err)
	}
	return pk
}

// create an Address from a hex string
func GetValAddressHex(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a bech32 string
func GetValAddressBech32(address string) (addr Address, err error) {
	bz, err := getFromBech32(address, Bech32PrefixValAddr)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// must create an Address from a bech32 string
func MustGetValAddressBech32(address string) Address {
	addr, err := GetValAddressBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}

// decode a validator public key into a PubKey
func GetValPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	bz, err := getFromBech32(pubkey, Bech32PrefixValPub)
	if err != nil {
		return nil, err
	}

	pk, err = crypto.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// must decode a validator public key into a PubKey
func MustGetValPubKeyBech32(pubkey string) crypto.PubKey {
	pk, err := GetValPubKeyBech32(pubkey)
	if err != nil {
		panic(err)
	}
	return pk
}

func getFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("must provide non-empty string")
	}
	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("Invalid bech32 prefix. Expected %s, Got %s", prefix, hrp)
	}

	return bz, nil
}
