package generators

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/astutil"
)

// UpdateDiscoveryFile adds new resource discoverers to an existing discovery file
func UpdateDiscoveryFile(baseDir, serviceName, displayName string, newResources []string) error {
	filePath := filepath.Join(baseDir, "pkg", "discovery", "services", serviceName+".go")

	if !fileExists(filePath) {
		return fmt.Errorf("discovery file %s does not exist", filePath)
	}

	fset, file, err := astutil.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing discovery file: %w", err)
	}

	existingTypes := existingTypeNames(file)

	var decls []ast.Decl
	for _, res := range newResources {
		typeName := displayName + ToPascal(res) + "Discoverer"
		if existingTypes[typeName] {
			continue
		}
		newDecls, err := buildDiscovererDecls(serviceName, displayName, res)
		if err != nil {
			return err
		}
		decls = append(decls, newDecls...)
	}

	if len(decls) == 0 {
		return nil
	}

	file.Decls = append(file.Decls, decls...)
	return astutil.WriteFile(filePath, fset, file)
}

func buildDiscovererDecls(serviceName, displayName, resource string) ([]ast.Decl, error) {
	titleRes := ToPascal(resource)
	snippet := fmt.Sprintf(`package services

// %s%sDiscoverer discovers %s resources of type %s.
// Placeholder implementation: returns no jobs. Fill in real OpenStack calls later.
//
// TODO(OSPA): Implement real discovery for %s/%s (pagination + jobs).
type %s%sDiscoverer struct{}

// ResourceType returns the resource type this discoverer handles
func (d *%s%sDiscoverer) ResourceType() string {
	return %q
}

// Discover discovers resources and sends them to the returned channel
func (d *%s%sDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan discovery.Job, error) {
	_ = ctx
	_ = client
	_ = allTenants
	// TODO(OSPA): Replace this placeholder with real discovery logic.
	ch := make(chan discovery.Job)
	close(ch)
	return ch, nil
}
`, displayName, titleRes, serviceName, resource, serviceName, resource, displayName, titleRes,
		displayName, titleRes, resource, displayName, titleRes)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", snippet, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing discoverer snippet: %w", err)
	}
	return file.Decls, nil
}

func existingTypeNames(file *ast.File) map[string]bool {
	names := make(map[string]bool)
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name != nil {
				names[ts.Name.Name] = true
			}
		}
	}
	return names
}
