package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/salab/iccheck/pkg/lsp"
	"github.com/salab/iccheck/pkg/utils/cli"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Starts ICCheck Language Server",
	Long: fmt.Sprintf(`ICCheck %v
Starts ICCheck Language Server.`, cli.GetFormattedVersion()),
	SilenceUsage: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		slog.Info("ICCheck " + cli.GetFormattedVersion())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Code is partially copied from https://github.com/vito/bass/blob/main/cmd/bass/lsp.go
		ctx := context.Background()
		handler := lsp.NewHandler(
			algorithm,
			time.Duration(lspTimeoutSeconds)*time.Second,
			ignoreCLIOptions,
			disableDefaultIgnore,
			detectMicro,
		)
		// rwc := lo.Must(newSocketRWC(8080))
		conn := jsonrpc2.NewConn(
			ctx,
			jsonrpc2.NewBufferedStream(stdRWC{}, jsonrpc2.VSCodeObjectCodec{}),
			handler,
		)
		<-conn.DisconnectNotify()
		return nil
	},
}

var (
	lspTimeoutSeconds int
)

func init() {
	lspCmd.Flags().IntVar(&lspTimeoutSeconds, "timeout-seconds", 15, "Timeout for detecting clones in seconds (default: 15)")
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
