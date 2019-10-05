package app

import (
	"encoding/json"

	wrsp "github.com/tepleton/wrsp/types"
	crypto "github.com/tepleton/go-crypto"
	"github.com/tepleton/go-wire"
	cmn "github.com/tepleton/tmlibs/common"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	bam "github.com/tepleton/tepleton-sdk/baseapp"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"

	"github.com/tepleton/tepleton-sdk/examples/basecoin/types"
	"github.com/tepleton/tepleton-sdk/examples/basecoin/x/sketchy"
)

const (
	appName = "BasecoinApp"
)

// Extended WRSP application
type BasecoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore *sdk.KVStoreKey
	capKeyIBCStore  *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
}

func NewBasecoinApp(logger log.Logger, db dbm.DB) *BasecoinApp {
	// create your application object
	var app = &BasecoinApp{
		BaseApp:         bam.NewBaseApp(appName, logger, db),
		cdc:             MakeTxCodec(),
		capKeyMainStore: sdk.NewKVStoreKey("main"),
		capKeyIBCStore:  sdk.NewKVStoreKey("ibc"),
	}

	// define the accountMapper
	app.accountMapper = auth.NewAccountMapperSealed(
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	)

	// add handlers
	coinKeeper := bank.NewCoinKeeper(app.accountMapper)
	app.Router().AddRoute("bank", bank.NewHandler(coinKeeper))
	app.Router().AddRoute("sketchy", sketchy.NewHandler())

	// initialize BaseApp
	app.SetTxDecoder(app.txDecoder)
	app.SetInitChainer(app.initChainer)
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyIBCStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper))
	err := app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// custom tx codec
func MakeTxCodec() *wire.Codec {
	cdc := wire.NewCodec()
	crypto.RegisterWire(cdc) // Register crypto.[PubKey,PrivKey,Signature] types.
	bank.RegisterWire(cdc)   // Register bank.[SendMsg,IssueMsg] types.
	return cdc
}

// custom logic for transaction decoding
func (app *BasecoinApp) txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx = sdk.StdTx{}
	// StdTx.Msg is an interface whose concrete
	// types are registered in app/msgs.go.
	err := app.cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxParse("").TraceCause(err, "")
	}
	return tx, nil
}

// custom logic for basecoin initialization
func (app *BasecoinApp) initChainer(ctx sdk.Context, req wrsp.RequestInitChain) wrsp.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(types.GenesisState)
	err := json.Unmarshal(stateJSON, genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAppAccount()
		if err != nil {
			panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}
		app.accountMapper.SetAccount(ctx, acc)
	}
	return wrsp.ResponseInitChain{}
}
