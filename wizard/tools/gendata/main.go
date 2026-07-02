// Command gendata copies the root-level declarative data (packages/, profiles.yaml)
// into the wizard module so //go:embed can bake it into the binary. Go embed cannot
// reference files outside the module directory, so this bridges that boundary at
// build time. Invoked via `go generate ./...` and the GoReleaser before hook.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	// Repo root is one level above the wizard module (this tool runs from wizard/).
	root := filepath.Join("..")

	if err := copyDir(filepath.Join(root, "packages"), "packages"); err != nil {
		fatal("copy packages: %v", err)
	}
	if err := copyFile(filepath.Join(root, "profiles.yaml"), "profiles.yaml"); err != nil {
		fatal("copy profiles.yaml: %v", err)
	}
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "gendata: "+format+"\n", a...)
	os.Exit(1)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue // catalog is flat; ignore nested dirs
		}
		if err := copyFile(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
