package main_test

import (
	"github.com/salab/iccheck/cmd"
	"github.com/samber/lo"
	"os"
	"runtime/pprof"
	"testing"
)

// For cpu profiling and algorithm optimization
func TestSearch(t *testing.T) {
	homedir := lo.Must(os.UserHomeDir())
	_ = os.Mkdir(homedir+"/pprof", 0777)
	f := lo.Must(os.Create(homedir + "/pprof/cpu.pprof"))
	defer f.Close()
	lo.Must0(pprof.StartCPUProfile(f))
	defer pprof.StopCPUProfile()

	err := cmd.RootCmd.RunE(cmd.RootCmd, []string{})
	if err != nil {
		t.Fatal(err)
	}
}
