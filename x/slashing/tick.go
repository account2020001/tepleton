package slashing

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/tepleton/tepleton-sdk/types"
	wrsp "github.com/tepleton/tepleton/wrsp/types"
	tmtypes "github.com/tepleton/tepleton/types"
)

// slashing begin block functionality
func BeginBlocker(ctx sdk.Context, req wrsp.RequestBeginBlock, sk Keeper) (tags sdk.Tags) {
	// Tag the height
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(req.Header.Height))
	tags = sdk.NewTags("height", heightBytes)

	// Iterate over all the validators  which *should* have signed this block
	// Store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	for _, signingValidator := range req.Validators {
		present := signingValidator.SignedLastBlock
		pubkey, err := tmtypes.PB2TM.PubKey(signingValidator.Validator.PubKey)
		if err != nil {
			panic(err)
		}
		sk.handleValidatorSignature(ctx, pubkey, signingValidator.Validator.Power, present)
	}

	// Iterate through any newly discovered evidence of infraction
	// Slash any validators (and since-unbonded stake within the unbonding period)
	// who contributed to valid infractions
	for _, evidence := range req.ByzantineValidators {
		pk, err := tmtypes.PB2TM.PubKey(evidence.Validator.PubKey)
		if err != nil {
			panic(err)
		}
		switch evidence.Type {
		case tmtypes.WRSPEvidenceTypeDuplicateVote:
			sk.handleDoubleSign(ctx, pk, evidence.Height, evidence.Time, evidence.Validator.Power)
		default:
			ctx.Logger().With("module", "x/slashing").Error(fmt.Sprintf("ignored unknown evidence type: %s", evidence.Type))
		}
	}

	return
}
