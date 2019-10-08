package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/tepleton/tmlibs/cli"

	"github.com/tepleton/tepleton-sdk/version"
)

// toncliCmd is the entry point for this binary
var (
	basecliCmd = &cobra.Command{
		Use:   "basecli",
		Short: "Basecoin light-client",
	}

	lineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}

	getAccountCmd = &cobra.Command{
		Use:   "account <address>",
		Short: "Query account balance",
		RunE:  todoNotImplemented,
	}
)

func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("TODO: Command not yet implemented")
}

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// generic client commands
	AddClientCommands(basecliCmd)
	// query commands (custom to binary)
	basecliCmd.AddCommand(
		GetCommands(getAccountCmd)...)
	// post tx commands (custom to binary)
	basecliCmd.AddCommand(
		PostCommands(postSendCommand())...)

	// add proxy, version and key info
	basecliCmd.AddCommand(
		lineBreak,
		serveCommand(),
		KeyCommands(),
		lineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(basecliCmd, "GA", os.ExpandEnv("$HOME/.basecli"))
	executor.Execute()
}
