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

const modulesVar = "modules"

const moduleSymbol = "Module"

func wire(root string, spec entrySpec, names Names) (change, error) {
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

	list := findModulesList(file)
	if list == nil {
		return change{}, fmt.Errorf("var %s slice literal not found in %s", modulesVar, display)
	}

	alias := names.Lower + spec.Layer

	if hasElement(list, alias, moduleSymbol) {
		return change{Action: "wired", Path: display + " (already)"}, nil
	}

	imports := importDecl(file)
	if imports == nil || imports.Rparen == token.NoPos {
		return change{}, fmt.Errorf("%s has no parenthesized import block", display)
	}

	importOffset := fset.Position(imports.Rparen).Offset
	elementOffset := fset.Position(list.Rbrace).Offset

	element := "\t" + alias + "." + moduleSymbol + ",\n"
	if fset.Position(list.Lbrace).Line == fset.Position(list.Rbrace).Line {
		element = "\n" + element
	}

	var builder strings.Builder

	builder.Write(source[:importOffset])
	builder.WriteString(importLine(spec, names))
	builder.Write(source[importOffset:elementOffset])
	builder.WriteString(element)
	builder.Write(source[elementOffset:])

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

func importLine(spec entrySpec, names Names) string {
	path := fmt.Sprintf("%s/internal/modules/%s/%s", names.Module, names.Lower, spec.Layer)

	return "\t" + names.Lower + spec.Layer + " " + strconv.Quote(path) + "\n"
}

func findModulesList(file *ast.File) *ast.CompositeLit {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				if name.Name != modulesVar || i >= len(valueSpec.Values) {
					continue
				}

				list, ok := valueSpec.Values[i].(*ast.CompositeLit)
				if ok {
					return list
				}
			}
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

func hasElement(list *ast.CompositeLit, alias, symbol string) bool {
	for _, element := range list.Elts {
		selector, ok := element.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		ident, ok := selector.X.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name == alias && selector.Sel.Name == symbol {
			return true
		}
	}

	return false
}
