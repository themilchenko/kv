package root

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/themilchenko/kv/internal/cli/start"
	"github.com/themilchenko/kv/internal/cli/status"
	"github.com/themilchenko/kv/internal/cli/stop"
	"github.com/themilchenko/kv/internal/config"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	binPath string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "clusterctl",
	Short: "CLI utility to manage an hraftd cluster",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		path, err := filepath.Abs(cfgFile)
		if err != nil {
			return fmt.Errorf("invalid config file path: %w", err)
		}

		cfg, err = config.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "path to config file (YAML)")

	rootCmd.AddCommand(start.Cmd(&cfg, &binPath))
	rootCmd.AddCommand(stop.Cmd(&cfg, &binPath))
	rootCmd.AddCommand(status.Cmd(&cfg))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
