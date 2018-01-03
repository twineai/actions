package deploymentmgr

import (
	"crypto/sha1"
	"fmt"
)

func int32Ptr(i int32) *int32 { return &i }

//
// Utilities
//

// NOTE: Changes to pretty much everything here break assumptions about the state of the world, so it's likely that
// an update will need to delete and re-create deployments.

func (mgr *deploymentManager) imageName() string {
	return fmt.Sprintf("%s:%s", mgr.serverImageName, mgr.serverVersion)
}

func (mgr *deploymentManager) setupImageName() string {
	return fmt.Sprintf("%s:%s", mgr.serverSetupImageName, mgr.serverVersion)
}

func (mgr *deploymentManager) deploymentName() string {
	h := sha1.New()
	h.Write([]byte(mgr.action.Name))
	bs := h.Sum(nil)
	normalized := fmt.Sprintf("%x", bs)

	return fmt.Sprintf("action-%s", normalized)
}

func (mgr *deploymentManager) actionVolumeName() string {
	return "action-dir"
}

func (mgr *deploymentManager) actionVolumePath() string {
	return "/user_code"
}

func (mgr *deploymentManager) deploymentLabels() map[string]string {
	// If you change things here, you may need to update the service counterpart as well.
	return map[string]string{
		"app":          "actionserver",
		"tier":         "actions",
		"twine-action": mgr.action.Name,
	}
}
