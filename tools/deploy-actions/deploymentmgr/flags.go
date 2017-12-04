package deploymentmgr

import (
	"log"
	"os/exec"
	"strings"

	"github.com/namsral/flag"
)

var (
	FlagActionServerVersion        string
	FlagActionServerImageName      string
	FlagActionServerSetupImageName string
)

func init() {
	getServerVersion()

	flag.StringVar(
		&FlagActionServerVersion, "action-server-version", getServerVersion(),
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

func getServerVersion() string {
	cmd := `gcloud container images list-tags gcr.io/twine-180301/actionserver --filter="tags!=latest" --limit=1 | tail -n 1 | awk '{ print $2; }'`
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatalf("Unable to get server version: %s", err)
	}
	rawTags := strings.Split(strings.TrimSpace(string(out)), ",")
	for _, tag := range rawTags {
		normalized := strings.TrimSpace(tag)
		if normalized != "latest" {
			return normalized
		}
	}

	return ""
}
