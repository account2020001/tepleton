package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tepleton/tmlibs/cli"

	"github.com/tepleton/tepleton-sdk/client"
	"github.com/tepleton/tepleton-sdk/client/keys"
	"github.com/tepleton/tepleton-sdk/client/lcd"
	"github.com/tepleton/tepleton-sdk/client/rpc"
	"github.com/tepleton/tepleton-sdk/client/tx"

	"github.com/tepleton/tepleton-sdk/version"
	authcmd "github.com/tepleton/tepleton-sdk/x/auth/client/cli"
	bankcmd "github.com/tepleton/tepleton-sdk/x/bank/client/cli"
	ibccmd "github.com/tepleton/tepleton-sdk/x/ibc/client/cli"
	simplestakingcmd "github.com/tepleton/tepleton-sdk/x/simplestake/client/cli"

	"github.com/tepleton/tepleton-sdk/examples/basecoin/app"
	"github.com/tepleton/tepleton-sdk/examples/basecoin/types"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "basecli",
		Short: "Basecoin light-client",
	}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.MakeCodec()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	// add standard rpc, and tx commands
	rpc.AddCommands(rootCmd)
	rootCmd.AddCommand(client.LineBreak)
	tx.AddCommands(rootCmd, cdc)
	rootCmd.AddCommand(client.LineBreak)

	// add query/post commands (custom to binary)
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("main", cdc, types.GetAccountDecoder(cdc)),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			ibccmd.IBCTransferCmd(cdc),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			ibccmd.IBCRelayCmd(cdc),
			simplestakingcmd.BondTxCmd(cdc),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			simplestakingcmd.UnbondTxCmd(cdc),
		)...)

	// add proxy, version and key info
	rootCmd.AddCommand(
		client.LineBreak,
		lcd.ServeCommand(cdc),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "BC", os.ExpandEnv("$HOME/.basecli"))
	executor.Execute()
}
