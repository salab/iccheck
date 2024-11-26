package cli

import (
	"fmt"
	"github.com/samber/lo"
	"runtime/debug"
)

var (
	version  = "SNAPSHOT"
	revision string
	dirty    bool
	time     string
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			revision = setting.Value[:7]
		}
		if setting.Key == "vcs.modified" {
			dirty = setting.Value == "true"
		}
		if setting.Key == "vcs.time" {
			time = setting.Value
		}
	}
}

func GetFormattedVersion() string {
	return fmt.Sprintf("%s (%s%s, %s)", version, revision, lo.Ternary(dirty, "+dirty", ""), time)
}
