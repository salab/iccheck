package cmd

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/search"
	"github.com/salab/iccheck/pkg/utils/cli"
	"github.com/salab/iccheck/pkg/utils/files"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"path/filepath"
	"time"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "A low-level command to search for code clones",
	Long: fmt.Sprintf(`ICCheck %v
search is a low-level command to search for code clones.
`, cli.GetFormattedVersion()),
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

		searchTree, err := resolveTree(repo, searchRef)
		if err != nil {
			return errors.Wrapf(err, "resolving search tree")
		}

		query := &domain.Source{
			// Sanitize query filename
			Filename: filepath.Clean(queryFile),
			StartL:   queryStartL,
			EndL:     queryEndL,
		}
		// Quick check for query file & line existence
		{
			contents, err := files.ReadAll(searchTree.Reader(query.Filename))
			if err != nil {
				return errors.Wrapf(err, "opening file %v, does it exist?", query.Filename)
			}
			lines := files.Lines(contents)
			if queryStartL > queryEndL {
				return errors.New("invalid flags: start-line is larger than end-line")
			}
			if queryStartL <= 0 || queryEndL <= 0 {
				return errors.New("invalid flags: start-line and end-line must be positive integers")
			}
			if queryEndL > len(lines) {
				return fmt.Errorf("invalid flags: end-line is greater than the file contents (%v lines)", len(lines))
			}
		}

		// Search for clones and report
		cloneSets, err := search.Search(ctx, algorithm, []*domain.Source{query}, searchTree, ignore)
		if err != nil {
			return err
		}
		reportClones(cloneSets)
		return nil
	},
}

var (
	searchRef string

	queryFile   string
	queryStartL int
	queryEndL   int
)

func init() {
	// NOTE: We could allow searching raw directories via any query file (even outside the search tree),
	// but unless we find a use-case it may not be worth implementing.
	searchCmd.Flags().StringVar(&searchRef, "ref", "HEAD", `Git ref to search against.
Can accept special value "WORKTREE" to specify the current worktree.`)

	searchCmd.Flags().StringVar(&queryFile, "file", "", "Path to query file (relative to git root).")
	lo.Must0(searchCmd.MarkFlagRequired("file"))
	searchCmd.Flags().IntVar(&queryStartL, "start-line", 0, "Starting line number of query file.")
	lo.Must0(searchCmd.MarkFlagRequired("start-line"))
	searchCmd.Flags().IntVar(&queryEndL, "end-line", 0, "Ending line number of query file.")
	lo.Must0(searchCmd.MarkFlagRequired("end-line"))
}
