package coin

import (
	"testing"

	cmn "github.com/tepleton/tmlibs/common"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/state"
)

func makeHandler() basecoin.Handler {
	return NewHandler()
}

func makeSimpleTx(from, to basecoin.Actor, amount Coins, seq int) basecoin.Tx {
	in := []TxInput{{Address: from, Coins: amount, Sequence: seq}}
	out := []TxOutput{{Address: to, Coins: amount}}
	return NewSendTx(in, out)
}

func BenchmarkSimpleTransfer(b *testing.B) {
	h := makeHandler()
	store := state.NewMemKVStore()
	logger := log.NewNopLogger()

	// set the initial account
	acct := NewAccountWithKey(Coins{{"mycoin", 1234567890}})
	h.SetOption(logger, store, NameCoin, "account", acct.MakeOption())
	sender := acct.Actor()
	receiver := basecoin.Actor{App: "foo", Address: cmn.RandBytes(20)}

	// now, loop...
	for i := 1; i <= b.N; i++ {
		ctx := stack.MockContext("foo").WithPermissions(sender)
		tx := makeSimpleTx(sender, receiver, Coins{{"mycoin", 2}}, i)
		_, err := h.DeliverTx(ctx, store, tx)
		// never should error
		if err != nil {
			panic(err)
		}
	}

}
