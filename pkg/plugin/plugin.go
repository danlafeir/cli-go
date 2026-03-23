package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// scanPlugins finds all <prefix>* executables in PATH and returns a map of plugin name to full path.
// prefix should include the trailing dash, e.g. "devctl-".
func scanPlugins(prefix string) map[string]string {
	plugins := make(map[string]string)
	pathEnv := os.Getenv("PATH")
	base := strings.TrimSuffix(prefix, "-")
	for _, dir := range filepath.SplitList(pathEnv) {
		matches, err := filepath.Glob(filepath.Join(dir, prefix+"*"))
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			if info.Mode()&0111 == 0 { // not executable
				continue
			}
			name := filepath.Base(match)
			if name == base || name == base+".exe" {
				continue
			}
			if len(name) > len(prefix) {
				pluginName := name[len(prefix):]
				plugins[pluginName] = match
			}
		}
	}
	return plugins
}

// RegisterPlugins scans for plugins and registers them as subcommands on the given root command.
// prefix should include the trailing dash, e.g. "devctl-".
func RegisterPlugins(rootCmd *cobra.Command, prefix string) {
	plugins := scanPlugins(prefix)
	for name, path := range plugins {
		pluginName := name // capture for closure
		pluginPath := path
		pluginCmd := &cobra.Command{
			Use:   pluginName,
			Short: "Plugin: " + pluginName,
			Run: func(cmd *cobra.Command, args []string) {
				c := exec.Command(pluginPath, args...)
				c.Stdin = os.Stdin
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					cmd.PrintErrf("Plugin %s failed: %v\n", pluginName, err)
					os.Exit(1)
				}
			},
		}
		rootCmd.AddCommand(pluginCmd)
	}
}
