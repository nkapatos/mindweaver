package imex

import (
	"flag"
	"fmt"
)

// RunExport handles export subcommand logic and flag parsing
func RunExport(args []string, dryRun bool, config string) {
	exportFlags := flag.NewFlagSet("export", flag.ExitOnError)
	var (
		dest            = exportFlags.String("dest", "", "Destination path for exported files (required)")
		collections     = exportFlags.String("collections", "", "Comma-separated list of collections to export")
		listCollections = exportFlags.Bool("list-collections", false, "List available collections on server")
		ping            = exportFlags.Bool("ping", false, "Ping server before export")
	)
	if err := exportFlags.Parse(args); err != nil {
		fmt.Println("Failed to parse export flags:", err)
		return
	}
	// Validate required flag
	if *dest == "" {
		fmt.Println("--dest is required for export")
		exportFlags.Usage()
		return
	}
	fmt.Printf("[export] dest=%s, collections=%s, listCollections=%v, ping=%v, dryRun=%v, config=%s\n",
		*dest, *collections, *listCollections, *ping, dryRun, config,
	)
	// TODO: Add core export logic here
}
