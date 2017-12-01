package servicemgr

import (
	"fmt"
	"strings"
)

//
// Utilities
//

// NOTE: Changes to pretty much everything here break assumptions about the state of the world, so it's likely that
// an update will need to delete and re-create deployments.

func (mgr *serviceManager) serviceName() string {
	normalized := strings.Replace(mgr.action.Name, ".", "-dot-", -1)
	return fmt.Sprintf("action-%s", normalized)
}

func (mgr *serviceManager) serviceLabels() map[string]string {
	// If you change things here, you may need to update the deployment counterpart as well.
	return map[string]string{
		"app":          "actionserver",
		"tier":         "actions",
		"twine-action": mgr.action.Name,
	}
}
