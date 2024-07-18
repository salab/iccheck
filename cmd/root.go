package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/printer"
	"github.com/salab/iccheck/pkg/search"
	"github.com/salab/iccheck/pkg/utils/cli"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "iccheck",
	Version: cli.GetFormattedVersion(),
	Short:   "Finds inconsistent changes in your git changes",
	Long: `Finds inconsistent changes in your git changes.

Specify special values in base or target git ref arguments to compare against some special filesystems.
  "WORKTREE" : Compare against the current worktree.`,
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

		// Prepare
		repo, err := git.PlainOpen(repoDir)
		if err != nil {
			return errors.Wrapf(err, "opening repository at %v", repoDir)
		}
		fromTree, err := resolveTree(repo, fromRef)
		if err != nil {
			return errors.Wrap(err, "resolving base tree")
		}
		toTree, err := resolveTree(repo, toRef)
		if err != nil {
			return errors.Wrap(err, "resolving target tree")
		}

		// Search for inconsistent changes
		slog.Info("Searching for inconsistent changes...", "repository", repoDir, "from", fromRef, "to", toRef)
		slog.Info(fmt.Sprintf("Base ref: %v", fromTree.String()))
		slog.Info(fmt.Sprintf("Target ref: %v", toTree.String()))
		cloneSets, err := search.Search(fromTree, toTree)
		if err != nil {
			return err
		}

		// Report the findings
		missingChanges := lo.SumBy(cloneSets, func(set *domain.CloneSet) int { return len(set.Missing) })
		if missingChanges == 0 {
			slog.Info(fmt.Sprintf("No clones are missing inconsistent changes."))
		} else {
			slog.Info(fmt.Sprintf("%d clone(s) are likely missing a consistent change.", missingChanges))
		}

		printer := getPrinter()
		out := printer.PrintClones(repoDir, cloneSets)
		fmt.Print(string(out))

		// If any inconsistent changes are found, exit with specified code
		if len(cloneSets) > 0 {
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
	rootCmd.Flags().StringVarP(&fromRef, "from", "f", "main", "Base git ref to compare against. Usually earlier in time.")
	rootCmd.Flags().StringVarP(&toRef, "to", "t", "HEAD", "Target git ref to compare from. Usually later in time.")

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&formatType, "format", "console", "Format type (console, json, github)")
	rootCmd.Flags().IntVar(&failCode, "fail-code", 0, "Exit code if it detects any inconsistent changes (default: 0)")

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(lspCmd)
}

const worktreeRef = "WORKTREE"

func resolveTree(repo *git.Repository, ref string) (domain.Tree, error) {
	// Special refs
	if ref == worktreeRef {
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, errors.Wrap(err, "retrieving worktree")
		}
		return domain.NewGoGitWorkTree(worktree), nil
	}

	// Normal git ref
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, errors.Wrapf(err, "resolving hash revision from %v", ref)
	}
	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving commit from hash %v", *hash)
	}
	return domain.NewGoGitCommitTree(commit), nil
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
