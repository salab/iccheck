package main_test

import (
	"github.com/salab/iccheck/cmd"
	"os"
	"testing"
)

// For cpu profiling and algorithm optimization
func TestSearch(t *testing.T) {
	os.Args = []string{"iccheck"}
	err := cmd.RootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}
