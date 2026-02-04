package imex

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// helper collects paths produced by walker.Walk concurrently and returns them and any error
func collectPaths(ctx context.Context, w *Walker) ([]string, error) {
	pathsCh := make(chan string, DefaultPathChanBuffer)
	var wg sync.WaitGroup
	var walkErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, walkErr = w.Walk(ctx, pathsCh)
	}()

	var results []string
	for p := range pathsCh {
		results = append(results, p)
	}

	wg.Wait()
	return results, walkErr
}

func TestWalkRecurses(t *testing.T) {
	tmp := t.TempDir()
	// create nested structure
	os.MkdirAll(filepath.Join(tmp, "a/nested"), 0o755)
	os.MkdirAll(filepath.Join(tmp, "b"), 0o755)

	os.WriteFile(filepath.Join(tmp, "a", "file1.md"), []byte("one"), 0o644)
	os.WriteFile(filepath.Join(tmp, "a", "nested", "file2.md"), []byte("two"), 0o644)
	os.WriteFile(filepath.Join(tmp, "b", "file3.txt"), []byte("txt"), 0o644)

	w := NewWalker(ImportOptions{
		Dir:        tmp,
		Extensions: []string{".md"},
	})

	paths, err := collectPaths(context.Background(), w)
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	// expect two .md files
	if len(paths) != 2 {
		t.Fatalf("expected 2 md files, got %d: %v", len(paths), paths)
	}
}

func TestWalkFollowSymlinksFalseAndTrue(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "target")
	os.MkdirAll(targetDir, 0o755)
	os.WriteFile(filepath.Join(targetDir, "t.md"), []byte("t"), 0o644)

	// symlink to file
	linkFile := filepath.Join(tmp, "link.md")
	os.Symlink(filepath.Join(targetDir, "t.md"), linkFile)

	// symlink to dir
	linkDir := filepath.Join(tmp, "linkdir")
	os.Symlink(targetDir, linkDir)

	// FollowSymlinks = false
	w1 := NewWalker(ImportOptions{
		Dir:            tmp,
		Extensions:     []string{".md"},
		FollowSymlinks: false,
	})
	paths1, err := collectPaths(context.Background(), w1)
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
	// When not following symlinks, link.md and linkdir contents should be skipped
	for _, p := range paths1 {
		if filepath.Base(p) == "link.md" || filepath.Base(p) == "t.md" && filepath.Dir(p) == linkDir {
			t.Fatalf("unexpected symlinked path when FollowSymlinks=false: %s", p)
		}
	}

	// FollowSymlinks = true
	w2 := NewWalker(ImportOptions{
		Dir:            tmp,
		Extensions:     []string{".md"},
		FollowSymlinks: true,
	})
	paths2, err := collectPaths(context.Background(), w2)
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	// Expect to see link.md (the symlink) reported and t.md from the resolved dir
	foundLink := false
	foundT := false
	for _, p := range paths2 {
		if filepath.Base(p) == "link.md" {
			foundLink = true
		}
		if filepath.Base(p) == "t.md" {
			foundT = true
		}
	}
	if !foundLink || !foundT {
		t.Fatalf("expected symlinked entries when FollowSymlinks=true; got %v", paths2)
	}
}

func TestWalkSymlinkLoop(t *testing.T) {
	tmp := t.TempDir()
	dir1 := filepath.Join(tmp, "d1")
	dir2 := filepath.Join(dir1, "d2")
	os.MkdirAll(dir2, 0o755)
	os.WriteFile(filepath.Join(dir1, "f.md"), []byte("f"), 0o644)
	// create symlink from d2 back to d1
	os.Symlink(dir1, filepath.Join(dir2, "back"))

	w := NewWalker(ImportOptions{
		Dir:            dir1,
		Extensions:     []string{".md"},
		FollowSymlinks: true,
	})

	paths, err := collectPaths(context.Background(), w)
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	// Should find the single f.md exactly once
	count := 0
	for _, p := range paths {
		if filepath.Base(p) == "f.md" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected f.md once, found %d occurrences; paths: %v", count, paths)
	}
}

func TestWalkMaxFileSize(t *testing.T) {
	tmp := t.TempDir()
	// create a file slightly larger than DefaultMaxFileSize
	big := filepath.Join(tmp, "big.md")
	f, err := os.Create(big)
	if err != nil {
		t.Fatalf("create big file: %v", err)
	}
	// write DefaultMaxFileSize + 1 bytes
	data := make([]byte, DefaultMaxFileSize+1)
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write big file: %v", err)
	}
	f.Close()

	w := NewWalker(ImportOptions{
		Dir:         tmp,
		Extensions:  []string{".md"},
		MaxFileSize: DefaultMaxFileSize,
	})

	paths, err := collectPaths(context.Background(), w)
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected big file to be skipped by MaxFileSize; got paths: %v", paths)
	}
}
