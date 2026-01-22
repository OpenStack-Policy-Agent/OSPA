package astutil

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

func ParseFile(path string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, file, nil
}

func WriteFile(path string, fset *token.FileSet, file *ast.File) error {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func FindFunc(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name != nil && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func FindSwitchOnIdent(fn *ast.FuncDecl, ident string) *ast.SwitchStmt {
	if fn == nil || fn.Body == nil {
		return nil
	}
	for _, stmt := range fn.Body.List {
		sw, ok := stmt.(*ast.SwitchStmt)
		if !ok {
			continue
		}
		if id, ok := sw.Tag.(*ast.Ident); ok && id.Name == ident {
			return sw
		}
	}
	return nil
}

func CaseValues(sw *ast.SwitchStmt) map[string]bool {
	values := make(map[string]bool)
	if sw == nil {
		return values
	}
	for _, stmt := range sw.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		for _, expr := range cc.List {
			if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				values[lit.Value] = true
			}
		}
	}
	return values
}

func InsertCasesBeforeDefault(sw *ast.SwitchStmt, cases []*ast.CaseClause) {
	if sw == nil || len(cases) == 0 {
		return
	}
	insertAt := len(sw.Body.List)
	for i, stmt := range sw.Body.List {
		if cc, ok := stmt.(*ast.CaseClause); ok && cc.List == nil {
			insertAt = i
			break
		}
	}
	newList := make([]ast.Stmt, 0, len(sw.Body.List)+len(cases))
	newList = append(newList, sw.Body.List[:insertAt]...)
	for _, cc := range cases {
		newList = append(newList, cc)
	}
	newList = append(newList, sw.Body.List[insertAt:]...)
	sw.Body.List = newList
}
