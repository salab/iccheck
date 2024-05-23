package cmd

import (
	"errors"
	"fmt"
	"github.com/salab/iccheck/pkg/printer"
	"github.com/salab/iccheck/pkg/search"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "iccheck",
	Short:         "Finds inconsistent changes in your git changes",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch logLevel {
		case "debug":
			slog.SetLogLoggerLevel(slog.LevelDebug)
		case "info":
			slog.SetLogLoggerLevel(slog.LevelInfo)
		case "warn":
			slog.SetLogLoggerLevel(slog.LevelWarn)
		case "error":
			slog.SetLogLoggerLevel(slog.LevelError)
		default:
			return errors.New("invalid log level")
		}

		// Search for inconsistent changes
		clones, err := search.Search(repoDir, fromRef, toRef)
		if err != nil {
			return err
		}

		// Report the findings
		if len(clones) == 0 {
			slog.Info(fmt.Sprintf("No clones are missing inconsistent changes."))
		} else {
			slog.Info(fmt.Sprintf("%d clone(s) are likely missing a consistent change.", len(clones)))
		}

		printer := getPrinter()
		out := printer.PrintClones(repoDir, clones)
		fmt.Print(string(out))

		// If any inconsistent changes are found, exit with specified code
		if len(clones) > 0 {
			os.Exit(failCode)
		} else {
			os.Exit(0)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	repoDir string
	fromRef string
	toRef   string

	logLevel   string
	formatType string
	failCode   int
)

func init() {
	rootCmd.Flags().StringVarP(&repoDir, "repo", "r", ".", "Source git directory")
	rootCmd.Flags().StringVarP(&fromRef, "from", "f", "main", "Target git ref to compare against. Usually earlier in time.")
	rootCmd.Flags().StringVarP(&toRef, "to", "t", "HEAD", "Source git ref to compare from. Usually later in time. Set to 'WORKTREE' to specify worktree.")

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&formatType, "format", "console", "Format type (console, json, github)")
	rootCmd.Flags().IntVar(&failCode, "fail-code", 0, "Exit code if it detects any inconsistent changes (default: 0)")
}

func getPrinter() printer.Printer {
	switch formatType {
	case "console":
		return printer.NewConsolePrinter()
	case "json":
		return printer.NewJsonPrinter()
	case "github":
		return printer.NewGitHubPrinter()
	default:
		panic(fmt.Sprintf("unknown format type: %s", formatType))
	}
}
