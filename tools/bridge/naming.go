package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type Names struct {
	Pascal string
	Camel  string
	Lower  string
	Snake  string
	Kebab  string
	Table  string
	Module string
}

var validName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

func newNames(input, table, module string) (Names, error) {
	if !validName.MatchString(input) {
		return Names{}, fmt.Errorf("invalid domain name %q: must start with a letter and contain only letters, digits, - or _", input)
	}

	words := splitWords(input)
	if len(words) == 0 {
		return Names{}, fmt.Errorf("invalid domain name %q", input)
	}

	names := Names{
		Pascal: joinWords(words, "", true),
		Camel:  lowerFirst(joinWords(words, "", true)),
		Lower:  joinWords(words, "", false),
		Snake:  joinWords(words, "_", false),
		Kebab:  joinWords(words, "-", false),
		Module: module,
	}

	names.Table = table
	if names.Table == "" {
		names.Table = names.Snake + "s"
	}

	return names, nil
}

func splitWords(input string) []string {
	var words []string
	var current []rune

	flush := func() {
		if len(current) > 0 {
			words = append(words, string(current))
			current = nil
		}
	}

	for _, r := range input {
		switch {
		case r == '-' || r == '_':
			flush()
		case unicode.IsUpper(r):
			flush()
			current = append(current, unicode.ToLower(r))
		default:
			current = append(current, r)
		}
	}

	flush()
	return words
}

func joinWords(words []string, sep string, title bool) string {
	parts := make([]string, len(words))

	for i, word := range words {
		if title {
			parts[i] = upperFirst(word)
			continue
		}
		parts[i] = word
	}

	return strings.Join(parts, sep)
}

func upperFirst(word string) string {
	if word == "" {
		return word
	}

	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func lowerFirst(word string) string {
	if word == "" {
		return word
	}

	runes := []rune(word)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found: run this from inside the project")
		}

		dir = parent
	}
}

func readModule(root string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if after, found := strings.CutPrefix(line, "module "); found {
			return strings.TrimSpace(after), nil
		}
	}

	return "", fmt.Errorf("module directive not found in go.mod")
}

func existingPascal(root, lower string) (string, bool) {
	dir := filepath.Join(root, "internal", "modules", lower, "service")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	fset := token.NewFileSet()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			continue
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if pascal, found := strings.CutSuffix(typeSpec.Name.Name, "Service"); found && pascal != "" {
					return pascal, true
				}
			}
		}
	}

	return "", false
}
