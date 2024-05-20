package cmd

import (
	"errors"
	"fmt"
	"github.com/salab/iccheck/pkg/search"
	"github.com/salab/iccheck/pkg/utils/printer"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "iccheck",
	Short:        "Finds inconsistent changes in your git changes",
	SilenceUsage: true,
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

		clones, err := search.Search(repoDir, fromRef, toRef)
		if err != nil {
			return err
		}

		// Report the findings
		printer := getPrinter()
		if len(clones) == 0 {
			slog.Info(fmt.Sprintf("No clones are missing inconsistent changes."))
		} else {
			slog.Info(fmt.Sprintf("%d clone(s) are likely missing a consistent change.", len(clones)))
		}
		printer.PrintClones(repoDir, clones)

		// If any inconsistent changes are found, exit with non-zero code
		if len(clones) == 0 {
			return nil
		} else {
			return errors.New("inconsistent changes found")
		}
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

	logLevel string

	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	isGitHubActions = os.Getenv("GITHUB_ACTIONS") == "true"
)

func init() {
	rootCmd.Flags().StringVarP(&repoDir, "repo", "r", ".", "Source git directory")
	rootCmd.Flags().StringVarP(&fromRef, "from", "f", "main", "Target git ref to compare against. Usually earlier in time.")
	rootCmd.Flags().StringVarP(&toRef, "to", "t", "HEAD", "Source git ref to compare from. Usually later in time.")

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}

func getPrinter() printer.Printer {
	if isGitHubActions {
		return printer.NewGitHubPrinter()
	}
	return printer.NewStdoutPrinter()
}
