package counter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/app"
	"github.com/tepleton/basecoin/modules/coin"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/types"
	"github.com/tepleton/go-wire"
	eyescli "github.com/tepleton/merkleeyes/client"
	"github.com/tepleton/tmlibs/log"
)

// TODO: actually handle the counter here...
func CounterHandler() basecoin.Handler {
	// use the default stack
	h := coin.NewHandler()
	return stack.NewDefault().Use(h)
}

func TestCounterPlugin(t *testing.T) {
	assert := assert.New(t)

	// Basecoin initialization
	eyesCli := eyescli.NewLocalClient("", 0)
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(CounterHandler(), eyesCli,
		log.TestingLogger().With("module", "app"))
	bcApp.SetOption("base/chain_id", chainID)
	// t.Log(bcApp.Info())

	// Account initialization
	test1PrivAcc := types.PrivAccountFromSecret("test1")

	// Seed Basecoin with account
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1000}, {"gold", 1000}}
	accOpt, err := json.Marshal(test1Acc)
	require.Nil(t, err)
	bcApp.SetOption("base/account", string(accOpt))

	// Deliver a CounterTx
	DeliverCounterTx := func(gas int64, fee types.Coin, inputCoins types.Coins, inputSequence int, appFee types.Coins) wrsp.Result {
		// Construct an AppTx signature
		tx := &types.AppTx{
			Gas:   gas,
			Fee:   fee,
			Name:  "counter",
			Input: types.NewTxInput(test1Acc.PubKey, inputCoins, inputSequence),
			Data:  wire.BinaryBytes(CounterTx{Valid: true, Fee: appFee}),
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		// t.Logf("Sign bytes: %X\n", signBytes)
		tx.Input.Signature = test1PrivAcc.Sign(signBytes)
		// t.Logf("Signed TX bytes: %X\n", wire.BinaryBytes(struct{ types.Tx }{tx}))

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		return bcApp.DeliverTx(txBytes)
	}

	// REF: DeliverCounterTx(gas, fee, inputCoins, inputSequence, appFee) {

	// Test a basic send, no fee
	res := DeliverCounterTx(0, types.Coin{}, types.Coins{{"", 1}}, 1, types.Coins{})
	assert.True(res.IsOK(), res.String())

	// Test fee prevented transaction
	res = DeliverCounterTx(0, types.Coin{"", 2}, types.Coins{{"", 1}}, 2, types.Coins{})
	assert.True(res.IsErr(), res.String())

	// Test input equals fee
	res = DeliverCounterTx(0, types.Coin{"", 2}, types.Coins{{"", 2}}, 2, types.Coins{})
	assert.True(res.IsOK(), res.String())

	// Test more input than fee
	res = DeliverCounterTx(0, types.Coin{"", 2}, types.Coins{{"", 3}}, 3, types.Coins{})
	assert.True(res.IsOK(), res.String())

	// Test input equals fee+appFee
	res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 3}, {"gold", 1}}, 4, types.Coins{{"", 2}, {"gold", 1}})
	assert.True(res.IsOK(), res.String())

	// Test fee+appFee prevented transaction, not enough ""
	res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 2}, {"gold", 1}}, 5, types.Coins{{"", 2}, {"gold", 1}})
	assert.True(res.IsErr(), res.String())

	// Test fee+appFee prevented transaction, not enough "gold"
	res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 3}, {"gold", 1}}, 5, types.Coins{{"", 2}, {"gold", 2}})
	assert.True(res.IsErr(), res.String())

	// Test more input than fee, more ""
	res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 4}, {"gold", 1}}, 6, types.Coins{{"", 2}, {"gold", 1}})
	assert.True(res.IsOK(), res.String())

	// Test more input than fee, more "gold"
	res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 3}, {"gold", 2}}, 7, types.Coins{{"", 2}, {"gold", 1}})
	assert.True(res.IsOK(), res.String())

	// REF: DeliverCounterTx(gas, fee, inputCoins, inputSequence, appFee) {w

}
