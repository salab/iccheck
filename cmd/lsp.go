package cmd

import (
	"context"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"
	"os"
)

var lspCmd = &cobra.Command{
	Use:           "lsp",
	Short:         "Starts ICCheck Language Server",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Code is partially copied from https://github.com/vito/bass/blob/main/cmd/bass/lsp.go
		ctx := context.Background()
		conn := jsonrpc2.NewConn(
			ctx,
			jsonrpc2.NewBufferedStream(stdRWC{}, jsonrpc2.VSCodeObjectCodec{}),
			nil,
		)
		<-conn.DisconnectNotify()
		return nil
	},
}

type stdRWC struct{}

func (stdRWC) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (c stdRWC) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (c stdRWC) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
