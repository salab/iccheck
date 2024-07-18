package cmd

import (
	"context"
	"fmt"
	"github.com/salab/iccheck/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"
	"net"
	"os"
)

var lspCmd = &cobra.Command{
	Use:          "lsp",
	Short:        "Starts ICCheck Language Server",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Code is partially copied from https://github.com/vito/bass/blob/main/cmd/bass/lsp.go
		ctx := context.Background()
		//rwc := lo.Must(newSocketRWC(8080))
		conn := jsonrpc2.NewConn(
			ctx,
			jsonrpc2.NewBufferedStream(stdRWC{}, jsonrpc2.VSCodeObjectCodec{}),
			lsp.NewHandler(),
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

type socketRWC struct {
	conn net.Conn
}

func newSocketRWC(port int) (*socketRWC, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	conn, err := l.Accept()
	if err != nil {
		return nil, err
	}
	return &socketRWC{conn: conn}, nil
}

func (s *socketRWC) Read(p []byte) (n int, err error) {
	return s.conn.Read(p)
}

func (s *socketRWC) Write(p []byte) (n int, err error) {
	return s.conn.Write(p)
}

func (s *socketRWC) Close() error {
	return s.conn.Close()
}
