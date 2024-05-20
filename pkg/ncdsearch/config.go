package ncdsearch

type TokenizeFunc = func(s []byte) [][]byte

func DefaultTokenizeFunc(s []byte) [][]byte {
	ret := make([][]byte, len(s))
	for i := range s {
		ret[i] = s[i : i+1]
	}
	return ret
}

const DefaultOverlapNGram = 5
const DefaultFilterThreshold = 0.5
const DefaultSearchThreshold = 0.5
const DefaultWindowSizeMultiplier = 1.2

type config struct {
	tokenize        TokenizeFunc
	overlapNGram    int
	filterThreshold float64
	searchThreshold float64
	windowSizeMult  float64
}

func defaultConfig() *config {
	return &config{
		tokenize:        DefaultTokenizeFunc,
		overlapNGram:    DefaultOverlapNGram,
		filterThreshold: DefaultFilterThreshold,
		searchThreshold: DefaultSearchThreshold,
		windowSizeMult:  DefaultWindowSizeMultiplier,
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

func WithTokenizeFunc(tokenize TokenizeFunc) ConfigFunc {
	return func(c *config) {
		c.tokenize = tokenize
	}
}

func WithOverlapNGram(nGram int) ConfigFunc {
	return func(c *config) {
		c.overlapNGram = nGram
	}
}

func WithFilterThreshold(filterThreshold float64) ConfigFunc {
	return func(c *config) {
		c.filterThreshold = filterThreshold
	}
}

func WithSearchThreshold(searchThreshold float64) ConfigFunc {
	return func(c *config) {
		c.searchThreshold = searchThreshold
	}
}

func WithWindowSizeMultiplier(windowSizeMult float64) ConfigFunc {
	return func(c *config) {
		c.windowSizeMult = windowSizeMult
	}
}
