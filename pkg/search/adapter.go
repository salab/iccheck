package search

import (
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/fleccs"
	"path/filepath"
	"strings"

	"github.com/salab/iccheck/pkg/ncdsearch"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/salab/iccheck/pkg/utils/files"
)

func ncdSearchOriginal(
	basePath string,
	filename string,
	startL, endL int,
) []domain.Clone {
	res := ncdsearch.SearchOriginal(basePath, filename, startL, endL)
	return ds.Map(res.Result, func(c *ncdsearch.OriginalOutClone) domain.Clone {
		return domain.Clone{
			Filename: strings.TrimPrefix(c.FileName, "/work/./"),
			StartL:   c.StartLine,
			EndL:     c.EndLine,
			Distance: c.Distance,
		}
	})
}

func ncdSearchReImpl(
	basePath string,
	filename string,
	startL, endL int,
) []domain.Clone {
	fullPath := filepath.Join(basePath, filename)
	query := files.ReadFileLines(fullPath, startL, endL)
	clones := ncdsearch.Search(query, basePath, ncdsearch.WithSearchThreshold(0.3))
	return ds.Map(clones, func(c ncdsearch.Clone) domain.Clone {
		return domain.Clone{
			Filename: strings.TrimPrefix(c.Filename, basePath+"/"),
			StartL:   c.StartLine,
			EndL:     c.EndLine,
			Distance: c.Distance,
		}
	})
}

func fleccsSearchMulti(
	sourceTree domain.Tree,
	sources []*domain.Source,
	searchTree domain.Tree,
) []*domain.Clone {
	queries := ds.Map(sources, func(s *domain.Source) *fleccs.Query {
		return &fleccs.Query{
			Filename: s.Filename,
			StartL:   s.StartL,
			EndL:     s.EndL,
		}
	})

	candidates := fleccs.Search(
		sourceTree,
		queries,
		searchTree,
	)

	return ds.Map(candidates, func(c *fleccs.Candidate) *domain.Clone {
		return &domain.Clone{
			Filename: c.Filename,
			StartL:   c.StartLine,
			EndL:     c.EndLine,
			Distance: 1 - c.Similarity,
			Sources: []*domain.Source{{
				Filename: c.Source.Filename,
				StartL:   c.Source.StartL,
				EndL:     c.Source.EndL,
			}},
		}
	})
}
