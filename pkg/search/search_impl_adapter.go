package search

import (
	"context"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/fleccs"
	"github.com/salab/iccheck/pkg/ncdsearch"
	"github.com/salab/iccheck/pkg/utils/ds"
	"strconv"
)

type AlgorithmFunc = func(
	ctx context.Context,
	sourceTree domain.Searcher,
	sources []*domain.Source,
	searchTree domain.Searcher,
	c *Config,
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
	c *Config,
) ([]*domain.Clone, error) {
	queries := ds.Map(sources, func(s *domain.Source) *ncdsearch.Query {
		return &ncdsearch.Query{
			Filename: s.Filename,
			StartL:   s.StartL,
			EndL:     s.EndL,
		}
	})

	var opts []ncdsearch.ConfigFunc
	opts = append(opts, ncdsearch.WithSearchThreshold(0.3))

	// Algorithm parameters
	if v, ok := c.AlgoParams["overlap-ngram"]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse overlap-ngram")
		}
		opts = append(opts, ncdsearch.WithOverlapNGram(i))
	}
	if v, ok := c.AlgoParams["filter-threshold"]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse filter-threshold")
		}
		opts = append(opts, ncdsearch.WithFilterThreshold(f))
	}
	if v, ok := c.AlgoParams["threshold"]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse threshold")
		}
		opts = append(opts, ncdsearch.WithSearchThreshold(f))
	}
	if v, ok := c.AlgoParams["window-size-mult"]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse window-size-mult")
		}
		opts = append(opts, ncdsearch.WithWindowSizeMultiplier(f))
	}

	clones, err := ncdsearch.Search(
		ctx,
		sourceTree,
		queries,
		searchTree,
		c.Ignore,
		opts...,
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
	c *Config,
) ([]*domain.Clone, error) {
	queries := ds.Map(sources, func(s *domain.Source) *fleccs.Query {
		return &fleccs.Query{
			Filename: s.Filename,
			StartL:   s.StartL,
			EndL:     s.EndL,
		}
	})

	var opts []fleccs.ConfigFunc

	// Algorithm parameters
	if v, ok := c.AlgoParams["threshold"]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse threshold")
		}
		opts = append(opts, fleccs.WithSimilarityThreshold(f))
	}
	if v, ok := c.AlgoParams["context-lines"]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse context-lines")
		}
		opts = append(opts, fleccs.WithContextLines(i))
	}

	candidates, err := fleccs.Search(
		ctx,
		sourceTree,
		queries,
		searchTree,
		c.Ignore,
		opts...,
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
