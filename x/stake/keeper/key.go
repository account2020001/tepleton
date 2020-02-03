package keeper

import (
	"encoding/binary"

	"github.com/tepleton/tepleton/crypto"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	"github.com/tepleton/tepleton-sdk/x/stake/types"
)

// TODO remove some of these prefixes once have working multistore

//nolint
var (
	// Keys for store prefixes
	ParamKey                         = []byte{0x00} // key for parameters relating to staking
	PoolKey                          = []byte{0x01} // key for the staking pools
	ValidatorsKey                    = []byte{0x02} // prefix for each key to a validator
	ValidatorsByPubKeyIndexKey       = []byte{0x03} // prefix for each key to a validator index, by pubkey
	ValidatorsBondedIndexKey         = []byte{0x04} // prefix for each key to a validator index, for bonded validators
	ValidatorsByPowerIndexKey        = []byte{0x05} // prefix for each key to a validator index, sorted by power
	ValidatorCliffIndexKey           = []byte{0x06} // key for the validator index of the cliff validator
	ValidatorPowerCliffKey           = []byte{0x07} // key for the power of the validator on the cliff
	TendermintUpdatesKey             = []byte{0x08} // prefix for each key to a validator which is being updated
	IntraTxCounterKey                = []byte{0x09} // key for intra-block tx index
	DelegationKey                    = []byte{0x0A} // key for a delegation
	UnbondingDelegationKey           = []byte{0x0B} // key for an unbonding-delegation
	UnbondingDelegationByValIndexKey = []byte{0x0C} // prefix for each key for an unbonding-delegation, by validator owner
	RedelegationKey                  = []byte{0x0D} // key for a redelegation
	RedelegationByValSrcIndexKey     = []byte{0x0E} // prefix for each key for an redelegation, by validator owner
	RedelegationByValDstIndexKey     = []byte{0x0F} // prefix for each key for an redelegation, by validator owner
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// get the key for the validator with address.
// the value at this key is of type stake/types.Validator
func GetValidatorKey(ownerAddr sdk.Address) []byte {
	return append(ValidatorsKey, ownerAddr.Bytes()...)
}

// get the key for the validator with pubkey.
// the value at this key should the address for a stake/types.Validator
func GetValidatorByPubKeyIndexKey(pubkey crypto.PubKey) []byte {
	return append(ValidatorsByPubKeyIndexKey, pubkey.Bytes()...)
}

// get the key for the current validator group, ordered like tepleton.
// the value at this key is the address of the owner of a validator
func GetValidatorsBondedIndexKey(ownerAddr sdk.Address) []byte {
	return append(ValidatorsBondedIndexKey, ownerAddr.Bytes()...)
}

// get the validator by power index. power index is the key used in the power-store,
// and represents the relative power ranking of the validator.
// the value at this key is of type address, the address being the Address
// of the corresponding validator.
func GetValidatorsByPowerIndexKey(validator types.Validator, pool types.Pool) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getValidatorPowerRank(validator, pool)
}

// get the power ranking of a validator
func getValidatorPowerRank(validator types.Validator, pool types.Pool) []byte {

	power := validator.EquivalentBondedShares(pool)
	powerBytes := []byte(power.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	// TODO ensure that the key will be a readable string.. probably should add seperators and have
	revokedBytes := make([]byte, 1)
	if validator.Revoked {
		revokedBytes[0] = byte(0x01)
	} else {
		revokedBytes[0] = byte(0x00)
	}

	// TODO ensure that the key will be a readable string.. probably should add separators and have
	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.BondHeight)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.BondIntraTxCounter)) // invert counter (first txns have priority)

	return append(ValidatorsByPowerIndexKey,
		append(revokedBytes,
			append(powerBytes,
				append(heightBytes, counterBytes...)...)...)...)
}

// get the key for the accumulated update validators.
// The value at this key is of type stake/types.Validator
func GetTendermintUpdatesKey(ownerAddr sdk.Address) []byte {
	return append(TendermintUpdatesKey, ownerAddr.Bytes()...)
}

//________________________________________________________________________________

