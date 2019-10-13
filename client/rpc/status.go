package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	wire "github.com/tepleton/go-wire"

	"github.com/tepleton/tepleton-sdk/client"
	ctypes "github.com/tepleton/tepleton/rpc/core/types"
)

func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  printNodeStatus,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	return cmd
}

func getNodeStatus() (*ctypes.ResultStatus, error) {
	// get the node
	node, err := client.GetNode()
	if err != nil {
		return &ctypes.ResultStatus{}, err
	}
	return node.Status()
}

// CMD

func printNodeStatus(cmd *cobra.Command, args []string) error {
	status, err := getNodeStatus()
	if err != nil {
		return err
	}

	output, err := wire.MarshalJSON(res)
	// output, err := json.MarshalIndent(res, "  ", "")
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

// REST

// TODO match desired spec output
func NodeStatusRequestHandler(w http.ResponseWriter, r *http.Request) {
	status, err := getNodeStatus()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	nodeInfo := status.NodeInfo
	output, err := json.MarshalIndent(nodeInfo, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(output)
}
