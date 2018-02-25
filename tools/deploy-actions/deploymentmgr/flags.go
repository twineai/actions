package deploymentmgr

import (
	"github.com/namsral/flag"
)

var (
	FlagActionServerVersion        string
	FlagActionServerImageName      string
	FlagActionServerSetupImageName string
	FlagInstanceCount              int
)

func init() {
	flag.StringVar(
		&FlagActionServerVersion, "action-server-version", "",
		"The version of the action server to use. If empty latest will be used")

	flag.StringVar(
		&FlagActionServerImageName, "action-server-image-name", "gcr.io/twine-180301/actionserver",
		"The name of the action server image to use")

	flag.StringVar(
		&FlagActionServerSetupImageName,
		"action-server-setup-image-name",
		"gcr.io/twine-180301/actionserver-setup",
		"The name of the action server setup image to use")

	flag.IntVar(
		&FlagInstanceCount, "instance-count", 1,
		"In single-instance mode, the number of instances to instantiate")

}
