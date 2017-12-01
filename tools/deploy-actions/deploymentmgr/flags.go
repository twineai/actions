package deploymentmgr

import (
	"github.com/namsral/flag"
)

var (
	FlagActionServerVersion        string
	FlagActionServerImageName      string
	FlagActionServerSetupImageName string
)

func init() {
	flag.StringVar(
		&FlagActionServerVersion, "action-server-version", "v0.0.14",
		"The version of the action server to use. If empty latest will be used")

	flag.StringVar(
		&FlagActionServerImageName, "action-server-image-name", "gcr.io/twine-180301/actionserver",
		"The name of the action server image to use")

	flag.StringVar(
		&FlagActionServerSetupImageName,
		"action-server-setup-image-name",
		"gcr.io/twine-180301/actionserver-setup",
		"The name of the action server setup image to use")
}
