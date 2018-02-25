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
	return fmt.Sprintf("action-%s", mgr.actionNameHash())
}

func (mgr *serviceManager) actionNameHash() string {
	h := sha1.New()
	h.Write([]byte(mgr.action.Name))
	bs := h.Sum(nil)
	normalized := fmt.Sprintf("%x", bs)

	return normalized
}

func (mgr *serviceManager) serviceLabels() map[string]string {
	// If you change things here, you may need to update the deployment counterpart as well.
	actionId := fmt.Sprintf("action.twine.ai/%s", mgr.actionNameHash())
	return map[string]string{
		"app":          "actionserver",
		"tier":         "actions",
		"twine-action": mgr.action.Name,
		actionId:       "true",
	}
}
