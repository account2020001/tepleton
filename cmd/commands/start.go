package commands

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli"

	"github.com/tepleton/wrsp/server"
	cmn "github.com/tepleton/go-common"
	eyes "github.com/tepleton/merkleeyes/client"

	tmcfg "github.com/tepleton/tepleton/config/tepleton"
	"github.com/tepleton/tepleton/node"
	"github.com/tepleton/tepleton/proxy"
	tmtypes "github.com/tepleton/tepleton/types"

	"github.com/tepleton/basecoin/app"
	"github.com/tepleton/basecoin/types"
)

const EyesCacheSize = 10000

var StartCmd = cli.Command{
	Name:      "start",
	Usage:     "Start basecoin",
	ArgsUsage: "",
	Action: func(c *cli.Context) error {
		return cmdStart(c)
	},
	Flags: []cli.Flag{
		AddrFlag,
		EyesFlag,
		WRSPServerFlag,
		ChainIDFlag,
	},
}

type plugin struct {
	name      string
	newPlugin func() types.Plugin
}

var plugins = []plugin{}

// RegisterStartPlugin is used to enable a plugin
func RegisterStartPlugin(name string, newPlugin func() types.Plugin) {
	plugins = append(plugins, plugin{name: name, newPlugin: newPlugin})
}

func cmdStart(c *cli.Context) error {
	basecoinDir := BasecoinRoot("")

	// Connect to MerkleEyes
	var eyesCli *eyes.Client
	if c.String("eyes") == "local" {
		eyesCli = eyes.NewLocalClient(path.Join(basecoinDir, "merkleeyes.db"), EyesCacheSize)
	} else {
		var err error
		eyesCli, err = eyes.NewClient(c.String("eyes"))
		if err != nil {
			return errors.New("connect to MerkleEyes: " + err.Error())
		}
	}

	// Create Basecoin app
	basecoinApp := app.NewBasecoin(eyesCli)

	// register IBC plugn
	basecoinApp.RegisterPlugin(NewIBCPlugin())

	// register all other plugins
	for _, p := range plugins {
		basecoinApp.RegisterPlugin(p.newPlugin())
	}

	// If genesis file exists, set key-value options
	genesisFile := path.Join(basecoinDir, "genesis.json")
	if _, err := os.Stat(genesisFile); err == nil {
		err := basecoinApp.LoadGenesis(genesisFile)
		if err != nil {
			return errors.New(cmn.Fmt("%+v", err))
		}
	} else {
		fmt.Printf("No genesis file at %s, skipping...\n", genesisFile)
	}

	if c.Bool("wrsp-server") {
		// run just the wrsp app/server
		if err := startBasecoinWRSP(c, basecoinApp); err != nil {
			return err
		}
	} else {
		// start the app with tepleton in-process
		startTendermint(basecoinDir, basecoinApp)
	}

	return nil
}

func startBasecoinWRSP(c *cli.Context, basecoinApp *app.Basecoin) error {
	// Start the WRSP listener
	svr, err := server.NewServer(c.String("address"), "socket", basecoinApp)
	if err != nil {
		return errors.New("create listener: " + err.Error())
	}
	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
	return nil

}

func startTendermint(dir string, basecoinApp *app.Basecoin) {
	// Get configuration
	tmConfig := tmcfg.GetConfig(path.Join(dir, "tepleton"))

	// logger.SetLogLevel("notice") //config.GetString("log_level"))

	// parseFlags(config, args[1:]) // Command line overrides

	// Create & start tepleton node
	privValidatorFile := tmConfig.GetString("priv_validator_file")
	privValidator := tmtypes.LoadOrGenPrivValidator(privValidatorFile)
	n := node.NewNode(tmConfig, privValidator, proxy.NewLocalClientCreator(basecoinApp))

	n.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		n.Stop()
	})
}
