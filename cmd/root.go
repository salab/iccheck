package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/pflag"
	"golang.org/x/term"
	"log/slog"
	"os"
	"path/filepath"
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

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "iccheck",
	Short: "Finds inconsistent changes in your git changes",
	Long: fmt.Sprintf(`ICCheck %v
Finds inconsistent changes in your git changes.`, cli.GetFormattedVersion()),
	Version:      cli.GetFormattedVersion(),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
		defer cancel()

		// Prepare
		if err := setLogLevel(); err != nil {
			return err
		}
		ignore, err := readIgnoreRules()
		if err != nil {
			return err
		}
		repoDir, err := getRepoDir()
		if err != nil {
			return err
		}
		repo, err := git.PlainOpen(repoDir)
		if err != nil {
			return errors.Wrapf(err, "opening repository at %v", repoDir)
		}

		if fromRef == "" && toRef != "" {
			// Only --to ref was given - a reasonable default would be to compare from parent of that ref.
			fromRef = toRef + "^"
		} else if fromRef != "" && toRef == "" {
			return errors.New("only one of --from was set, this is invalid - do not set for automatic discovery or set both")
		} else if fromRef == "" && toRef == "" {
			fromRef, toRef, err = autoDetermineRefs(repo)
			if err != nil {
				return errors.Wrapf(err, "determining refs")
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
		queries, changedFiles, err := search.DiffTrees(ctx, fromTree, toTree)
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("%d changed text chunk(s) were found within %d changed file(s).", len(queries), changedFiles), "from", fromTree, "to", toTree)
		for i, q := range queries {
			slog.Debug(fmt.Sprintf("Query#%d", i), "query", q)
		}
		cloneSets, err := search.Search(ctx, algorithm, queries, toTree, ignore)
		if err != nil {
			return err
		}
		reportClones(cloneSets)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	fromRef string
	toRef   string
)

var (
	repoDir        string
	formatType     string
	failCode       int
	timeoutSeconds int
)

var (
	logLevel             string
	algorithm            string
	ignoreCLIOptions     []string
	disableDefaultIgnore bool
)

func init() {
	// Root command specific
	RootCmd.Flags().StringVarP(&fromRef, "from", "f", "", "Base git ref to compare against. Usually earlier in time.")
	RootCmd.Flags().StringVarP(&toRef, "to", "t", "", `Target git ref to compare from. Usually later in time.
Can accept special value "WORKTREE" to specify the current worktree.`)

	// Common to root command and "search" command
	searchFlags := pflag.NewFlagSet("search", pflag.ContinueOnError)
	searchFlags.StringVarP(&repoDir, "repo", "r", "", "Source git directory (supports bare)")
	searchFlags.StringVar(&formatType, "format", "console", "Format type (console, json, github)")
	searchFlags.IntVar(&failCode, "fail-code", 0, "Exit code if it detects any inconsistent changes")
	searchFlags.IntVar(&timeoutSeconds, "timeout-seconds", 60, "Timeout for detecting clones in seconds")

	RootCmd.Flags().AddFlagSet(searchFlags)
	searchCmd.Flags().AddFlagSet(searchFlags)

	// Common to all commands
	pfs := RootCmd.PersistentFlags()
	pfs.StringVar(&logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	pfs.StringVar(&algorithm, "algorithm", "fleccs", "Clone search algorithm to use (fleccs, ncdsearch)")
	pfs.StringArrayVar(&ignoreCLIOptions, "ignore", nil, `Regexp of file paths (and its contents) to ignore.
If specifying both file paths and contents ignore regexp, split them by ':'.
Example (ignore dist directory): --ignore '^dist/'
Example (ignore import statements in js files): --ignore '\.m?[jt]s$:^import'`)
	pfs.BoolVar(&disableDefaultIgnore, "disable-default-ignore", false, "Disable default ignore configs")

	// Disable "completion" command
	RootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add child commands
	RootCmd.AddCommand(lspCmd)
	RootCmd.AddCommand(searchCmd)
}

func autoDetermineLogLevel() string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "info"
	} else {
		// Suppress verbose logging messages by default, if output is not a tty - likely a pipe or a file.
		return "warn"
	}
}

func setLogLevel() error {
	if logLevel == "" {
		logLevel = autoDetermineLogLevel()
	}
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
	return nil
}

func readIgnoreRules() (domain.IgnoreRules, error) {
	ignore, err := domain.ReadIgnoreRules(repoDir, ignoreCLIOptions, disableDefaultIgnore)
	if err != nil {
		return nil, errors.Wrapf(err, "reading ignore rules")
	}
	return ignore, nil
}

