package servicemgr

import (
	"crypto/sha1"
	"fmt"
)

//
// Utilities
//

// NOTE: Changes to pretty much everything here break assumptions about the state of the world, so it's likely that
// an update will need to delete and re-create deployments.

func (mgr *serviceManager) serviceName() string {
	h := sha1.New()
	h.Write([]byte(mgr.action.Name))
	bs := h.Sum(nil)
	normalized := fmt.Sprintf("%x", bs)

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
