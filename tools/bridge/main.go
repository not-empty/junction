package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `bridge - scaffolds domains and their delivery bridges

usage:
  bridge api <name> [--table=name]      HTTP bridge, wires cmd/api/router.go
  bridge worker <name> [--table=name]   queue bridge, wires cmd/worker/queues.go
  bridge event <name> [--table=name]    event bridge, wires cmd/event/events.go
  bridge status                         show domains and their bridges

Each command creates the shared core (domain, repository, service) when it is
missing and reuses it when it is already there, so bridges can be added in any
order and any combination.
`

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "bridge:", err)
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

	if _, ok := bridges[args[0]]; !ok {
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}

	return add(root, module, args[0], args[1:])
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: bridge %s <name> [--table=name]\n", name)
		flags.PrintDefaults()
	}

	return flags
}
