package generators

import (
	"path/filepath"
	"text/template"
)

// GenerateE2ETest generates the e2e test file.
func GenerateE2ETest(baseDir, serviceName, displayName string, resources []string) error {
	specs, err := buildResourceSpecs(serviceName, resources)
	if err != nil {
		return err
	}
	return generateE2ETestWithSpecs(baseDir, serviceName, displayName, specs)
}

func generateE2ETestWithSpecs(baseDir, serviceName, displayName string, resources []ResourceSpec) error {
	e2eFile := filepath.Join(baseDir, "e2e", serviceName+"_test.go")

	tmpl := `//go:build e2e

package e2e

import (
	"testing"
)

{{range .Resources}}
func Test{{$.DisplayName}}_{{.Name | Pascal}}Audit(t *testing.T) {
	// This e2e test requires a real OpenStack cloud:
	//   OS_CLIENT_CONFIG_FILE pointing to clouds.yaml
	//   OS_CLOUD set to a valid cloud entry
	engine := NewTestEngine(t)

	policyYAML := ` + "`" + `version: v1
defaults:
  workers: 2
policies:
  - {{$.ServiceName}}:
    - name: test-{{.Name}}-check
      description: Test {{.Name}} check
      service: {{$.ServiceName}}
      resource: {{.Name}}
      check:
        status: active
      action: log` + "`" + `

	policy := engine.LoadPolicyFromYAML(t, policyYAML)
	results := engine.RunAudit(t, policy)

	{{.Name | Pascal}}Results := results.FilterByService("{{$.ServiceName}}").FilterByResourceType("{{.Name}}")
	{{.Name | Pascal}}Results.LogSummary(t)

	if {{.Name | Pascal}}Results.Errors > 0 {
		t.Logf("Warning: %d errors encountered during {{.Name}} audit", {{.Name | Pascal}}Results.Errors)
	}
}

{{end}}
`

	data := struct {
		ServiceName string
		DisplayName string
		Resources   []ResourceSpec
	}{
		ServiceName: serviceName,
		DisplayName: displayName,
		Resources:   resources,
	}

	funcMap := template.FuncMap{
		"Pascal":     ToPascal,
		"JoinOrNone": JoinOrNone,
	}

	t, err := template.New("e2etest").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	return writeFile(e2eFile, t, data)
}
