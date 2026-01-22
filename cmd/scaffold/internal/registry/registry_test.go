package registry

import (
	"strings"
	"testing"
)

func TestValidateService_ValidService(t *testing.T) {
	validServices := []string{"nova", "neutron", "cinder", "glance", "keystone"}

	for _, service := range validServices {
		t.Run(service, func(t *testing.T) {
			if err := ValidateService(service); err != nil {
				t.Errorf("ValidateService(%q) = %v, want nil", service, err)
			}
		})
	}
}

func TestValidateService_InvalidService(t *testing.T) {
	invalidServices := []string{"invalid", "unknown", "test"}

	for _, service := range invalidServices {
		t.Run(service, func(t *testing.T) {
			err := ValidateService(service)
			if err == nil {
				t.Errorf("ValidateService(%q) = nil, want error", service)
			}
			if !strings.Contains(err.Error(), "not a known OpenStack service") {
				t.Errorf("ValidateService(%q) error = %v, want error containing 'not a known OpenStack service'", service, err)
			}
			if !strings.Contains(err.Error(), "Available services:") {
				t.Errorf("ValidateService(%q) error = %v, want error containing 'Available services:'", service, err)
			}
		})
	}
}

func TestValidateService_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		name    string
		service string
		wantErr bool
	}{
		{"lowercase", "nova", false},
		{"uppercase", "NOVA", false},
		{"mixed", "NoVa", false},
		{"title", "Nova", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateService(tc.service)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateService(%q) = %v, want error = %v", tc.service, err, tc.wantErr)
			}
		})
	}
}

func TestValidateResources_ValidResources(t *testing.T) {
	testCases := []struct {
		service   string
		resources []string
	}{
		{"nova", []string{"instance", "keypair"}},
		{"neutron", []string{"security_group", "floating_ip"}},
		{"cinder", []string{"volume", "snapshot"}},
		{"glance", []string{"image", "member"}},
		{"keystone", []string{"user", "role", "project"}},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			if err := ValidateResources(tc.service, tc.resources); err != nil {
				t.Errorf("ValidateResources(%q, %v) = %v, want nil", tc.service, tc.resources, err)
			}
		})
	}
}

func TestValidateResources_InvalidResource(t *testing.T) {
	testCases := []struct {
		service   string
		resources []string
	}{
		{"nova", []string{"invalid_resource"}},
		{"neutron", []string{"unknown"}},
		{"cinder", []string{"fake"}},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			err := ValidateResources(tc.service, tc.resources)
			if err == nil {
				t.Errorf("ValidateResources(%q, %v) = nil, want error", tc.service, tc.resources)
			}
			if !strings.Contains(err.Error(), "invalid resources") {
				t.Errorf("ValidateResources(%q, %v) error = %v, want error containing 'invalid resources'", tc.service, tc.resources, err)
			}
			if !strings.Contains(err.Error(), "Available resources:") {
				t.Errorf("ValidateResources(%q, %v) error = %v, want error containing 'Available resources:'", tc.service, tc.resources, err)
			}
		})
	}
}

func TestValidateResources_MixedValidInvalid(t *testing.T) {
	err := ValidateResources("nova", []string{"instance", "invalid_resource"})
	if err == nil {
		t.Error("ValidateResources with mixed valid/invalid = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid_resource") {
		t.Errorf("ValidateResources error = %v, want error containing 'invalid_resource'", err)
	}
}

func TestValidateResources_EmptyList(t *testing.T) {
	err := ValidateResources("nova", []string{})
	if err != nil {
		t.Errorf("ValidateResources with empty list = %v, want nil", err)
	}
}

func TestGetServiceInfo_ValidService(t *testing.T) {
	testCases := []struct {
		service      string
		expectedType string
		expectedName string
	}{
		{"nova", "compute", "Nova"},
		{"neutron", "network", "Neutron"},
		{"cinder", "volumev3", "Cinder"},
		{"glance", "image", "Glance"},
		{"keystone", "identity", "Keystone"},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			info, err := GetServiceInfo(tc.service)
			if err != nil {
				t.Fatalf("GetServiceInfo(%q) = %v, want nil", tc.service, err)
			}
			if info.ServiceType != tc.expectedType {
				t.Errorf("GetServiceInfo(%q).ServiceType = %q, want %q", tc.service, info.ServiceType, tc.expectedType)
			}
			if info.DisplayName != tc.expectedName {
				t.Errorf("GetServiceInfo(%q).DisplayName = %q, want %q", tc.service, info.DisplayName, tc.expectedName)
			}
		})
	}
}

func TestGetServiceInfo_InvalidService(t *testing.T) {
	info, err := GetServiceInfo("invalid")
	if err == nil {
		t.Error("GetServiceInfo(\"invalid\") = nil, want error")
	}
	if info.DisplayName != "" {
		t.Errorf("GetServiceInfo(\"invalid\").DisplayName = %q, want empty", info.DisplayName)
	}
}

func TestGetServiceType_ValidService(t *testing.T) {
	testCases := []struct {
		service string
		want    string
	}{
		{"nova", "compute"},
		{"neutron", "network"},
		{"cinder", "volumev3"},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			got, err := GetServiceType(tc.service)
			if err != nil {
				t.Fatalf("GetServiceType(%q) = %v, want nil", tc.service, err)
			}
			if got != tc.want {
				t.Errorf("GetServiceType(%q) = %q, want %q", tc.service, got, tc.want)
			}
		})
	}
}

