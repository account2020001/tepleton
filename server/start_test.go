package server

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/tepleton/tepleton-sdk/mock"
	"github.com/tepleton/wrsp/server"
	"github.com/tepleton/tmlibs/log"
)

func TestStartStandAlone(t *testing.T) {
	home, err := ioutil.TempDir("", "mock-sdk-cmd")
	defer func() {
		os.RemoveAll(home)
	}()

	logger := log.NewNopLogger()
	initCmd := InitCmd(mock.GenInitOptions, logger)
	err = initCmd.RunE(nil, nil)
	require.NoError(t, err)

	app, err := mock.NewApp(home, logger)
	require.Nil(t, err)
	svr, err := server.NewServer(FreeTCPAddr(t), "socket", app)
	require.Nil(t, err, "Error creating listener")
	svr.SetLogger(logger.With("module", "wrsp-server"))
	svr.Start()

	timer := time.NewTimer(time.Duration(5) * time.Second)
	select {
	case <-timer.C:
		svr.Stop()
	}

}

func TestStartWithTendermint(t *testing.T) {
	defer setupViper(t)()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "mock-cmd")
	initCmd := InitCmd(mock.GenInitOptions, logger)
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// set up app and start up
	viper.Set(flagWithTendermint, true)
	startCmd := StartCmd(mock.NewApp, logger)
	startCmd.Flags().Set(flagAddress, FreeTCPAddr(t)) // set to a new free address
	timeout := time.Duration(5) * time.Second

	close(RunOrTimeout(startCmd, timeout, t))
}
