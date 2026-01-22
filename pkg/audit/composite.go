package audit

import (
	"fmt"
	"sync"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/discovery"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// CompositeAuditor evaluates composite rules across multiple resource types.
type CompositeAuditor interface {
	Service() string
	Check(resources map[string][]discovery.Job, rule *policy.CompositeRule) (*Result, error)
	Fix(resources map[string][]discovery.Job, rule *policy.CompositeRule) error
}

var (
	compositeRegistry     = make(map[string]CompositeAuditor)
	compositeRegistryLock sync.RWMutex
)

// RegisterComposite registers a composite auditor for a service.
func RegisterComposite(auditor CompositeAuditor) error {
	if auditor == nil {
		return fmt.Errorf("composite auditor is nil")
	}
	service := auditor.Service()
	if service == "" {
		return fmt.Errorf("composite auditor has empty service")
	}
	compositeRegistryLock.Lock()
	defer compositeRegistryLock.Unlock()
	if _, exists := compositeRegistry[service]; exists {
		return fmt.Errorf("composite auditor for service %q already registered", service)
	}
	compositeRegistry[service] = auditor
	return nil
}

// GetComposite returns the composite auditor for a service.
func GetComposite(service string) (CompositeAuditor, bool) {
	compositeRegistryLock.RLock()
	defer compositeRegistryLock.RUnlock()
	auditor, ok := compositeRegistry[service]
	return auditor, ok
}