func TestGetDisplayName_ValidService(t *testing.T) {
	testCases := []struct {
		service string
		want    string
	}{
		{"nova", "Nova"},
		{"neutron", "Neutron"},
		{"cinder", "Cinder"},
		{"glance", "Glance"},
		{"keystone", "Keystone"},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			got, err := GetDisplayName(tc.service)
			if err != nil {
				t.Fatalf("GetDisplayName(%q) = %v, want nil", tc.service, err)
			}
			if got != tc.want {
				t.Errorf("GetDisplayName(%q) = %q, want %q", tc.service, got, tc.want)
			}
		})
	}
}

func TestListServices_AllServices(t *testing.T) {
	services := ListServices()

	if len(services) == 0 {
		t.Error("ListServices() returned empty list")
	}

	expectedServices := []string{"nova", "neutron", "cinder", "glance", "keystone"}
	serviceMap := make(map[string]bool)
	for _, s := range services {
		serviceMap[s] = true
	}

	for _, expected := range expectedServices {
		if !serviceMap[expected] {
			t.Errorf("ListServices() missing expected service: %q", expected)
		}
	}
}

func TestListResources_ValidService(t *testing.T) {
	testCases := []struct {
		service      string
		minResources int
	}{
		{"nova", 2},
		{"neutron", 5},
		{"cinder", 2},
		{"glance", 2},
		{"keystone", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			resources, err := ListResources(tc.service)
			if err != nil {
				t.Fatalf("ListResources(%q) = %v, want nil", tc.service, err)
			}
			if len(resources) < tc.minResources {
				t.Errorf("ListResources(%q) returned %d resources, want at least %d", tc.service, len(resources), tc.minResources)
			}
		})
	}
}

func TestListResources_InvalidService(t *testing.T) {
	resources, err := ListResources("invalid")
	if err == nil {
		t.Error("ListResources(\"invalid\") = nil, want error")
	}
	if resources != nil {
		t.Errorf("ListResources(\"invalid\") returned resources = %v, want nil", resources)
	}
}

func TestServiceRegistry_Completeness(t *testing.T) {
	services := ListServices()

	if len(services) == 0 {
		t.Fatal("No services in registry")
	}

	for _, serviceName := range services {
		t.Run(serviceName, func(t *testing.T) {
			info, err := GetServiceInfo(serviceName)
			if err != nil {
				t.Fatalf("GetServiceInfo(%q) = %v", serviceName, err)
			}

			if info.ServiceType == "" {
				t.Errorf("Service %q has empty ServiceType", serviceName)
			}

			if info.DisplayName == "" {
				t.Errorf("Service %q has empty DisplayName", serviceName)
			}

			if len(info.Resources) == 0 {
				t.Errorf("Service %q has no resources", serviceName)
			}

			for resourceName, resourceInfo := range info.Resources {
				if resourceName == "" {
					t.Errorf("Service %q has empty resource name", serviceName)
				}
				if resourceInfo.Description == "" {
					t.Errorf("Service %q resource %q has empty description", serviceName, resourceName)
				}
			}
		})
	}
}

func TestServiceRegistry_NoDuplicates(t *testing.T) {
	services := ListServices()
	serviceMap := make(map[string]bool)

	for _, service := range services {
		if serviceMap[service] {
			t.Errorf("Duplicate service name in registry: %q", service)
		}
		serviceMap[service] = true
	}

	// Check for duplicate resources within each service
	for _, serviceName := range services {
		info, err := GetServiceInfo(serviceName)
		if err != nil {
			continue
		}

		resourceMap := make(map[string]bool)
		for resourceName := range info.Resources {
			if resourceMap[resourceName] {
				t.Errorf("Service %q has duplicate resource: %q", serviceName, resourceName)
			}
			resourceMap[resourceName] = true
		}
	}
}

func TestValidateService_EmptyString(t *testing.T) {
	err := ValidateService("")
	if err == nil {
		t.Error("ValidateService(\"\") = nil, want error")
	}
}

func TestValidateService_Whitespace(t *testing.T) {
	testCases := []string{" nova", "nova ", " nova ", "\tnova", "nova\n"}

	for _, service := range testCases {
		t.Run(service, func(t *testing.T) {
			err := ValidateService(service)
			// Should either error or normalize (trim whitespace)
			if err == nil {
				// If it doesn't error, it should work with trimmed version
				trimmed := strings.TrimSpace(service)
				if err2 := ValidateService(trimmed); err2 != nil {
					t.Errorf("ValidateService(%q) succeeded but ValidateService(%q) failed", service, trimmed)
				}
			}
		})
	}
}

func TestValidateResources_Duplicates(t *testing.T) {
	// Duplicate resources should be handled (either deduplicated or error)
	resources := []string{"instance", "instance", "keypair"}
	err := ValidateResources("nova", resources)
	// Should either succeed (with deduplication) or error
	// For now, we'll just verify it doesn't crash
	if err != nil && !strings.Contains(err.Error(), "invalid resources") {
		t.Logf("ValidateResources with duplicates returned: %v", err)
	}
}

func TestValidateResources_CaseVariations(t *testing.T) {
	testCases := []struct {
		service   string
		resources []string
		wantErr   bool
	}{
		{"nova", []string{"Instance", "INSTANCE", "instance"}, false},
		{"neutron", []string{"Security_Group", "security_group"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.service, func(t *testing.T) {
			err := ValidateResources(tc.service, tc.resources)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateResources(%q, %v) = %v, want error = %v", tc.service, tc.resources, err, tc.wantErr)
			}
		})
	}
}
