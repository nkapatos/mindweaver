package imex

import (
	"flag"
	"fmt"
)

// RunImport handles import subcommand logic and flag parsing
func RunImport(args []string, dryRun bool, config string) {
	importFlags := flag.NewFlagSet("import", flag.ExitOnError)
	var (
		src              = importFlags.String("src", "", "Source path or URI for import (required)")
		collection       = importFlags.String("collection", "", "Collection name to import into")
		cacheFingerprint = importFlags.Bool("cache-fingerprint", false, "Enable fingerprinting for caching")
		followSymlinks   = importFlags.Bool("follow-symlinks", false, "Follow symlinks during import")
	)
	if err := importFlags.Parse(args); err != nil {
		fmt.Println("Failed to parse import flags:", err)
		return
	}
	// Validate required flag
	if *src == "" {
		fmt.Println("--src is required for import")
		importFlags.Usage()
		return
	}
	fmt.Printf("[import] src=%s, collection=%s, cacheFingerprint=%v, followSymlinks=%v, dryRun=%v, config=%s\n",
		*src, *collection, *cacheFingerprint, *followSymlinks, dryRun, config,
	)
	// TODO: Add core import logic here
}
