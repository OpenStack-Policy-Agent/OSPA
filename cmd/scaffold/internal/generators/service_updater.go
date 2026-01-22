package generators

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/OpenStack-Policy-Agent/OSPA/cmd/scaffold/internal/astutil"
)

// UpdateServiceFile adds new resources to an existing service file
func UpdateServiceFile(baseDir, serviceName, displayName string, newResources []string) error {
	filePath := filepath.Join(baseDir, "pkg", "services", "services", serviceName+".go")

	if !fileExists(filePath) {
		return fmt.Errorf("service file %s does not exist", filePath)
	}

	fset, file, err := astutil.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing service file: %w", err)
	}

	added, err := addRegisterResourceCallsAST(file, serviceName, newResources)
	if err != nil {
		return err
	}

	addedAuditor, err := addAuditorCasesAST(file, serviceName, displayName, newResources)
	if err != nil {
		return err
	}

	addedDiscoverer, err := addDiscovererCasesAST(file, displayName, newResources)
	if err != nil {
		return err
	}

	if !added && !addedAuditor && !addedDiscoverer {
		return nil
	}

	return astutil.WriteFile(filePath, fset, file)
}
func addRegisterResourceCallsAST(file *ast.File, serviceName string, resources []string) (bool, error) {
	initFn := astutil.FindFunc(file, "init")
	if initFn == nil || initFn.Body == nil {
		return false, fmt.Errorf("init() not found in service file")
	}

	existing := make(map[string]bool)
	qualifier := ""

	for _, stmt := range initFn.Body.List {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if fun.Name != "RegisterResource" {
				continue
			}
		case *ast.SelectorExpr:
			if fun.Sel == nil || fun.Sel.Name != "RegisterResource" {
				continue
			}
			if ident, ok := fun.X.(*ast.Ident); ok {
				qualifier = ident.Name
			}
		default:
			continue
		}
		if len(call.Args) < 2 {
			continue
		}
		if lit, ok := call.Args[1].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			existing[strings.Trim(lit.Value, "\"")] = true
		}
	}

	added := false
	for _, res := range resources {
		if existing[res] {
			continue
		}
		added = true
		var fun ast.Expr
		if qualifier != "" {
			fun = &ast.SelectorExpr{X: ast.NewIdent(qualifier), Sel: ast.NewIdent("RegisterResource")}
		} else {
			fun = ast.NewIdent("RegisterResource")
		}
		call := &ast.CallExpr{
			Fun: fun,
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(serviceName)},
				&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(res)},
			},
		}
		initFn.Body.List = append(initFn.Body.List, &ast.ExprStmt{X: call})
	}

	return added, nil
}

func addAuditorCasesAST(file *ast.File, serviceName, displayName string, resources []string) (bool, error) {
	fn := astutil.FindFunc(file, "GetResourceAuditor")
	sw := astutil.FindSwitchOnIdent(fn, "resourceType")
	if sw == nil {
		return false, nil
	}

	existing := astutil.CaseValues(sw)
	var cases []*ast.CaseClause
	for _, res := range resources {
		if existing[strconv.Quote(res)] {
			continue
		}
		cases = append(cases, auditorCase(serviceName, res))
	}
	astutil.InsertCasesBeforeDefault(sw, cases)
	return len(cases) > 0, nil
}

func addDiscovererCasesAST(file *ast.File, displayName string, resources []string) (bool, error) {
	fn := astutil.FindFunc(file, "GetResourceDiscoverer")
	sw := astutil.FindSwitchOnIdent(fn, "resourceType")
	if sw == nil {
		return false, nil
	}

	existing := astutil.CaseValues(sw)
	var cases []*ast.CaseClause
	for _, res := range resources {
		if existing[strconv.Quote(res)] {
			continue
		}
		cases = append(cases, discovererCase(displayName, res))
	}
	astutil.InsertCasesBeforeDefault(sw, cases)
	return len(cases) > 0, nil
}

func auditorCase(serviceName, resource string) *ast.CaseClause {
	titleRes := ToPascal(resource)
	return &ast.CaseClause{
		List: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(resource)}},
		Body: []ast.Stmt{
			&ast.ReturnStmt{Results: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X: &ast.CompositeLit{
						Type: &ast.SelectorExpr{X: ast.NewIdent(serviceName), Sel: ast.NewIdent(titleRes + "Auditor")},
					},
				},
				ast.NewIdent("nil"),
			}},
		},
	}
}

func discovererCase(displayName, resource string) *ast.CaseClause {
	titleRes := ToPascal(resource)
	return &ast.CaseClause{
		List: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(resource)}},
		Body: []ast.Stmt{
			&ast.ReturnStmt{Results: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X: &ast.CompositeLit{
						Type: &ast.SelectorExpr{X: ast.NewIdent("discovery_services"), Sel: ast.NewIdent(displayName + titleRes + "Discoverer")},
					},
				},
				ast.NewIdent("nil"),
			}},
		},
	}
}
