package stop

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/themilchenko/kv/internal/config"
)

// readPID reads the pid from node.pid file in the given dataDir
func readPID(dataDir string) (int, error) {
	pidFile := filepath.Join(dataDir, "node.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("cannot read pid file %s: %w", pidFile, err)
	}
	pidStr := string(data)
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid pid %q in file %s: %w", pidStr, pidFile, err)
	}
	return pid, nil
}

// Cmd returns the stop command
func Cmd(cfg **config.Config, _ *string) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the cluster nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := *cfg
			// stop followers then leader to avoid split-brain
			for _, n := range c.Cluster {
				if n.Alias == c.Leader {
					continue
				}
				dataDir := filepath.Join(c.DataDir, n.Alias)
				pid, err := readPID(dataDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: %v\n", err)
					continue
				}
				proc, err := os.FindProcess(pid)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: cannot find process %d: %v\n", pid, err)
					continue
				}
				if err := proc.Signal(syscall.SIGTERM); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to terminate %s (pid %d): %v\n", n.Alias, pid, err)
				} else {
					fmt.Printf("Stopped %s (pid %d)\n", n.Alias, pid)
				}
			}
			// now stop leader
			// find leader node
			var leaderNode *config.Node
			for _, n := range c.Cluster {
				if n.Alias == c.Leader {
					leaderNode = n
					break
				}
			}
			if leaderNode != nil {
				dataDir := filepath.Join(c.DataDir, leaderNode.Alias)
				pid, err := readPID(dataDir)
				if err != nil {
					return err
				}
				proc, err := os.FindProcess(pid)
				if err != nil {
					return fmt.Errorf("cannot find leader process %d: %w", pid, err)
				}
				if err := proc.Signal(syscall.SIGTERM); err != nil {
					if errors.Is(err, os.ErrProcessDone) {
						fmt.Println(err)
					} else {
						return fmt.Errorf("failed to terminate leader %s (pid %d): %w", leaderNode.Alias, pid, err)
					}
				}
				fmt.Printf("Stopped leader %s (pid %d)\n", leaderNode.Alias, pid)
			}
			return nil
		},
	}
}
