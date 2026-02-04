package imex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeImportSummaryFromOptions(t *testing.T) {
	tmp := t.TempDir()
	// create collections: a/x.md, b/y.md
	os.MkdirAll(filepath.Join(tmp, "a"), 0o755)
	os.MkdirAll(filepath.Join(tmp, "b"), 0o755)
	os.WriteFile(filepath.Join(tmp, "a", "x.md"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(tmp, "b", "y.md"), []byte("yy"), 0o644)

	opts := ImportOptions{
		Dir:            tmp,
		Extensions:     []string{".md"},
		PathChanBuffer: DefaultPathChanBuffer,
		Collection:     "imports",
	}

	sum, err := ComputeImportSummaryFromOptions(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sum.TotalFiles != 2 {
		t.Fatalf("expected 2 files, got %d", sum.TotalFiles)
	}
	if len(sum.Collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(sum.Collections))
	}
}

func TestImporterComputeImportSummaryIntegration(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "c"), 0o755)
	os.WriteFile(filepath.Join(tmp, "c", "f.md"), []byte("f"), 0o644)

	imp := NewImporter(ImportOptions{
		Dir:        tmp,
		Extensions: []string{".md"},
	})
	defer imp.Stop()

	// Simulate a single file processed
	pd := FileData{Path: filepath.Join(tmp, "c", "f.md"), Size: 1}
	imp.recordFileForSummary(pd)

	sum := imp.ComputeImportSummary()
	if sum.TotalFiles != 1 {
		t.Fatalf("expected 1 total file, got %d", sum.TotalFiles)
	}
	if len(sum.Collections) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(sum.Collections))
	}
}
