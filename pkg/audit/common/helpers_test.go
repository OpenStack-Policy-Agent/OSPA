package common

import (
	"testing"
	"time"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/audit"
	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

type fakeResource struct {
	id        string
	name      string
	projectID string
	status    string
	createdAt time.Time
	updatedAt time.Time
}

func (f fakeResource) GetID() string           { return f.id }
func (f fakeResource) GetName() string         { return f.name }
func (f fakeResource) GetProjectID() string    { return f.projectID }
func (f fakeResource) GetStatus() string       { return f.status }
func (f fakeResource) GetCreatedAt() time.Time { return f.createdAt }
func (f fakeResource) GetUpdatedAt() time.Time { return f.updatedAt }

func TestBuildBaseResult(t *testing.T) {
	r := fakeResource{id: "id-1", name: "my-res", projectID: "proj-1", status: "ACTIVE"}
	rule := &policy.Rule{Name: "test-rule"}
	result := BuildBaseResult(r, rule)

	if result.ResourceID != "id-1" {
		t.Errorf("ResourceID = %q, want %q", result.ResourceID, "id-1")
	}
	if result.ResourceName != "my-res" {
		t.Errorf("ResourceName = %q, want %q", result.ResourceName, "my-res")
	}
	if result.Status != "ACTIVE" {
		t.Errorf("Status = %q, want %q", result.Status, "ACTIVE")
	}
	if !result.Compliant {
		t.Error("expected Compliant=true by default")
	}
}

func TestCheckExemptByName_Exact(t *testing.T) {
	r := fakeResource{name: "default"}
	rule := &policy.Rule{Check: policy.CheckConditions{ExemptNames: []string{"default"}}}
	result := &audit.Result{}

	if !CheckExemptByName(r, rule, result) {
		t.Error("expected exempt=true for exact name match")
	}
	if result.Observation != "exempt by name" {
		t.Errorf("observation = %q, want %q", result.Observation, "exempt by name")
	}
}

func TestCheckExemptByName_Glob(t *testing.T) {
	r := fakeResource{name: "ospa-e2e-network-42"}
	rule := &policy.Rule{Check: policy.CheckConditions{ExemptNames: []string{"ospa-e2e-*"}}}
	result := &audit.Result{}

	if !CheckExemptByName(r, rule, result) {
		t.Error("expected exempt=true for glob pattern match")
	}
}

func TestCheckExemptByName_NoMatch(t *testing.T) {
	r := fakeResource{name: "production-net"}
	rule := &policy.Rule{Check: policy.CheckConditions{ExemptNames: []string{"test-*"}}}
	result := &audit.Result{}

	if CheckExemptByName(r, rule, result) {
		t.Error("expected exempt=false for non-matching name")
	}
}

func TestCheckExemptByName_Empty(t *testing.T) {
	r := fakeResource{name: "anything"}
	rule := &policy.Rule{}
	result := &audit.Result{}

	if CheckExemptByName(r, rule, result) {
		t.Error("expected exempt=false when no exempt_names specified")
	}
}

func TestCheckStatus_Match(t *testing.T) {
	r := fakeResource{status: "ERROR"}
	rule := &policy.Rule{Check: policy.CheckConditions{Status: "ERROR"}}
	result := &audit.Result{Compliant: true}

	CheckStatus(r, rule, result)

	if result.Compliant {
		t.Error("expected non-compliant when status matches")
	}
}

func TestCheckStatus_NoMatch(t *testing.T) {
	r := fakeResource{status: "ACTIVE"}
	rule := &policy.Rule{Check: policy.CheckConditions{Status: "ERROR"}}
	result := &audit.Result{Compliant: true}

	CheckStatus(r, rule, result)

	if !result.Compliant {
		t.Error("expected compliant when status does not match")
	}
}

func TestCheckStatus_Empty(t *testing.T) {
	r := fakeResource{status: "ACTIVE"}
	rule := &policy.Rule{}
	result := &audit.Result{Compliant: true}

	CheckStatus(r, rule, result)

	if !result.Compliant {
		t.Error("expected compliant when no status check specified")
	}
}

func TestCheckAgeGT_OldResource(t *testing.T) {
	old := time.Now().Add(-48 * time.Hour)
	r := fakeResource{updatedAt: old}
	rule := &policy.Rule{Check: policy.CheckConditions{AgeGT: "1d"}}
	result := &audit.Result{Compliant: true}

	if err := CheckAgeGT(r, rule, result); err != nil {
		t.Fatalf("CheckAgeGT() error = %v", err)
	}
	if result.Compliant {
		t.Error("expected non-compliant for resource older than 1 day")
	}
}

func TestCheckAgeGT_YoungResource(t *testing.T) {
	recent := time.Now().Add(-1 * time.Hour)
	r := fakeResource{updatedAt: recent}
	rule := &policy.Rule{Check: policy.CheckConditions{AgeGT: "1d"}}
	result := &audit.Result{Compliant: true}

	if err := CheckAgeGT(r, rule, result); err != nil {
		t.Fatalf("CheckAgeGT() error = %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant for resource younger than 1 day")
	}
}

func TestCheckAgeGT_FallbackToCreatedAt(t *testing.T) {
	old := time.Now().Add(-48 * time.Hour)
	r := fakeResource{createdAt: old}
	rule := &policy.Rule{Check: policy.CheckConditions{AgeGT: "1d"}}
	result := &audit.Result{Compliant: true}

	if err := CheckAgeGT(r, rule, result); err != nil {
		t.Fatalf("CheckAgeGT() error = %v", err)
	}
	if result.Compliant {
		t.Error("expected non-compliant when falling back to createdAt")
	}
}

func TestCheckAgeGT_ZeroTimestamp(t *testing.T) {
	r := fakeResource{}
	rule := &policy.Rule{Check: policy.CheckConditions{AgeGT: "1d"}}
	result := &audit.Result{Compliant: true}

	if err := CheckAgeGT(r, rule, result); err != nil {
		t.Fatalf("CheckAgeGT() error = %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant when timestamps are zero")
	}
}

func TestCheckAgeGT_NotRequested(t *testing.T) {
	r := fakeResource{}
	rule := &policy.Rule{}
	result := &audit.Result{Compliant: true}

	if err := CheckAgeGT(r, rule, result); err != nil {
		t.Fatalf("CheckAgeGT() error = %v", err)
	}
	if !result.Compliant {
		t.Error("expected compliant when age_gt not specified")
	}
}

func TestCheckAgeGT_InvalidFormat(t *testing.T) {
	r := fakeResource{updatedAt: time.Now()}
	rule := &policy.Rule{Check: policy.CheckConditions{AgeGT: "abc"}}
	result := &audit.Result{Compliant: true}

	err := CheckAgeGT(r, rule, result)
	if err == nil {
		t.Error("expected error for invalid age_gt format")
	}
}

func TestRunCommonChecks_ExemptShortCircuits(t *testing.T) {
	r := fakeResource{name: "excluded", status: "ERROR"}
	rule := &policy.Rule{
		Check: policy.CheckConditions{
			Status:      "ERROR",
			ExemptNames: []string{"excluded"},
		},
	}
	result := &audit.Result{Compliant: true}

	exempt, err := RunCommonChecks(r, rule, result)
	if err != nil {
		t.Fatalf("RunCommonChecks() error = %v", err)
	}
	if !exempt {
		t.Error("expected exempt=true")
	}
	if !result.Compliant {
		t.Error("expected compliant for exempt resource")
	}
}

func TestRunCommonChecks_StatusAndAge(t *testing.T) {
	old := time.Now().Add(-72 * time.Hour)
	r := fakeResource{name: "prod", status: "SHUTOFF", updatedAt: old}
	rule := &policy.Rule{
		Check: policy.CheckConditions{
			Status: "SHUTOFF",
			AgeGT:  "1d",
		},
	}
	result := &audit.Result{Compliant: true}

	exempt, err := RunCommonChecks(r, rule, result)
	if err != nil {
		t.Fatalf("RunCommonChecks() error = %v", err)
	}
	if exempt {
		t.Error("expected exempt=false")
	}
	if result.Compliant {
		t.Error("expected non-compliant when both status and age_gt trigger")
	}
}
