package search

import (
	"context"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/fleccs"
	"github.com/salab/iccheck/pkg/ncdsearch"
	"github.com/salab/iccheck/pkg/utils/ds"
)

type AlgorithmFunc = func(
	ctx context.Context,
	sourceTree domain.Searcher,
	sources []*domain.Source,
	searchTree domain.Searcher,
	ignore domain.IgnoreRules,
) ([]*domain.Clone, error)

var algorithms = map[string]AlgorithmFunc{
	"fleccs":    fleccsSearchMulti,
	"ncdsearch": ncdSearchReImpl,
}

func ncdSearchReImpl(
	ctx context.Context,
	sourceTree domain.Searcher,
	sources []*domain.Source,
	searchTree domain.Searcher,
	ignore domain.IgnoreRules,
) ([]*domain.Clone, error) {
	queries := ds.Map(sources, func(s *domain.Source) *ncdsearch.Query {
		return &ncdsearch.Query{
			Filename: s.Filename,
			StartL:   s.StartL,
			EndL:     s.EndL,
		}
	})

	clones, err := ncdsearch.Search(
		ctx,
		sourceTree,
		queries,
		searchTree,
		ignore,
		ncdsearch.WithSearchThreshold(0.3),
	)
	if err != nil {
		return nil, err
	}

	return ds.Map(clones, func(c *ncdsearch.Clone) *domain.Clone {
		return &domain.Clone{
			Filename: c.Filename,
			StartL:   c.StartLine,
			EndL:     c.EndLine,
			Distance: c.Distance,
		}
	}), nil
}

func fleccsSearchMulti(
	ctx context.Context,
	sourceTree domain.Searcher,
	sources []*domain.Source,
	searchTree domain.Searcher,
	ignore domain.IgnoreRules,
) ([]*domain.Clone, error) {
	queries := ds.Map(sources, func(s *domain.Source) *fleccs.Query {
		return &fleccs.Query{
			Filename: s.Filename,
			StartL:   s.StartL,
			EndL:     s.EndL,
		}
	})

	candidates, err := fleccs.Search(
		ctx,
		sourceTree,
		queries,
		searchTree,
		ignore,
	)
	if err != nil {
		return nil, err
	}

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
	}), nil
}
