package app

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/modules/coin"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/txs"
	"github.com/tepleton/basecoin/types"
	wire "github.com/tepleton/go-wire"
	eyes "github.com/tepleton/merkleeyes/client"
	"github.com/tepleton/tmlibs/log"
)

//--------------------------------------------------------
// test environment is a list of input and output accounts

type appTest struct {
	t       *testing.T
	chainID string
	app     *Basecoin
	acctIn  *coin.AccountWithKey
	acctOut *coin.AccountWithKey
}

func newAppTest(t *testing.T) *appTest {
	at := &appTest{
		t:       t,
		chainID: "test_chain_id",
	}
	at.reset()
	return at
}

// make a tx sending 5mycoin from each acctIn to acctOut
func (at *appTest) getTx(seq int, coins types.Coins) basecoin.Tx {
	in := []coin.TxInput{{Address: at.acctIn.Actor(), Coins: coins, Sequence: seq}}
	out := []coin.TxOutput{{Address: at.acctOut.Actor(), Coins: coins}}
	tx := coin.NewSendTx(in, out)
	tx = txs.NewChain(at.chainID, tx)
	stx := txs.NewMulti(tx)
	txs.Sign(stx, at.acctIn.Key)
	return stx.Wrap()
}

// set the account on the app through SetOption
func (at *appTest) initAccount(acct *coin.AccountWithKey) {
	res := at.app.SetOption("coin/account", acct.MakeOption())
	require.EqualValues(at.t, res, "Success")
}

// reset the in and out accs to be one account each with 7mycoin
func (at *appTest) reset() {
	at.acctIn = coin.NewAccountWithKey(types.Coins{{"mycoin", 7}})
	at.acctOut = coin.NewAccountWithKey(types.Coins{{"mycoin", 7}})

	eyesCli := eyes.NewLocalClient("", 0)
	// logger := log.TestingLogger().With("module", "app"),
	logger := log.NewTMLogger(os.Stdout).With("module", "app")
	logger = log.NewTracingLogger(logger)
	at.app = NewBasecoin(
		DefaultHandler(),
		eyesCli,
		logger,
	)

	res := at.app.SetOption("base/chain_id", at.chainID)
	require.EqualValues(at.t, res, "Success")

	at.initAccount(at.acctIn)
	at.initAccount(at.acctOut)

	reswrsp := at.app.Commit()
	require.True(at.t, reswrsp.IsOK(), reswrsp)
}

func getBalance(key basecoin.Actor, state types.KVStore) (types.Coins, error) {
	acct, err := coin.NewAccountant("").GetAccount(state, key)
	return acct.Coins, err
}

func getAddr(addr []byte, state types.KVStore) (types.Coins, error) {
	actor := stack.SigPerm(addr)
	return getBalance(actor, state)
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) exec(t *testing.T, tx basecoin.Tx, checkTx bool) (res wrsp.Result, diffIn, diffOut types.Coins) {
	require := require.New(t)

	initBalIn, err := getBalance(at.acctIn.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)
	initBalOut, err := getBalance(at.acctOut.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)

	txBytes := wire.BinaryBytes(tx)
	if checkTx {
		res = at.app.CheckTx(txBytes)
	} else {
		res = at.app.DeliverTx(txBytes)
	}

	endBalIn, err := getBalance(at.acctIn.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)
	endBalOut, err := getBalance(at.acctOut.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)
	return res, endBalIn.Minus(initBalIn), endBalOut.Minus(initBalOut)
}

//--------------------------------------------------------

func TestSetOption(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	eyesCli := eyes.NewLocalClient("", 0)
	app := NewBasecoin(
		DefaultHandler(),
		eyesCli,
		log.TestingLogger().With("module", "app"),
	)

	//testing ChainID
	chainID := "testChain"
	res := app.SetOption("base/chain_id", chainID)
	assert.EqualValues(app.GetState().GetChainID(), chainID)
	assert.EqualValues(res, "Success")

	// make a nice account...
	bal := types.Coins{{"atom", 77}, {"eth", 12}}
	acct := coin.NewAccountWithKey(bal)
	res = app.SetOption("coin/account", acct.MakeOption())
	require.EqualValues(res, "Success")

	// make sure it is set correctly, with some balance
	coins, err := getBalance(acct.Actor(), app.state)
	require.Nil(err)
	assert.Equal(bal, coins)

	// let's parse an account with badly sorted coins...
	unsortAddr, err := hex.DecodeString("C471FB670E44D219EE6DF2FC284BE38793ACBCE1")
	require.Nil(err)
	unsortCoins := types.Coins{{"BTC", 789}, {"eth", 123}}
	unsortAcc := `{
  "pub_key": {
    "type": "ed25519",
    "data": "AD084F0572C116D618B36F2EB08240D1BAB4B51716CCE0E7734B89C8936DCE9A"
  },
  "coins": [
    {
      "denom": "eth",
      "amount": 123
    },
    {
      "denom": "BTC",
      "amount": 789
    }
  ]
}`
	res = app.SetOption("coin/account", unsortAcc)
	require.EqualValues(res, "Success")

	coins, err = getAddr(unsortAddr, app.state)
	require.Nil(err)
	assert.True(coins.IsValid())
	assert.Equal(unsortCoins, coins)

	res = app.SetOption("base/dslfkgjdas", "")
	assert.NotEqual(res, "Success")

	res = app.SetOption("dslfkgjdas", "")
	assert.NotEqual(res, "Success")

	res = app.SetOption("dslfkgjdas/szfdjzs", "")
	assert.NotEqual(res, "Success")

}

// Test CheckTx and DeliverTx with insufficient and sufficient balance
func TestTx(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	//Bad Balance
	at.acctIn.Coins = types.Coins{{"mycoin", 2}}
	at.initAccount(at.acctIn)
	res, _, _ := at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), true)
	assert.True(res.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res)
	res, diffIn, diffOut := at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), false)
	assert.True(res.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res)
	assert.True(diffIn.IsZero())
	assert.True(diffOut.IsZero())

	//Regular CheckTx
	at.reset()
	res, _, _ = at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), true)
	assert.True(res.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res)

	//Regular DeliverTx
	at.reset()
	amt := types.Coins{{"mycoin", 3}}
	res, diffIn, diffOut = at.exec(t, at.getTx(1, amt), false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.Equal(amt.Negative(), diffIn)
	assert.Equal(amt, diffOut)
}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	res, _, _ := at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), false)
	assert.True(res.IsOK(), "Commit, DeliverTx: Expected OK return from DeliverTx, Error: %v", res)

	resQueryPreCommit := at.app.Query(wrsp.RequestQuery{
		Path: "/account",
		Data: at.acctIn.Address(),
	})

	res = at.app.Commit()
	assert.True(res.IsOK(), res)

	resQueryPostCommit := at.app.Query(wrsp.RequestQuery{
		Path: "/account",
		Data: at.acctIn.Address(),
	})
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}

func TestSplitKey(t *testing.T) {
	assert := assert.New(t)
	prefix, suffix := splitKey("foo/bar")
	assert.EqualValues("foo", prefix)
	assert.EqualValues("bar", suffix)

	prefix, suffix = splitKey("foobar")
	assert.EqualValues("base", prefix)
	assert.EqualValues("foobar", suffix)

	prefix, suffix = splitKey("some/complex/issue")
	assert.EqualValues("some", prefix)
	assert.EqualValues("complex/issue", suffix)

}
