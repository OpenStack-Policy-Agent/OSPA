package discovery

// Job represents a generic resource to audit
type Job struct {
	ResourceType string
	ResourceID   string
	Resource     interface{} // Service-specific resource struct
	Service      string
	ProjectID    string
}
