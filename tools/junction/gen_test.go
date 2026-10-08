package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGeneratedCodeCompiles(t *testing.T) {
	root, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}

	module, err := readModule(root)
	if err != nil {
		t.Fatal(err)
	}

	const name = "junctiongentest"

	generated := filepath.Join(root, "internal", "modules", name)

	if _, err := os.Stat(generated); err == nil {
		t.Fatalf("%s already exists, refusing to overwrite it", generated)
	}

	t.Cleanup(func() {
		os.RemoveAll(generated)
	})

	names, err := newNames(name, "", module)
	if err != nil {
		t.Fatal(err)
	}

	specs := append([]fileSpec{}, coreFiles...)

	for _, entryName := range entryOrder {
		specs = append(specs, entrySpecs[entryName].Files...)
	}

	_, err = writeFiles(root, specs, names)
	if err != nil {
		t.Fatal(err)
	}

	// go test, not go build: the generated _test.go files must compile and pass.
	cmd := exec.Command("go", "test", "-count=1", "./internal/modules/"+name+"/...")
	cmd.Dir = root

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("the generated code does not compile or its tests fail: %v\n%s", err, output)
	}
}

func TestWireIsIdempotent(t *testing.T) {
	root, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}

	module, err := readModule(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, entryName := range entryOrder {
		t.Run(entryName, func(t *testing.T) {
			wireOnce(t, root, module, entrySpecs[entryName])
		})
	}
}

func wireOnce(t *testing.T, root, module string, spec entrySpec) {
	t.Helper()

	temp := t.TempDir()
	target := filepath.Join(temp, filepath.FromSlash(spec.Target))

	err := os.MkdirAll(filepath.Dir(target), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(spec.Target)))
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(target, source, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	names, err := newNames("junctionwiretest", "", module)
	if err != nil {
		t.Fatal(err)
	}

	first, err := wire(temp, spec, names)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(first.Path, "already") {
		t.Fatalf("first wire reported the call as already present")
	}

	wired, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	_, err = parser.ParseFile(token.NewFileSet(), target, wired, parser.ParseComments)
	if err != nil {
		t.Fatalf("wired file does not parse: %v", err)
	}

	element := names.Lower + spec.Layer + "." + moduleSymbol + ","
	if !bytes.Contains(wired, []byte(element)) {
		t.Fatalf("wired file is missing %s", element)
	}

	if !bytes.Contains(wired, []byte(names.Module+"/internal/modules/"+names.Lower+"/"+spec.Layer)) {
		t.Fatalf("wired file is missing the %s import", spec.Layer)
	}

	second, err := wire(temp, spec, names)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(second.Path, "already") {
		t.Fatalf("second wire did not detect the existing element")
	}

	rewired, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(wired, rewired) {
		t.Fatalf("second wire modified the file")
	}
}

func TestNames(t *testing.T) {
	cases := []struct {
		input string
		want  Names
	}{
		{"product", Names{Pascal: "Product", Camel: "product", Lower: "product", Snake: "product", Kebab: "product", Table: "product"}},
		{"order-item", Names{Pascal: "OrderItem", Camel: "orderItem", Lower: "orderitem", Snake: "order_item", Kebab: "order-item", Table: "order_item"}},
		{"order_item", Names{Pascal: "OrderItem", Camel: "orderItem", Lower: "orderitem", Snake: "order_item", Kebab: "order-item", Table: "order_item"}},
		{"OrderItem", Names{Pascal: "OrderItem", Camel: "orderItem", Lower: "orderitem", Snake: "order_item", Kebab: "order-item", Table: "order_item"}},
	}

	for _, testCase := range cases {
		got, err := newNames(testCase.input, "", "example.com/app")
		if err != nil {
			t.Fatalf("%s: %v", testCase.input, err)
		}

		testCase.want.Module = "example.com/app"

		if got != testCase.want {
			t.Errorf("%s:\n got %+v\nwant %+v", testCase.input, got, testCase.want)
		}
	}

	for _, invalid := range []string{"", "1product", "pro duct", "pro/duct"} {
		_, err := newNames(invalid, "", "example.com/app")
		if err == nil {
			t.Errorf("expected %q to be rejected", invalid)
		}
	}
}

func TestReorderFlags(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{[]string{"product"}, []string{"product"}},
		{[]string{"product", "--table=tbl"}, []string{"--table=tbl", "product"}},
		{[]string{"--table=tbl", "product"}, []string{"--table=tbl", "product"}},
	}

	for _, testCase := range cases {
		got := reorderFlags(testCase.args)

		if strings.Join(got, " ") != strings.Join(testCase.want, " ") {
			t.Errorf("%v:\n got %v\nwant %v", testCase.args, got, testCase.want)
		}
	}
}

func TestTableDefaultsToTheGivenName(t *testing.T) {
	cases := map[string]string{
		"product":    "product",
		"order-item": "order_item",
		"OrderItem":  "order_item",
	}

	for input, want := range cases {
		names, err := newNames(input, "", "example.com/app")
		if err != nil {
			t.Fatalf("%s: %v", input, err)
		}

		if names.Table != want {
			t.Errorf("%s: got table %q, want %q", input, names.Table, want)
		}
	}

	names, err := newNames("category", "tbl_categories", "example.com/app")
	if err != nil {
		t.Fatal(err)
	}

	if names.Table != "tbl_categories" {
		t.Errorf("explicit table was ignored: got %q", names.Table)
	}
}

func TestDomainMigrationIsVersionedAndWrittenOnce(t *testing.T) {
	root, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}

	module, err := readModule(root)
	if err != nil {
		t.Fatal(err)
	}

	names, err := newNames("junctionmigrationtest", "", module)
	if err != nil {
		t.Fatal(err)
	}

	title := "create_" + names.Snake + "_table"

	found, err := existingMigration(root, title)
	if err != nil {
		t.Fatal(err)
	}

	if found != "" {
		t.Fatalf("%s already exists, refusing to overwrite it", found)
	}

	t.Cleanup(func() {
		matches, _ := filepath.Glob(filepath.Join(root, migrationsDir, "*_"+title+".*.sql"))
		for _, match := range matches {
			os.Remove(match)
		}
	})

	first, err := domainMigration(root, names)
	if err != nil {
		t.Fatal(err)
	}

	if len(first) != 2 {
		t.Fatalf("got %d files, want the up and down pair", len(first))
	}

	versioned := regexp.MustCompile(`^migrations/\d{14}_` + title + `\.(up|down)\.sql$`)

	for _, item := range first {
		if item.Action != "created" {
			t.Errorf("got action %q for %s, want created", item.Action, item.Path)
		}

		if !versioned.MatchString(item.Path) {
			t.Errorf("%s is not a 14 digit versioned migration", item.Path)
		}
	}

	upPath := filepath.Join(root, filepath.FromSlash(first[0].Path))

	content, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatal(err)
	}

	// The DDL must carry the columns the generated repository queries.
	for _, column := range []string{names.Table, "`id`", "`name`", "`created_at`", "`updated_at`", "`deleted_at`"} {
		if !bytes.Contains(content, []byte(column)) {
			t.Errorf("the migration is missing %s", column)
		}
	}

	second, err := domainMigration(root, names)
	if err != nil {
		t.Fatal(err)
	}

	if len(second) != 1 || second[0].Action != "exists" {
		t.Fatalf("a second run did not reuse the migration: %+v", second)
	}

	matches, err := filepath.Glob(filepath.Join(root, migrationsDir, "*_"+title+".up.sql"))
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 1 {
		t.Fatalf("got %d up migrations, want exactly 1", len(matches))
	}
}
