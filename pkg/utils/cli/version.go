package cli

import "fmt"

var (
	_version  = "dev"
	_revision = "local"
)

func SetVersion(version, revision string) {
	_version, _revision = version, revision
}

func GetVersion() (version, revision string) {
	return _version, _revision
}

func GetFormattedVersion() string {
	return fmt.Sprintf("%s (%s)", _version, _revision)
}
