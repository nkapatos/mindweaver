package main

import (
	"fmt"
	"io"
	"os"

	"github.com/nkapatos/mindweaver/config"
)

func CopyIcons() error {
	srcDir := "node_modules/lucide-static/icons"
	// This dir is serving the other asset files bundled from esbuild
	dstDir := "static/icons"

	// Ensure destination directory exists
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
	}

	for _, iconName := range config.SvgIcons {
		srcPath := fmt.Sprintf("%s/%s.svg", srcDir, iconName)
		dstPath := fmt.Sprintf("%s/%s.svg", dstDir, iconName)

		srcFile, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", srcPath, err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", dstPath, err)
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return fmt.Errorf("failed to copy from %s to %s: %w", srcPath, dstPath, err)
		}
	}

	return nil
}

func main() {
	err := CopyIcons()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error copying icons: %v\n", err)
		os.Exit(1)
	}
}
