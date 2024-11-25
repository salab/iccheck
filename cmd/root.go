package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/printer"
	"github.com/salab/iccheck/pkg/search"
	"github.com/salab/iccheck/pkg/utils/cli"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "iccheck",
	Version:      cli.GetFormattedVersion(),
	Short:        "Finds inconsistent changes in your git changes",
	Long:         `Finds inconsistent changes in your git changes.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
		defer cancel()

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

		// Read ignore rules
		ignore, err := domain.ReadIgnoreRules(repoDir, ignoreCLIOptions, disableDefaultIgnore)
		if err != nil {
			return errors.Wrapf(err, "reading ignore rules")
		}

		// Prepare
		repo, err := git.PlainOpen(repoDir)
		if err != nil {
			return errors.Wrapf(err, "opening repository at %v", repoDir)
		}

		if (fromRef == "" && toRef != "") || (fromRef != "" && toRef == "") {
			return errors.New("only one of --from or --to is set, this is invalid - do not set for automatic ref detection or set both")
		}
		if fromRef == "" && toRef == "" {
			fromRef, toRef, err = autoDetermineRefs(repo)
			if err != nil {
				return errors.Wrapf(err, "automatically determining refs")
			}
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
		cloneSets, err := search.Search(ctx, fromTree, toTree, ignore)
		if err != nil {
			return err
		}

		// If all clones in a set went through some changes, no need to notify
		cloneSets = lo.Filter(cloneSets, func(cs *domain.CloneSet, _ int) bool { return len(cs.Missing) > 0 })

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

	logLevel       string
	formatType     string
	failCode       int
	timeoutSeconds int
)

var (
	ignoreCLIOptions     []string
	disableDefaultIgnore bool
)

func init() {
	rootCmd.Flags().StringVarP(&repoDir, "repo", "r", ".", "Source git directory")
	rootCmd.Flags().StringVarP(&fromRef, "from", "f", "", "Base git ref to compare against. Usually earlier in time.")
	rootCmd.Flags().StringVarP(&toRef, "to", "t", "", `Target git ref to compare from. Usually later in time.
Can accept special value "WORKTREE" to specify the current worktree.`)

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&formatType, "format", "console", "Format type (console, json, github)")
	rootCmd.Flags().IntVar(&failCode, "fail-code", 0, "Exit code if it detects any inconsistent changes (default: 0)")
	rootCmd.Flags().IntVar(&timeoutSeconds, "timeout-seconds", 15, "Timeout for detecting clones in seconds (default: 15)")

	rootCmd.PersistentFlags().StringArrayVar(&ignoreCLIOptions, "ignore", nil, `Regexp of file paths (and its contents) to ignore.
If specifying both file paths and contents ignore regexp, split them by ':'.
Example (ignore dist directory): --ignore '^dist/'
Example (ignore import statements in js files): --ignore '\.m?[jt]s$:^import'`)
	rootCmd.PersistentFlags().BoolVar(&disableDefaultIgnore, "disable-default-ignore", false, "Disable default ignore configs")

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(lspCmd)
}

func autoDetermineRefs(repo *git.Repository) (from, to string, err error) {
	ref, err := repo.Reference(plumbing.HEAD, true)
	if err != nil {
		return "", "", errors.Wrapf(err, "resolving HEAD ref")
	}

	// Just assume that the "default branch" for this repository is master or main
	// - default branch per repository (except for init.defaultBranch config) is a remote-specific config
	// and cannot be retrieved from local repository.
	// cf. https://stackoverflow.com/a/70080259
	headName := ref.Name().Short()
	if headName == "main" || headName == "master" {
		// If we're on default branch, a reasonable default would be from this branch to the current worktree.
		return headName, worktreeRef, nil
	}

	// We're not on default branch, so let's calculate diff from default branch to current HEAD.
	// Let's check if 'main' or 'master' is present in this repository
	ref, err = repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	if err == nil {
		return ref.Name().Short(), "HEAD", nil
	}
	ref, err = repo.Reference(plumbing.NewBranchReferenceName("master"), true)
	if err == nil {
		return ref.Name().Short(), "HEAD", nil
	}

	// We were not able to determine default branch from this local repository
	// - fallback to listing remote refs and finding HEAD from the result.
	remote, err := repo.Remote("origin")
	if err == nil {
		slog.Info("Fetching origin to determine default branch...")
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			return "", "", errors.Wrapf(err, "listing refs from origin")
		}
		for _, ref := range refs {
			if ref.Type() == plumbing.SymbolicReference && ref.Name().String() == "HEAD" {
				// We found refs/heads/origin/HEAD - its symbolic ref target is the default branch.
				defaultBranch := ref.Target().Short()
				return defaultBranch, "HEAD", nil
			}
		}
	}

	return "", "", errors.New("unable to determine default compare refs - please manually set 'from' and 'to' parameters")
}

const worktreeRef = "WORKTREE"

func resolveTree(repo *git.Repository, ref string) (domain.Tree, error) {
	// Special refs
	if ref == worktreeRef {
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, errors.Wrap(err, "retrieving worktree")
		}
		return domain.NewGoGitWorkTree(worktree)
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