// get the key for delegator bond with validator.
// The value at this key is of type stake/types.Delegation
func GetDelegationKey(delegatorAddr, validatorAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetDelegationsKey(delegatorAddr, cdc), validatorAddr.Bytes()...)
}

// get the prefix for a delegator for all validators
func GetDelegationsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res := cdc.MustMarshalBinary(&delegatorAddr)
	return append(DelegationKey, res...)
}

//________________________________________________________________________________

// get the key for an unbonding delegation by delegator and validator addr.
// The value at this key is of type stake/types.UnbondingDelegation
func GetUBDKey(delegatorAddr, validatorAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetUBDsKey(delegatorAddr, cdc), validatorAddr.Bytes()...)
}

// get the index-key for an unbonding delegation, stored by validator-index
// The value at this key is a key for the corresponding unbonding delegation.
func GetUBDByValIndexKey(delegatorAddr, validatorAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetUBDsByValIndexKey(validatorAddr, cdc), delegatorAddr.Bytes()...)
}

//______________

// get the prefix for all unbonding delegations from a delegator
func GetUBDsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res := cdc.MustMarshalBinary(&delegatorAddr)
	return append(UnbondingDelegationKey, res...)
}

// get the prefix keyspace for the indexes of unbonding delegations for a validator
func GetUBDsByValIndexKey(validatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res := cdc.MustMarshalBinary(&validatorAddr)
	return append(UnbondingDelegationByValIndexKey, res...)
}

//________________________________________________________________________________

// get the key for a redelegation
// The value at this key is of type stake/types.RedelegationKey
func GetREDKey(delegatorAddr, validatorSrcAddr,
	validatorDstAddr sdk.Address, cdc *wire.Codec) []byte {

	return append(
		GetREDsKey(delegatorAddr, cdc),
		append(
			validatorSrcAddr.Bytes(),
			validatorDstAddr.Bytes()...)...,
	)
}

// get the index-key for a redelegation, stored by source-validator-index
// The value at this key is a key for the corresponding redelegation.
func GetREDByValSrcIndexKey(delegatorAddr, validatorSrcAddr,
	validatorDstAddr sdk.Address, cdc *wire.Codec) []byte {

	return append(
		GetREDsFromValSrcIndexKey(validatorSrcAddr, cdc),
		append(
			delegatorAddr.Bytes(),
			validatorDstAddr.Bytes()...)...,
	)
}

// get the index-key for a redelegation, stored by destination-validator-index
// The value at this key is a key for the corresponding redelegation.
func GetREDByValDstIndexKey(delegatorAddr, validatorSrcAddr,
	validatorDstAddr sdk.Address, cdc *wire.Codec) []byte {

	return append(
		GetREDsToValDstIndexKey(validatorDstAddr, cdc),
		append(
			delegatorAddr.Bytes(),
			validatorSrcAddr.Bytes()...)...,
	)
}

//______________

// get the prefix keyspace for redelegations from a delegator
func GetREDsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res := cdc.MustMarshalBinary(&delegatorAddr)
	return append(RedelegationKey, res...)
}

// get the prefix keyspace for all redelegations redelegating away from a source validator
func GetREDsFromValSrcIndexKey(validatorSrcAddr sdk.Address, cdc *wire.Codec) []byte {
	res := cdc.MustMarshalBinary(&validatorSrcAddr)
	return append(RedelegationByValSrcIndexKey, res...)
}

// get the prefix keyspace for all redelegations redelegating towards a destination validator
func GetREDsToValDstIndexKey(validatorDstAddr sdk.Address, cdc *wire.Codec) []byte {
	res := cdc.MustMarshalBinary(&validatorDstAddr)
	return append(RedelegationByValDstIndexKey, res...)
}

// get the prefix keyspace for all redelegations redelegating towards a destination validator
// from a particular delegator
func GetREDsByDelToValDstIndexKey(delegatorAddr sdk.Address,
	validatorDstAddr sdk.Address, cdc *wire.Codec) []byte {

	return append(
		GetREDsToValDstIndexKey(validatorDstAddr, cdc),
		delegatorAddr.Bytes()...)
}
