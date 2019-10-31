package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tepleton/tepleton-sdk/client/keys"
	"github.com/tepleton/tepleton-sdk/server"
	"github.com/tepleton/tepleton-sdk/tests"
	"github.com/tepleton/tepleton-sdk/x/auth"
	crypto "github.com/tepleton/go-crypto"
	crkeys "github.com/tepleton/go-crypto/keys"
)

//func TestGaiaCLISend(t *testing.T) {

//tests.ExecuteT(t, "tond unsafe_reset_all")
//pass := "1234567890"
//executeWrite(t, "toncli keys delete foo", pass)
//executeWrite(t, "toncli keys delete bar", pass)
//masterKey, chainID := executeInit(t, "tond init")

//// get a free port, also setup some common flags
//servAddr := server.FreeTCPAddr(t)
//flags := fmt.Sprintf("--node=%v --chain-id=%v", servAddr, chainID)

//// start tond server
//cmd, _, _ := tests.GoExecuteT(t, fmt.Sprintf("tond start --rpc.laddr=%v", servAddr))
//defer cmd.Process.Kill()

//executeWrite(t, "toncli keys add foo --recover", pass, masterKey)
//executeWrite(t, "toncli keys add bar", pass)

//fooAddr, _ := executeGetAddr(t, "toncli keys show foo --output=json")
//barAddr, _ := executeGetAddr(t, "toncli keys show bar --output=json")

//fooAcc := executeGetAccount(t, fmt.Sprintf("toncli account %v %v", fooAddr, flags))
//assert.Equal(t, int64(100000), fooAcc.GetCoins().AmountOf("fermion"))

//executeWrite(t, fmt.Sprintf("toncli send %v --amount=10fermion --to=%v --name=foo", flags, barAddr), pass)
//time.Sleep(time.Second * 3) // waiting for some blocks to pass

//barAcc := executeGetAccount(t, fmt.Sprintf("toncli account %v %v", barAddr, flags))
//assert.Equal(t, int64(10), barAcc.GetCoins().AmountOf("fermion"))
//fooAcc = executeGetAccount(t, fmt.Sprintf("toncli account %v %v", fooAddr, flags))
//assert.Equal(t, int64(99990), fooAcc.GetCoins().AmountOf("fermion"))
//}

func TestGaiaCLIDeclareCandidacy(t *testing.T) {

	tests.ExecuteT(t, "tond unsafe_reset_all")
	pass := "1234567890"
	executeWrite(t, "toncli keys delete foo", pass)
	masterKey, chainID := executeInit(t, "tond init")

	// get a free port, also setup some common flags
	servAddr := server.FreeTCPAddr(t)
	flags := fmt.Sprintf("--node=%v --chain-id=%v", servAddr, chainID)

	// start tond server
	cmd, _, _ := tests.GoExecuteT(t, fmt.Sprintf("tond start --rpc.laddr=%v", servAddr))
	defer cmd.Process.Kill()

	executeWrite(t, "toncli keys add foo --recover", pass, masterKey)
	fooAddr, fooPubKey := executeGetAddr(t, "toncli keys show foo --output=json")
	fooAcc := executeGetAccount(t, fmt.Sprintf("toncli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(100000), fooAcc.GetCoins().AmountOf("fermion"))

	// declare candidacy
	//--address-candidate string   hex address of the validator/candidate
	//--amount string              Amount of coins to bond (default "1fermion")
	//--chain-id string            Chain ID of tepleton node
	//--fee string                 Fee to pay along with transaction
	//--keybase-sig string         optional keybase signature
	//--moniker string             validator-candidate name
	//--name string                Name of private key with which to sign
	//--node string                <host>:<port> to tepleton rpc interface for this chain (default "tcp://localhost:46657")
	//--pubkey string              PubKey of the validator-candidate
	//--sequence int               Sequence number to sign the tx
	//--website string             optional website
	//_ = fooPubKey
	declStr := fmt.Sprintf("toncli declare-candidacy %v", flags)
	declStr += fmt.Sprintf(" --name=%v", "foo")
	declStr += fmt.Sprintf(" --address-candidate=%v", fooAddr)
	declStr += fmt.Sprintf(" --pubkey=%v", fooPubKey)
	declStr += fmt.Sprintf(" --amount=%v", "3fermion")
	declStr += fmt.Sprintf(" --moniker=%v", "foo-vally")
	fmt.Printf("debug declStr: %v\n", declStr)
	executeWrite(t, declStr, pass)
	fooAcc = executeGetAccount(t, fmt.Sprintf("toncli account %v %v", fooAddr, flags))
	time.Sleep(time.Second * 3) // waiting for some blocks to pass
	assert.Equal(t, int64(99997), fooAcc.GetCoins().AmountOf("fermion"))
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) {
	cmd, wc, _ := tests.GoExecuteT(t, cmdStr)

	for _, write := range writes {
		_, err := wc.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}
	cmd.Wait()
}

func executeWritePrint(t *testing.T, cmdStr string, writes ...string) {
	cmd, wc, rc := tests.GoExecuteT(t, cmdStr)

	for _, write := range writes {
		_, err := wc.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}
	cmd.Wait()

	bz := make([]byte, 100000)
	rc.Read(bz)
	fmt.Printf("debug read: %v\n", string(bz))
}

func executeInit(t *testing.T, cmdStr string) (masterKey, chainID string) {
	out := tests.ExecuteT(t, cmdStr)
	outCut := "{" + strings.SplitN(out, "{", 2)[1] // weird I'm sorry

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(outCut), &initRes)
	require.NoError(t, err)
	err = json.Unmarshal(initRes["secret"], &masterKey)
	require.NoError(t, err)
	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(t, err)
	return
}

func executeGetAddr(t *testing.T, cmdStr string) (addr, pubKey string) {
	out := tests.ExecuteT(t, cmdStr)
	var info crkeys.Info
	keys.UnmarshalJSON([]byte(out), &info)
	pubKey = hex.EncodeToString(info.PubKey.(crypto.PubKeyEd25519).Bytes())
	pubKey = strings.TrimLeft(pubKey, "1624de6220")
	fmt.Printf("debug pubKey: %v\n", pubKey)
	addr = info.PubKey.Address().String()
	fmt.Printf("debug addr: %v\n", addr)
	return
}

func executeGetAccount(t *testing.T, cmdStr string) auth.BaseAccount {
	out := tests.ExecuteT(t, cmdStr)
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	_ = json.Unmarshal(value, &acc) //XXX pubkey can't be decoded go amino issue
	require.NoError(t, err, "value %v, err %v", string(value), err)
	return acc
}