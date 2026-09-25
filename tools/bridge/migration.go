package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const migrationsDir = "migrations"

func newVersion() string {
	return time.Now().UTC().Format("20060102150405")
}

type migrationData struct {
	Names
	Title string
}

func domainMigration(root string, names Names) ([]change, error) {
	title := "create_" + names.Snake + "_table"

	found, err := existingMigration(root, title)
	if err != nil {
		return nil, err
	}

	if found != "" {
		return []change{{Action: "exists", Path: found}}, nil
	}

	return writeMigration(root, migrationData{Names: names, Title: title},
		"migration_up.sql.tmpl", "migration_down.sql.tmpl")
}

// manualMigration writes an empty pair for a change the generator cannot guess.
func manualMigration(root, module, input string) ([]change, error) {
	words := splitWords(input)
	if len(words) == 0 {
		return nil, fmt.Errorf("invalid migration name %q", input)
	}

	if !validName.MatchString(input) {
		return nil, fmt.Errorf("invalid migration name %q: must start with a letter and contain only letters, digits, - or _", input)
	}

	title := joinWords(words, "_", false)

	return writeMigration(root, migrationData{Names: Names{Module: module}, Title: title},
		"migration_manual_up.sql.tmpl", "migration_manual_down.sql.tmpl")
}

func existingMigration(root, title string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(root, migrationsDir, "*_"+title+".up.sql"))
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", nil
	}

	return filepath.ToSlash(filepath.Join(migrationsDir, filepath.Base(matches[0]))), nil
}

func writeMigration(root string, data migrationData, upTemplate, downTemplate string) ([]change, error) {
	version := newVersion()

	steps := []struct {
		direction string
		template  string
	}{
		{direction: "up", template: upTemplate},
		{direction: "down", template: downTemplate},
	}

	err := os.MkdirAll(filepath.Join(root, migrationsDir), 0o755)
	if err != nil {
		return nil, err
	}

	var changes []change

	for _, step := range steps {
		name := fmt.Sprintf("%s_%s.%s.sql", version, data.Title, step.direction)
		display := filepath.ToSlash(filepath.Join(migrationsDir, name))

		content, err := renderMigration(step.template, data)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(filepath.Join(root, migrationsDir, name), content, 0o644)
		if err != nil {
			return nil, err
		}

		changes = append(changes, change{Action: "created", Path: display})
	}

	return changes, nil
}
