package search

import "testing"

func TestSearch(t *testing.T) {
	_, err := Search("/home/moto/projects/salab/iccheck/data/NeoShowcase", "e79b8d658dde4c68186c4dfdf887183db0093430~", "e79b8d658dde4c68186c4dfdf887183db0093430")
	if err != nil {
		t.Error(err)
	}
}
