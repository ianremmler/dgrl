package main

import (
	"github.com/ianremmler/dgrl"

	"flag"
	"fmt"
	"os"
)

func main() {
	var doJSON bool
	flag.BoolVar(&doJSON, "j", false, "Export to JSON")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "dgrl reads from stdin only.")
		os.Exit(1)
	}
	parser := dgrl.NewParser()
	tree := parser.Parse(os.Stdin)
	switch {
	case doJSON:
		tree.WriteJSON(os.Stdout)
	default:
		tree.Write(os.Stdout)
	}
}
