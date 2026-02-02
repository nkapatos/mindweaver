package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nkapatos/mindweaver/internal/imex"
)

func main() {
	// Global flags
	var (
		dryRun = flag.Bool("dry-run", false, "Preview actions, do not modify files")
		config = flag.String("config", "", "Path to config file")
	)
	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  imex [--dry-run] [--config <file>] import [import-flags]")
		fmt.Println("  imex [--dry-run] [--config <file>] export [export-flags]")
		fmt.Println("Subcommands:")
		fmt.Println("  import    Import files into Mindweaver")
		fmt.Println("  export    Export collections from Mindweaver")
		fmt.Println("Global flags:")
		fmt.Println("  --dry-run")
		fmt.Println("  --config <file>")
	}
	flag.Parse()

	// Find subcommand in os.Args after flags
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: missing subcommand ('import' or 'export')")
		flag.Usage()
		os.Exit(1)
	}
	cmd := args[0]

	// Route to subcommands
	os.Args = os.Args[:1] // workaround to avoid flag conflicts downstream
	switch cmd {
	case "import":
		imex.RunImport(args[1:], *dryRun, *config)
	case "export":
		imex.RunExport(args[1:], *dryRun, *config)
	default:
		fmt.Printf("Unknown subcommand: %s\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}
