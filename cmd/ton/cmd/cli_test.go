package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tepleton/tepleton-sdk/server"
	"github.com/tepleton/tepleton-sdk/tests"
	"github.com/tepleton/tepleton-sdk/x/auth"
)

func TestGaiaCLI(t *testing.T) {

	tests.ExecuteT(t, "tond unsafe_reset_all")
	pass := "1234567890"
	executeWrite(t, "toncli keys delete foo", pass)
	executeWrite(t, "toncli keys delete bar", pass)
	masterKey, chainID := executeInit(t, "tond init")

	// get a free port, also setup some common flags
	servAddr := server.FreeTCPAddr(t)
	flags := fmt.Sprintf("--node=%v --chain-id=%v", servAddr, chainID)

	// start tond server
	cmd, _, _ := tests.GoExecuteT(t, fmt.Sprintf("tond start --rpc.laddr=%v", servAddr))
	defer cmd.Process.Kill()

	executeWrite(t, "toncli keys add foo --recover", pass, masterKey)
	executeWrite(t, "toncli keys add bar", pass)

	fooAddr := executeGetAddr(t, "toncli keys show foo")
	barAddr := executeGetAddr(t, "toncli keys show bar")
	executeWrite(t, fmt.Sprintf("toncli send %v --amount=10fermion --to=%v --name=foo", flags, barAddr), pass)
	time.Sleep(time.Second * 3) // waiting for some blocks to pass

	barAcc := executeGetAccount(t, fmt.Sprintf("toncli account %v %v", barAddr, flags))
	assert.Equal(t, int64(10), barAcc.GetCoins().AmountOf("fermion"))
	fooAcc := executeGetAccount(t, fmt.Sprintf("toncli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(99990), fooAcc.GetCoins().AmountOf("fermion"))

	// declare candidacy
	//executeWrite(t, "toncli declare-candidacy -", pass)
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) {
	cmd, wc, _ := tests.GoExecuteT(t, cmdStr)
	for _, write := range writes {
		_, err := wc.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}
	cmd.Wait()
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

func executeGetAddr(t *testing.T, cmdStr string) (addr string) {
	out := tests.ExecuteT(t, cmdStr)
	name := strings.SplitN(cmdStr, " show ", 2)[1]
	return strings.TrimLeft(out, name+"\t")
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
