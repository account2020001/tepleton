package app

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/modules/auth"
	"github.com/tepleton/basecoin/modules/base"
	"github.com/tepleton/basecoin/modules/coin"
	"github.com/tepleton/basecoin/modules/fee"
	"github.com/tepleton/basecoin/modules/nonce"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/state"
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

// baseTx is the
func (at *appTest) baseTx(coins coin.Coins) basecoin.Tx {
	in := []coin.TxInput{{Address: at.acctIn.Actor(), Coins: coins}}
	out := []coin.TxOutput{{Address: at.acctOut.Actor(), Coins: coins}}
	tx := coin.NewSendTx(in, out)
	return tx
}

func (at *appTest) signTx(tx basecoin.Tx) basecoin.Tx {
	stx := auth.NewMulti(tx)
	auth.Sign(stx, at.acctIn.Key)
	return stx.Wrap()
}

func (at *appTest) getTx(coins coin.Coins) basecoin.Tx {
	tx := at.baseTx(coins)
	tx = base.NewChainTx(at.chainID, 0, tx)
	tx = nonce.NewTx(1, []basecoin.Actor{at.acctIn.Actor()}, tx)
	return at.signTx(tx)
}

func (at *appTest) feeTx(coins coin.Coins, toll coin.Coin) basecoin.Tx {
	tx := at.baseTx(coins)
	tx = fee.NewFee(tx, toll, at.acctIn.Actor())
	tx = base.NewChainTx(at.chainID, 0, tx)
	tx = nonce.NewTx(1, []basecoin.Actor{at.acctIn.Actor()}, tx)
	return at.signTx(tx)
}

// set the account on the app through SetOption
func (at *appTest) initAccount(acct *coin.AccountWithKey) {
	res := at.app.SetOption("coin/account", acct.MakeOption())
	require.EqualValues(at.t, res, "Success")
}

// reset the in and out accs to be one account each with 7mycoin
func (at *appTest) reset() {
	at.acctIn = coin.NewAccountWithKey(coin.Coins{{"mycoin", 7}})
	at.acctOut = coin.NewAccountWithKey(coin.Coins{{"mycoin", 7}})

	eyesCli := eyes.NewLocalClient("", 0)
	// logger := log.TestingLogger().With("module", "app"),
	logger := log.NewTMLogger(os.Stdout).With("module", "app")
	logger = log.NewTracingLogger(logger)
	at.app = NewBasecoin(
		DefaultHandler("mycoin"),
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

func getBalance(key basecoin.Actor, store state.KVStore) (coin.Coins, error) {
	cspace := stack.PrefixedStore(coin.NameCoin, store)
	acct, err := coin.GetAccount(cspace, key)
	return acct.Coins, err
}

func getAddr(addr []byte, state state.KVStore) (coin.Coins, error) {
	actor := auth.SigPerm(addr)
	return getBalance(actor, state)
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) exec(t *testing.T, tx basecoin.Tx, checkTx bool) (res wrsp.Result, diffIn, diffOut coin.Coins) {
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
		DefaultHandler("atom"),
		eyesCli,
		log.TestingLogger().With("module", "app"),
	)

	//testing ChainID
	chainID := "testChain"
	res := app.SetOption("base/chain_id", chainID)
	assert.EqualValues(app.GetState().GetChainID(), chainID)
	assert.EqualValues(res, "Success")

	// make a nice account...
	bal := coin.Coins{{"atom", 77}, {"eth", 12}}
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
	unsortCoins := coin.Coins{{"BTC", 789}, {"eth", 123}}
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
	at.acctIn.Coins = coin.Coins{{"mycoin", 2}}
	at.initAccount(at.acctIn)
	res, _, _ := at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}), true)
	assert.True(res.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res)
	res, diffIn, diffOut := at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}), false)
	assert.True(res.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res)
	assert.True(diffIn.IsZero())
	assert.True(diffOut.IsZero())

	//Regular CheckTx
	at.reset()
	res, _, _ = at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}), true)
	assert.True(res.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res)

	//Regular DeliverTx
	at.reset()
	amt := coin.Coins{{"mycoin", 3}}
	res, diffIn, diffOut = at.exec(t, at.getTx(amt), false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.Equal(amt.Negative(), diffIn)
	assert.Equal(amt, diffOut)

	//DeliverTx with fee.... 4 get to recipient, 1 extra taxed
	at.reset()
	amt = coin.Coins{{"mycoin", 4}}
	toll := coin.Coin{"mycoin", 1}
	res, diffIn, diffOut = at.exec(t, at.feeTx(amt, toll), false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	payment := amt.Plus(coin.Coins{toll}).Negative()
	assert.Equal(payment, diffIn)
	assert.Equal(amt, diffOut)

}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	res, _, _ := at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}), false)
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
