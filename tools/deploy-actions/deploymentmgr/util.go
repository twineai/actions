package deploymentmgr

import (
	"crypto/sha1"
	"fmt"

	"github.com/twineai/actions/tools/deploy-actions/action"
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
	if len(mgr.actions) > 1 {
		return "actionserver"
	} else {
		return fmt.Sprintf("action-%s", mgr.actionNameHash(mgr.actions[0]))
	}
}

func (mgr *deploymentManager) actionNameHash(a action.Action) string {
	h := sha1.New()
	h.Write([]byte(a.Name))
	bs := h.Sum(nil)
	normalized := fmt.Sprintf("%x", bs)

	return normalized
}

func (mgr *deploymentManager) actionVolumeName() string {
	return "action-dir"
}

func (mgr *deploymentManager) actionVolumePath() string {
	return "/user_code"
}

func (mgr *deploymentManager) deploymentLabels() map[string]string {
	// If you change things here, you may need to update the service counterpart as well.

	result := map[string]string{
		"app":  "actionserver",
		"tier": "actions",
	}

	for _, action := range mgr.actions {
		actionId := fmt.Sprintf("action.twine.ai/%s", mgr.actionNameHash(action))
		result[actionId] = "true"
	}

	return result
}
