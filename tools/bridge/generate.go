package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates
var templatesFS embed.FS

var templates = template.Must(template.ParseFS(templatesFS, "templates/*"))

type change struct {
	Action string
	Path   string
}

func render(name string, names Names) ([]byte, error) {
	var buf bytes.Buffer

	err := templates.ExecuteTemplate(&buf, name, names)
	if err != nil {
		return nil, fmt.Errorf("rendering %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

func renderPath(pattern string, names Names) (string, error) {
	tmpl, err := template.New("path").Parse(pattern)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, names)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func writeFiles(root string, specs []fileSpec, names Names) ([]change, error) {
	var changes []change

	for _, spec := range specs {
		relative, err := renderPath(spec.Path, names)
		if err != nil {
			return nil, err
		}

		full := filepath.Join(root, "internal", "modules", names.Lower, relative)
		display := filepath.ToSlash(filepath.Join("internal/modules", names.Lower, relative))

		if _, err := os.Stat(full); err == nil {
			changes = append(changes, change{Action: "exists", Path: display})
			continue
		}

		content, err := render(spec.Template, names)
		if err != nil {
			return nil, err
		}

		formatted, err := format.Source(content)
		if err != nil {
			return nil, fmt.Errorf("formatting %s: %w", display, err)
		}

		err = os.MkdirAll(filepath.Dir(full), 0o755)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(full, formatted, 0o644)
		if err != nil {
			return nil, err
		}

		changes = append(changes, change{Action: "created", Path: display})
	}

	return changes, nil
}

func add(root, module, bridgeName string, args []string) error {
	spec, ok := bridges[bridgeName]
	if !ok {
		return fmt.Errorf("bridge %q is not implemented yet", bridgeName)
	}

	flags := newFlagSet(bridgeName)
	table := flags.String("table", "", "database table name (default: domain name in snake_case, pluralized)")

	err := flags.Parse(reorderFlags(args))
	if err != nil {
		return err
	}

	if flags.NArg() != 1 {
		return fmt.Errorf("usage: bridge %s <name> [--table=name]", bridgeName)
	}

	names, err := newNames(flags.Arg(0), *table, module)
	if err != nil {
		return err
	}

	if pascal, found := existingPascal(root, names.Lower); found && pascal != names.Pascal {
		names, err = newNames(pascal, *table, module)
		if err != nil {
			return err
		}
	}

	coreChanges, err := writeFiles(root, coreFiles, names)
	if err != nil {
		return err
	}

	layerChanges, err := writeFiles(root, spec.Files, names)
	if err != nil {
		return err
	}

	wiringChange, err := wire(root, spec, names)
	if err != nil {
		return err
	}

	report(append(append(coreChanges, layerChanges...), wiringChange))
	return nil
}

func reorderFlags(args []string) []string {
	var flagArgs []string
	var positional []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			continue
		}

		positional = append(positional, arg)
	}

	return append(flagArgs, positional...)
}

func report(changes []change) {
	width := 0

	for _, item := range changes {
		if len(item.Action) > width {
			width = len(item.Action)
		}
	}

	for _, item := range changes {
		fmt.Printf("%-*s  %s\n", width, item.Action, item.Path)
	}
}

func status(root string) error {
	modules := filepath.Join(root, "internal", "modules")

	entries, err := os.ReadDir(modules)
	if err != nil {
		return err
	}

	var rows [][]string
	width := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		row := []string{entry.Name()}

		for _, bridgeName := range bridgeOrder {
			layer := filepath.Join(modules, entry.Name(), bridges[bridgeName].Layer)
			mark := "."

			if info, err := os.Stat(layer); err == nil && info.IsDir() {
				mark = "x"
			}

			row = append(row, fmt.Sprintf("%s %s", bridgeName, mark))
		}

		if len(entry.Name()) > width {
			width = len(entry.Name())
		}

		rows = append(rows, row)
	}

	for _, row := range rows {
		fmt.Printf("%-*s   %s\n", width, row[0], strings.Join(row[1:], "   "))
	}

	return nil
}
