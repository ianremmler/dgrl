package main

import (
	"github.com/ianremmler/dgrl"

	"flag"
	"fmt"
	"os"
)

func main() {
	doJson := flag.Bool("j", false, "Export to JSON")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "dgrl reads from stdin only.")
		os.Exit(1)
	}
	parser := dgrl.NewParser()
	tree := parser.Parse(os.Stdin)
	switch {
	case *doJson:
		fmt.Println(tree.ToJson())
	default:
		fmt.Print(tree)
	}
}
