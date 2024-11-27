package main_test

import (
	"github.com/salab/iccheck/cmd"
	"testing"
)

// For cpu profiling and algorithm optimization
func TestSearch(t *testing.T) {
	err := cmd.RootCmd.RunE(cmd.RootCmd, []string{})
	if err != nil {
		t.Fatal(err)
	}
}
