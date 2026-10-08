package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `junction - scaffolds domains and the entries that reach them

usage:
  junction api <name> [--table=name]      HTTP entry, mounts it in cmd/api/modules.go
  junction worker <name> [--table=name]   queue entry, mounts it in cmd/worker/modules.go
  junction event <name> [--table=name]    event entry, mounts it in cmd/event/modules.go
  junction migration <name>               empty migration pair in migrations/
  junction table <name>                   create table migration pair, without a domain
  junction status                         show domains and their entries

Each command creates the shared core (domain, repository, service) when it is
missing and reuses it when it is already there, so entries can be added in any
order and any combination.
`

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "junction:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(usage)
		return nil
	}

	root, err := findRoot()
	if err != nil {
		return err
	}

	if args[0] == "status" {
		return status(root)
	}

	module, err := readModule(root)
	if err != nil {
		return err
	}

	if args[0] == "migration" {
		if len(args) != 2 {
			return fmt.Errorf("usage: junction migration <name>")
		}

		changes, err := manualMigration(root, module, args[1])
		if err != nil {
			return err
		}

		report(changes)
		return nil
	}

	if args[0] == "table" {
		if len(args) != 2 {
			return fmt.Errorf("usage: junction table <name>")
		}

		changes, err := tableMigration(root, module, args[1])
		if err != nil {
			return err
		}

		report(changes)
		return nil
	}

	if _, ok := entrySpecs[args[0]]; !ok {
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}

	return add(root, module, args[0], args[1:])
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: junction %s <name> [--table=name]\n", name)
		flags.PrintDefaults()
	}

	return flags
}
