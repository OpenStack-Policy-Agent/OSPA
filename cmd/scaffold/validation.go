package main

import (
	"fmt"
	"strings"
)

// OpenStackServiceRegistry contains known OpenStack services and their resources
// Based on OpenStack official documentation and API references
var OpenStackServiceRegistry = map[string]ServiceInfo{
	"nova": {
		ServiceType: "compute",
		DisplayName: "Nova",
		Resources: map[string]ResourceInfo{
			"instance": {Description: "Server instances"},
			"keypair":  {Description: "SSH keypairs"},
			"server":   {Description: "Server instances (alias for instance)"},
			"flavor":   {Description: "Flavor definitions"},
			"hypervisor": {Description: "Hypervisor information"},
		},
	},
	"neutron": {
		ServiceType: "network",
		DisplayName: "Neutron",
		Resources: map[string]ResourceInfo{
			"security_group":      {Description: "Security groups"},
			"security_group_rule": {Description: "Security group rules"},
			"floating_ip":         {Description: "Floating IP addresses"},
			"network":             {Description: "Networks"},
			"subnet":              {Description: "Subnets"},
			"port":                {Description: "Ports"},
			"router":              {Description: "Routers"},
			"loadbalancer":        {Description: "Load balancers"},
			"pool":                {Description: "Load balancer pools"},
			"member":              {Description: "Load balancer pool members"},
		},
	},
	"cinder": {
		ServiceType: "volumev3",
		DisplayName: "Cinder",
		Resources: map[string]ResourceInfo{
			"volume":   {Description: "Block storage volumes"},
			"snapshot": {Description: "Volume snapshots"},
			"backup":   {Description: "Volume backups"},
			"qos":      {Description: "Quality of service specifications"},
		},
	},
	"glance": {
		ServiceType: "image",
		DisplayName: "Glance",
		Resources: map[string]ResourceInfo{
			"image":  {Description: "Images"},
			"member": {Description: "Image members"},
		},
	},
	"keystone": {
		ServiceType: "identity",
		DisplayName: "Keystone",
		Resources: map[string]ResourceInfo{
			"user":    {Description: "Users"},
			"role":    {Description: "Roles"},
			"project": {Description: "Projects"},
			"domain":  {Description: "Domains"},
			"group":   {Description: "Groups"},
			"service": {Description: "Services"},
		},
	},
	"heat": {
		ServiceType: "orchestration",
		DisplayName: "Heat",
		Resources: map[string]ResourceInfo{
			"stack":        {Description: "Stacks"},
			"resource":     {Description: "Stack resources"},
			"template":     {Description: "Templates"},
			"snapshot":     {Description: "Stack snapshots"},
		},
	},
	"swift": {
		ServiceType: "object-store",
		DisplayName: "Swift",
		Resources: map[string]ResourceInfo{
			"container": {Description: "Containers"},
			"object":    {Description: "Objects"},
			"account":   {Description: "Accounts"},
		},
	},
	"trove": {
		ServiceType: "database",
		DisplayName: "Trove",
		Resources: map[string]ResourceInfo{
			"instance": {Description: "Database instances"},
			"cluster":  {Description: "Database clusters"},
			"backup":   {Description: "Database backups"},
			"datastore": {Description: "Datastores"},
		},
	},
	"magnum": {
		ServiceType: "container-infra",
		DisplayName: "Magnum",
		Resources: map[string]ResourceInfo{
			"cluster":      {Description: "Container clusters"},
			"cluster_template": {Description: "Cluster templates"},
			"bay":          {Description: "Bays (deprecated)"},
			"baymodel":     {Description: "Bay models (deprecated)"},
		},
	},
	"barbican": {
		ServiceType: "key-manager",
		DisplayName: "Barbican",
		Resources: map[string]ResourceInfo{
			"secret":    {Description: "Secrets"},
			"container": {Description: "Secret containers"},
			"order":     {Description: "Orders"},
		},
	},
	"manila": {
		ServiceType: "sharev2",
		DisplayName: "Manila",
		Resources: map[string]ResourceInfo{
			"share":        {Description: "Shared file systems"},
			"share_snapshot": {Description: "Share snapshots"},
			"share_network": {Description: "Share networks"},
			"share_server":  {Description: "Share servers"},
		},
	},
	"ironic": {
		ServiceType: "baremetal",
		DisplayName: "Ironic",
		Resources: map[string]ResourceInfo{
			"node":    {Description: "Bare metal nodes"},
			"port":    {Description: "Node ports"},
			"driver":  {Description: "Drivers"},
			"chassis": {Description: "Chassis"},
		},
	},
	"designate": {
		ServiceType: "dns",
		DisplayName: "Designate",
		Resources: map[string]ResourceInfo{
			"zone":    {Description: "DNS zones"},
			"recordset": {Description: "DNS recordsets"},
			"record":  {Description: "DNS records"},
		},
	},
	"octavia": {
		ServiceType: "load-balancer",
		DisplayName: "Octavia",
		Resources: map[string]ResourceInfo{
			"loadbalancer": {Description: "Load balancers"},
			"listener":     {Description: "Listeners"},
			"pool":         {Description: "Pools"},
			"member":       {Description: "Pool members"},
			"healthmonitor": {Description: "Health monitors"},
		},
	},
	"senlin": {
		ServiceType: "clustering",
		DisplayName: "Senlin",
		Resources: map[string]ResourceInfo{
			"cluster":  {Description: "Clusters"},
			"profile":  {Description: "Profiles"},
			"node":     {Description: "Nodes"},
			"policy":   {Description: "Policies"},
		},
	},
	"zaqar": {
		ServiceType: "messaging",
		DisplayName: "Zaqar",
		Resources: map[string]ResourceInfo{
			"queue":    {Description: "Message queues"},
			"message":  {Description: "Messages"},
			"subscription": {Description: "Subscriptions"},
		},
	},
}

