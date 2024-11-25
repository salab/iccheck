package domain

import (
	"errors"
	"fmt"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/salab/iccheck/pkg/utils/files"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// defaultIgnoreConfigs lists rules (that are pretty much safe to assume) for some languages.
// To contributors: Feel free to add more to this default config.
var defaultIgnoreConfigs = IgnoreConfigs{
	{
		Files: []string{"\\.go$"},
		Patterns: []string{
			`^package .+$`,
			`^import (.+ )?".+"$`,
			`^import \(\n(\s+(.+ )?".+"\n)*\)$`,
		},
	},
	{
		Files: []string{"\\.java$"},
		Patterns: []string{
			`^package .+;$`,
			`^import .+;$`,
		},
	},
	{
		Files: []string{"\\.m?[jt]s$"},
		Patterns: []string{
			`^import`, // ESM import
		},
	},
}

func ReadIgnoreRules(repoDir string, cliOptions []string, disableDefault bool) (IgnoreRules, error) {
	allConfigs := make(IgnoreConfigs, 0)
	if !disableDefault {
		allConfigs = append(allConfigs, defaultIgnoreConfigs...)
	}

	// Check if ignore file is present in the following locations:
	// 1. ${repoDir}/.iccheckignore.{yaml,yml}
	// 2. ~/.config/.iccheckignore.{yaml,yml}
	const baseFileName = ".iccheckignore"
	paths := []string{
		filepath.Join(repoDir, baseFileName+".yaml"),
		filepath.Join(repoDir, baseFileName+".yml"),
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".config", baseFileName+".yaml"))
		paths = append(paths, filepath.Join(homeDir, ".config", baseFileName+".yml"))
	}
	for _, path := range paths {
		if f, err := os.Open(path); err == nil {
			var configs IgnoreConfigs
			if err = yaml.NewDecoder(f).Decode(&configs); err != nil {
				return nil, fmt.Errorf("decoding %s: %w", path, err)
			}
			allConfigs = append(allConfigs, configs...)
		}
		// Ignore os.Open() error - file might not exist
	}

	// Parse CLI options, if any
	for _, cliOption := range cliOptions {
		config, err := readIgnoreCLIOption(cliOption)
		if err != nil {
			return nil, err
		}
		allConfigs = append(allConfigs, config)
	}

	return allConfigs.Compile()
}

type IgnoreConfigs []*IgnoreConfig

type IgnoreConfig struct {
	Files    []string `yaml:"files"`
	Patterns []string `yaml:"patterns"`
}

func readIgnoreCLIOption(opt string) (*IgnoreConfig, error) {
	parts := strings.Split(opt, ":")
	if len(parts) != 1 && len(parts) != 2 {
		return nil, fmt.Errorf("invalid ignore format: %s (specify file regexp path, or file regexp path and pattern regexp split by ':')", opt)
	}
	if len(parts) == 1 {
		i := IgnoreConfig{
			Files: []string{parts[0]},
		}
		return &i, nil
	} else {
		i := IgnoreConfig{
			Files:    []string{parts[0]},
			Patterns: []string{parts[1]},
		}
		return &i, nil
	}
}

func (i *IgnoreConfig) Compile() (*IgnoreRule, error) {
	if len(i.Files) == 0 && len(i.Patterns) == 0 {
		return nil, errors.New("no files or patterns specified")
	}

	var ret IgnoreRule
	var err error
	ret.files, err = ds.MapError(i.Files, regexp.Compile)
	if err != nil {
		return nil, err
	}
	// Enable multi-line mode
	i.Patterns = ds.Map(i.Patterns, func(p string) string {
		if strings.HasPrefix(p, "(?m)") {
			return p
		}
		return "(?m)" + p
	})
	ret.patterns, err = ds.MapError(i.Patterns, regexp.Compile)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (i IgnoreConfigs) Compile() (IgnoreRules, error) {
	return ds.MapError(i, (*IgnoreConfig).Compile)
}

type IgnoreRules []*IgnoreRule

type IgnoreRule struct {
	files    []*regexp.Regexp
	patterns []*regexp.Regexp
}

func (i *IgnoreRule) matchFile(path string) bool {
	if len(i.files) == 0 {
		return true
	}
	for _, f := range i.files {
		if !f.MatchString(path) {
			return false
		}
	}
	return true
}

// matchContents returns 0-indexed line numbers whose contents match the ignore patterns.
func (i *IgnoreRule) matchContents(contents []byte) (ignoreLines map[int]struct{}) {
	lineIndices := files.LineIndices(contents)
	ignoreLines = make(map[int]struct{})

	toLineNumber := func(index int) int {
		return sort.Search(len(lineIndices), func(lineIdx int) bool {
			return index < lineIndices[lineIdx]
		}) - 1
	}

	for _, p := range i.patterns {
		matches := p.FindAllIndex(contents, -1)
		for _, match := range matches {
			start, end := match[0], match[1]
			startLine := toLineNumber(start)
			endLine := toLineNumber(end)
			for l := startLine; l <= endLine; l++ {
				ignoreLines[l] = struct{}{}
			}
		}
	}

	return
}

// Match checks the file and its contents to ignore.
// If skipEntireFile is true, callers are expected to skip this entire file (ignoreRule is nil).
// Otherwise, callers are expected to call IgnoreLineRule.CanSkip() method according to its doc to
// check file ignore pattern.
func (i IgnoreRules) Match(path string, contents []byte) (skipEntireFile bool, ignoreRule *IgnoreLineRule) {
	// Check if any patterns match the whole file first
	instances := make([]*IgnoreRule, 0)
	for _, instance := range i {
		matchFile := instance.matchFile(path)
		if !matchFile {
			continue
		}
		if len(instance.patterns) == 0 {
			return true, nil
		}
		instances = append(instances, instance)
	}

	// If no rules specify whole file skip, check for pattern skip next
	var mergedIgnoreLines map[int]struct{}
	for _, instance := range instances {
		ignoreLines := instance.matchContents(contents)
		if mergedIgnoreLines == nil {
			mergedIgnoreLines = ignoreLines
		} else {
			mergedIgnoreLines = ds.MergeMap(mergedIgnoreLines, ignoreLines)
		}
	}

	return false, &IgnoreLineRule{
		IgnoreLines: mergedIgnoreLines,
		safeUntil:   -1,
	}
}

type IgnoreLineRule struct {
	IgnoreLines map[int]struct{}
	safeUntil   int
}

// CanSkip determines if window size starting from startLine can be skipped.
//
// Caller expectations:
//   - Callers are expected to monotonically increase startLine.
//   - Next call to CanSkip() starts from skipUntil+1, if canSkip is true.
func (l *IgnoreLineRule) CanSkip(startLine int, windowSize int) (canSkip bool, skipUntil int) {
	searchUpper := startLine + windowSize - 1
	searchLower := max(startLine, l.safeUntil+1)
	for i := searchUpper; i >= searchLower; i-- {
		if _, ok := l.IgnoreLines[i]; ok {
			return true, i + windowSize - 1
		}
	}
	l.safeUntil = searchUpper
	return false, -1
}
