package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func wire(root string, spec bridgeSpec, names Names) (change, error) {
	display := spec.Target
	full := filepath.Join(root, filepath.FromSlash(spec.Target))

	source, err := os.ReadFile(full)
	if err != nil {
		return change{}, fmt.Errorf("reading %s: %w", display, err)
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, full, source, parser.ParseComments)
	if err != nil {
		return change{}, fmt.Errorf("parsing %s: %w", display, err)
	}

	fn := findFunc(file, spec.Fn)
	if fn == nil {
		return change{}, fmt.Errorf("function %s not found in %s", spec.Fn, display)
	}

	alias := names.Lower + spec.Layer

	if hasCall(fn, alias, spec.Register) {
		return change{Action: "wired", Path: display + " (already)"}, nil
	}

	returnStmt := trailingReturn(fn)
	if returnStmt == nil {
		return change{}, fmt.Errorf("%s does not end with a return statement", spec.Fn)
	}

	imports := importDecl(file)
	if imports == nil || imports.Rparen == token.NoPos {
		return change{}, fmt.Errorf("%s has no parenthesized import block", display)
	}

	snippet, err := render(spec.Wiring, names)
	if err != nil {
		return change{}, err
	}

	importOffset := fset.Position(imports.Rparen).Offset
	returnOffset := fset.Position(returnStmt.Pos()).Offset

	var builder strings.Builder

	builder.Write(source[:importOffset])
	builder.WriteString(importBlock(spec, names))
	builder.Write(source[importOffset:returnOffset])
	builder.WriteString(strings.TrimSpace(string(snippet)))
	builder.WriteString("\n\n\t")
	builder.Write(source[returnOffset:])

	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return change{}, fmt.Errorf("formatting %s: %w", display, err)
	}

	err = os.WriteFile(full, formatted, 0o644)
	if err != nil {
		return change{}, err
	}

	return change{Action: "wired", Path: display}, nil
}

func importBlock(spec bridgeSpec, names Names) string {
	packages := []string{spec.Layer, "service", "repository"}

	var builder strings.Builder

	for _, pkg := range packages {
		path := fmt.Sprintf("%s/internal/modules/%s/%s", names.Module, names.Lower, pkg)
		builder.WriteString("\t" + names.Lower + pkg + " " + strconv.Quote(path) + "\n")
	}

	return builder.String()
}

func findFunc(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Recv == nil && fn.Name.Name == name {
			return fn
		}
	}

	return nil
}

func importDecl(file *ast.File) *ast.GenDecl {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if ok && genDecl.Tok == token.IMPORT {
			return genDecl
		}
	}

	return nil
}

func trailingReturn(fn *ast.FuncDecl) *ast.ReturnStmt {
	if fn.Body == nil || len(fn.Body.List) == 0 {
		return nil
	}

	returnStmt, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt)
	if !ok {
		return nil
	}

	return returnStmt
}

func hasCall(fn *ast.FuncDecl, alias, register string) bool {
	found := false

	ast.Inspect(fn, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == alias && selector.Sel.Name == register {
			found = true
			return false
		}

		return true
	})

	return found
}
