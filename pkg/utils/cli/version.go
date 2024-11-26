package cli

import (
	"fmt"
	"github.com/samber/lo"
	"runtime/debug"
)

var (
	version  string
	revision string
	dirty    bool
	time     string
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if version == "" {
		version = info.Main.Version
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
	revisionMeta := revision +
		lo.Ternary(dirty, "+dirty", "") +
		lo.Ternary(time != "", ", "+time, "")
	if revisionMeta == "" {
		return version
	} else {
		return fmt.Sprintf("%s (%s)", version, revisionMeta)
	}
}
