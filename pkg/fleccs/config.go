package fleccs

const (
	DefaultContextLines        = 4
	DefaultSimilarityThreshold = 0.7
)

type config struct {
	contextLines        int
	similarityThreshold float64
}

func defaultConfig() *config {
	return &config{
		contextLines:        DefaultContextLines,
		similarityThreshold: DefaultSimilarityThreshold,
	}
}

func applyConfig(options ...ConfigFunc) *config {
	c := defaultConfig()
	for _, option := range options {
		option(c)
	}
	return c
}

type ConfigFunc func(c *config)

func WithContextLines(lines int) ConfigFunc {
	return func(c *config) {
		c.contextLines = lines
	}
}

func WithSimilarityThreshold(threshold float64) ConfigFunc {
	return func(c *config) {
		c.similarityThreshold = threshold
	}
}
