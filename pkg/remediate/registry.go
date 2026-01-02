package remediate

import (
	"fmt"
	"sync"
)

var (
	remediatorRegistry     = make(map[string]Remediator)
	remediatorRegistryLock sync.RWMutex
)

// Register registers a remediator for an action
func Register(remediator Remediator) {
	remediatorRegistryLock.Lock()
	defer remediatorRegistryLock.Unlock()

	action := remediator.Action()
	remediatorRegistry[action] = remediator
}

// Get retrieves a remediator for an action
func Get(action string) (Remediator, error) {
	remediatorRegistryLock.RLock()
	defer remediatorRegistryLock.RUnlock()

	remediator, exists := remediatorRegistry[action]
	if !exists {
		return nil, fmt.Errorf("no remediator registered for action %q", action)
	}

	return remediator, nil
}

// List returns all registered action names
func List() []string {
	remediatorRegistryLock.RLock()
	defer remediatorRegistryLock.RUnlock()

	actions := make([]string, 0, len(remediatorRegistry))
	for action := range remediatorRegistry {
		actions = append(actions, action)
	}
	return actions
}

