package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"

	wrsp "github.com/tepleton/wrsp/types"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/tepleton-sdk/store"
	sdk "github.com/tepleton/tepleton-sdk/types"
	wire "github.com/tepleton/tepleton-sdk/wire"

	"github.com/tepleton/tepleton-sdk/x/auth"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	authKey := sdk.NewKVStoreKey("authkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authKey
}

func TestKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, wrsp.Header{}, false, nil, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
	coinKeeper := NewKeeper(accountMapper)

	addr := sdk.Address([]byte("addr1"))
	addr2 := sdk.Address([]byte("addr2"))
	addr3 := sdk.Address([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 10)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 10)}))

	// Test HasCoins
	assert.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 10)}))
	assert.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 5)}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 15)}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 5)}))

	// Test AddCoins
	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 15)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 25)}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 15)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 15), sdk.NewCoin("foocoin", 25)}))

	// Test SubtractCoins
	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 10)})
	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 5)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 15)}))

	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 11)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 15)}))

	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 10)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 15)}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 1)}))

	// Test SendCoins
	coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewCoin("foocoin", 5)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 10)}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 5)}))

	_, err2 := coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewCoin("foocoin", 50)})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 10)}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 5)}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 30)})
	coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 5)})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 20), sdk.NewCoin("foocoin", 5)}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 10)}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, sdk.Coins{sdk.NewCoin("foocoin", 2)})
	output1 := NewOutput(addr, sdk.Coins{sdk.NewCoin("foocoin", 2)})
	coinKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 20), sdk.NewCoin("foocoin", 7)}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 8)}))

	inputs := []Input{
		NewInput(addr, sdk.Coins{sdk.NewCoin("foocoin", 3)}),
		NewInput(addr2, sdk.Coins{sdk.NewCoin("barcoin", 3), sdk.NewCoin("foocoin", 2)}),
	}

	outputs := []Output{
		NewOutput(addr, sdk.Coins{sdk.NewCoin("barcoin", 1)}),
		NewOutput(addr3, sdk.Coins{sdk.NewCoin("barcoin", 2), sdk.NewCoin("foocoin", 5)}),
	}
	coinKeeper.InputOutputCoins(ctx, inputs, outputs)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 21), sdk.NewCoin("foocoin", 4)}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 7), sdk.NewCoin("foocoin", 6)}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 2), sdk.NewCoin("foocoin", 5)}))

}

func TestSendKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, wrsp.Header{}, false, nil, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
	coinKeeper := NewKeeper(accountMapper)
	sendKeeper := NewSendKeeper(accountMapper)

	addr := sdk.Address([]byte("addr1"))
	addr2 := sdk.Address([]byte("addr2"))
	addr3 := sdk.Address([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 10)})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 10)}))

	// Test HasCoins
	assert.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 10)}))
	assert.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 5)}))
	assert.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 15)}))
	assert.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 5)}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 15)})

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewCoin("foocoin", 5)})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 10)}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 5)}))

	_, err2 := sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewCoin("foocoin", 50)})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 10)}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 5)}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 30)})
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 5)})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 20), sdk.NewCoin("foocoin", 5)}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 10)}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, sdk.Coins{sdk.NewCoin("foocoin", 2)})
	output1 := NewOutput(addr, sdk.Coins{sdk.NewCoin("foocoin", 2)})
	sendKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 20), sdk.NewCoin("foocoin", 7)}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 10), sdk.NewCoin("foocoin", 8)}))

	inputs := []Input{
		NewInput(addr, sdk.Coins{sdk.NewCoin("foocoin", 3)}),
		NewInput(addr2, sdk.Coins{sdk.NewCoin("barcoin", 3), sdk.NewCoin("foocoin", 2)}),
	}

	outputs := []Output{
		NewOutput(addr, sdk.Coins{sdk.NewCoin("barcoin", 1)}),
		NewOutput(addr3, sdk.Coins{sdk.NewCoin("barcoin", 2), sdk.NewCoin("foocoin", 5)}),
	}
	sendKeeper.InputOutputCoins(ctx, inputs, outputs)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 21), sdk.NewCoin("foocoin", 4)}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 7), sdk.NewCoin("foocoin", 6)}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{sdk.NewCoin("barcoin", 2), sdk.NewCoin("foocoin", 5)}))

}

func TestViewKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, wrsp.Header{}, false, nil, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
	coinKeeper := NewKeeper(accountMapper)
	viewKeeper := NewViewKeeper(accountMapper)

	addr := sdk.Address([]byte("addr1"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 10)})
	assert.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewCoin("foocoin", 10)}))

	// Test HasCoins
	assert.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 10)}))
	assert.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 5)}))
	assert.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("foocoin", 15)}))
	assert.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewCoin("barcoin", 5)}))
}
