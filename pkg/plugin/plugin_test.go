package plugin

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
)

func makeExecutable(t *testing.T, dir, name string) string {
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	f.Close()
	if runtime.GOOS == "windows" {
		// On Windows, just having .exe is enough
		return path
	}
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	return path
}

func TestScanPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	// Create fake plugins
	makeExecutable(t, tmpDir, "devctl-foo")
	makeExecutable(t, tmpDir, "devctl-bar")
	makeExecutable(t, tmpDir, "devctl")                                         // should be ignored
	os.WriteFile(filepath.Join(tmpDir, "devctl-baz"), []byte("not exec"), 0644) // not executable
	os.Mkdir(filepath.Join(tmpDir, "devctl-dir"), 0755)                         // directory

	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", tmpDir)

	plugins := scanPlugins("devctl-")
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}
	if _, ok := plugins["foo"]; !ok {
		t.Error("expected plugin 'foo' to be found")
	}
	if _, ok := plugins["bar"]; !ok {
		t.Error("expected plugin 'bar' to be found")
	}
	if _, ok := plugins["baz"]; ok {
		t.Error("did not expect 'baz' (not executable) to be found")
	}
}

func TestRegisterPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	makeExecutable(t, tmpDir, "devctl-hello")
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", tmpDir)

	root := &cobra.Command{Use: "devctl"}
	RegisterPlugins(root, "devctl-")

	cmd, _, err := root.Find([]string{"hello"})
	if err != nil {
		t.Fatalf("could not find registered plugin command: %v", err)
	}
	if cmd == nil || cmd.Use != "hello" {
		t.Errorf("expected to find 'hello' command, got %v", cmd)
	}
}