// ServiceInfo contains information about an OpenStack service
type ServiceInfo struct {
	ServiceType string
	DisplayName string
	Resources   map[string]ResourceInfo
}

// ResourceInfo contains information about a resource type
type ResourceInfo struct {
	Description string
}

// ValidateService checks if a service exists in OpenStack
func ValidateService(serviceName string) error {
	serviceName = strings.ToLower(serviceName)
	_, exists := OpenStackServiceRegistry[serviceName]
	if !exists {
		available := make([]string, 0, len(OpenStackServiceRegistry))
		for name := range OpenStackServiceRegistry {
			available = append(available, name)
		}
		return fmt.Errorf("service %q is not a known OpenStack service. Available services: %v", serviceName, available)
	}
	return nil
}

// ValidateResources checks if resources exist for a given service
func ValidateResources(serviceName string, resources []string) error {
	serviceName = strings.ToLower(serviceName)
	serviceInfo, exists := OpenStackServiceRegistry[serviceName]
	if !exists {
		return ValidateService(serviceName)
	}

	var invalidResources []string
	for _, resource := range resources {
		resource = strings.ToLower(resource)
		if _, exists := serviceInfo.Resources[resource]; !exists {
			invalidResources = append(invalidResources, resource)
		}
	}

	if len(invalidResources) > 0 {
		available := make([]string, 0, len(serviceInfo.Resources))
		for name := range serviceInfo.Resources {
			available = append(available, name)
		}
		return fmt.Errorf("invalid resources for service %q: %v. Available resources: %v", 
			serviceName, invalidResources, available)
	}

	return nil
}

// GetServiceInfo returns information about a service
func GetServiceInfo(serviceName string) (ServiceInfo, error) {
	serviceName = strings.ToLower(serviceName)
	info, exists := OpenStackServiceRegistry[serviceName]
	if !exists {
		return ServiceInfo{}, fmt.Errorf("service %q not found", serviceName)
	}
	return info, nil
}

// GetServiceType returns the OpenStack service type for a service name
func GetServiceType(serviceName string) (string, error) {
	info, err := GetServiceInfo(serviceName)
	if err != nil {
		return "", err
	}
	return info.ServiceType, nil
}

// GetDisplayName returns the display name for a service
func GetDisplayName(serviceName string) (string, error) {
	info, err := GetServiceInfo(serviceName)
	if err != nil {
		return "", err
	}
	return info.DisplayName, nil
}

// ListServices returns all available OpenStack services
func ListServices() []string {
	services := make([]string, 0, len(OpenStackServiceRegistry))
	for name := range OpenStackServiceRegistry {
		services = append(services, name)
	}
	return services
}

// ListResources returns all available resources for a service
func ListResources(serviceName string) ([]string, error) {
	serviceName = strings.ToLower(serviceName)
	serviceInfo, exists := OpenStackServiceRegistry[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %q not found", serviceName)
	}
	
	resources := make([]string, 0, len(serviceInfo.Resources))
	for name := range serviceInfo.Resources {
		resources = append(resources, name)
	}
	return resources, nil
}