func autoDetermineTopLevelGit() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "getting current working directory")
	}
	for {
		// Is this directory a git (bare or non-bare) repository?
		if _, err := git.PlainOpen(dir); err == nil {
			return dir, nil
		}

		// Recurse up to parent directory...
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no git repository found")
		}
		dir = parent
	}
}

func getRepoDir() (string, error) {
	if repoDir != "" {
		return repoDir, nil
	}
	repoDir, err := autoDetermineTopLevelGit()
	if err != nil {
		return "", errors.Wrap(err, "determining git repository")
	}
	return repoDir, nil
}

func autoDetermineRefs(repo *git.Repository) (from, to string, err error) {
	_, wtErr := repo.Worktree()
	isBare := errors.Is(wtErr, git.ErrIsBareRepository)

	if !isBare {
		// Let's see if there are any local changes on worktree
		wt, err := domain.NewGoGitWorkTree(repo)
		if err != nil {
			return "", "", errors.Wrapf(err, "getting worktree")
		}
		st, err := wt.Status()
		if err != nil {
			return "", "", errors.Wrapf(err, "querying worktree status")
		}
		if !st.IsClean() {
			// If there are local changes, then a reasonable default would be from HEAD to the current worktree.
			return "HEAD", worktreeRef, nil
		}
	}

	// Either the repository is bare, or the working tree is clean.
	// Let's check the default branch of this repository.
	// A reasonable default would be from default branch to HEAD.
	defaultBranch, err := determineDefaultBranch(repo)
	if err != nil {
		return "", "", errors.Wrapf(err, "determining default branch")
	}

	// Unless, we're already on that default branch - then a reasonable default would be from HEAD^ to HEAD.
	ref, err := repo.Reference(plumbing.HEAD, true)
	if err != nil {
		return "", "", errors.Wrapf(err, "resolving HEAD ref")
	}
	headName := ref.Name().Short()
	if headName == defaultBranch {
		return defaultBranch + "^", defaultBranch, nil
	}

	return defaultBranch, "HEAD", nil
}

// determineDefaultBranch detects 'default branch' on this repository.
//
// NOTE: Default branch per repository (except for init.defaultBranch config) is a remote-specific config
// and cannot be determined from local repository.
// cf. https://stackoverflow.com/a/70080259
func determineDefaultBranch(repo *git.Repository) (string, error) {
	// Fast path: Check if 'main' or 'master' is present in this repository.
	ref, err := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	if err == nil {
		return ref.Name().Short(), nil
	}
	ref, err = repo.Reference(plumbing.NewBranchReferenceName("master"), true)
	if err == nil {
		return ref.Name().Short(), nil
	}

	// Fallback to listing remote refs and finding HEAD from the result.
	remote, err := repo.Remote("origin")
	if err != nil {
		return "", errors.Wrapf(err, "resolving remote origin")
	}
	slog.Info("Fetching origin to determine default branch...")
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "listing refs from origin")
	}
	for _, ref := range refs {
		if ref.Type() == plumbing.SymbolicReference && ref.Name().String() == "HEAD" {
			// We found refs/heads/origin/HEAD - its symbolic ref target is the default branch.
			defaultBranch := ref.Target().Short()
			return defaultBranch, nil
		}
	}
	return "", errors.New("origin did not contain HEAD symbolic ref, something could be off here?")
}

const worktreeRef = "WORKTREE"

func resolveTree(repo *git.Repository, ref string) (domain.Tree, error) {
	// Special refs
	if ref == worktreeRef {
		return domain.NewGoGitWorkTree(repo)
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
	return domain.NewGoGitCommitTree(commit, ref), nil
}

func reportClones(cloneSets []*domain.CloneSet) {
	// If all clones in a set went through some changes, no need to notify
	cloneSets = lo.Filter(cloneSets, func(cs *domain.CloneSet, _ int) bool { return len(cs.Missing) > 0 })

	// Report the findings
	missingChanges := lo.SumBy(cloneSets, func(set *domain.CloneSet) int { return len(set.Missing) })
	if missingChanges == 0 {
		slog.Info(fmt.Sprintf("No clones are missing consistent change."))
	} else {
		slog.Info(fmt.Sprintf("%d clone(s) are likely missing consistent change.", missingChanges))
	}

	printer := getPrinter()
	out := printer.PrintClones(cloneSets)
	fmt.Print(string(out))

	// If any inconsistent changes are found, exit with specified code
	if len(cloneSets) > 0 && failCode != 0 {
		os.Exit(failCode)
	}
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
