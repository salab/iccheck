package cli

import (
	"fmt"
)

var (
	version  = "dev"
	revision = "local"
)

func GetVersion() (_version, _revision string) {
	return version, revision
}

func GetFormattedVersion() string {
	return fmt.Sprintf("%s (%s)", version, revision)
}
