package types

import (
	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/go-crypto"
)

// status of a validator
type BondStatus byte

// nolint
const (
	Unbonded  BondStatus = 0x00
	Unbonding BondStatus = 0x01
	Bonded    BondStatus = 0x02
)

// validator for a delegated proof of stake system
type Validator interface {
	GetStatus() BondStatus    // status of the validator
	GetOwner() Address        // owner address to receive/return validators coins
	GetPubKey() crypto.PubKey // validation pubkey
	GetPower() Rat            // validation power
	GetBondHeight() int64     // height in which the validator became active
}

// validator which fulfills wrsp validator interface for use in Tendermint
func WRSPValidator(v Validator) wrsp.Validator {
	return wrsp.Validator{
		PubKey: v.GetPubKey().Bytes(),
		Power:  v.GetPower().Evaluate(),
	}
}

// properties for the set of all validators
type ValidatorSet interface {
	// iterate through validator by owner-address, execute func for each validator
	IterateValidators(Context,
		func(index int64, validator Validator) (stop bool))

	// iterate through bonded validator by pubkey-address, execute func for each validator
	IterateValidatorsBonded(Context,
		func(index int64, validator Validator) (stop bool))

	Validator(Context, Address) Validator // get a particular validator by owner address
	TotalPower(Context) Rat               // total power of the validator set
}

//_______________________________________________________________________________

// delegation bond for a delegated proof of stake system
type Delegation interface {
	GetDelegator() Address // delegator address for the bond
	GetValidator() Address // validator owner address for the bond
	GetBondShares() Rat    // amount of validator's shares
}

// properties for the set of all delegations for a particular
type DelegationSet interface {

	// iterate through all delegations from one delegator by validator-address,
	//   execute func for each validator
	IterateDelegators(Context, delegator Address,
		fn func(index int64, delegation Delegation) (stop bool))
}
