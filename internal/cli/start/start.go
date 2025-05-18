package start

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/themilchenko/kv/internal/config"

	"github.com/spf13/cobra"
)

// save PID to file in dataDir
func savePID(dataDir string, pid int) error {
	pidFile := filepath.Join(dataDir, "node.pid")
	pidStr := fmt.Sprintf("%d", pid)

	if err := os.WriteFile(pidFile, []byte(pidStr), 0o644); err != nil {
		return fmt.Errorf("cannot write pid file %s: %w", pidFile, err)
	}

	return nil
}

func startNode(cfg *config.Config, node *config.Node) error {
	dataDir := filepath.Join(cfg.DataDir, node.Alias)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("cannot create data dir %s: %w", dataDir, err)
	}

	bin := filepath.Join(cfg.BinPath, "node")
	logFile := filepath.Join(dataDir, "node.log")

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("cannot open log %s: %w", logFile, err)
	}

	var cmd *exec.Cmd

	if node.Alias == cfg.Leader {
		log.Printf("Starting leader %q...", node.Alias)

		cmd = exec.Command(bin,
			"-id", node.Alias,
			"-haddr", node.HttpAddress,
			"-raddr", node.RpcAddress,
			dataDir,
		)
	} else {
		log.Printf("Starting follower %q...", node.Alias)

		leaderHttpAddr := ""
		for _, n := range cfg.Cluster {
			if n.Alias == cfg.Leader {
				leaderHttpAddr = n.HttpAddress

				break
			}
		}

		cmd = exec.Command(bin,
			"-id", node.Alias,
			"-haddr", node.HttpAddress,
			"-raddr", node.RpcAddress,
			"-join", leaderHttpAddr,
			dataDir,
		)
	}

	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start node %s: %w", node.Alias, err)
	}

	pid := cmd.Process.Pid
	log.Printf("Started %q (pid %d)", node.Alias, pid)

	if err := savePID(dataDir, pid); err != nil {
		return err
	}

	return nil
}

// Cmd returns the start command
func Cmd(cfg **config.Config, binPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the cluster nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := *cfg // dereference loaded config

			// launch leader then followers
			var leaderNode *config.Node
			for _, n := range c.Cluster {
				if n.Alias == c.Leader {
					leaderNode = n
					break
				}
			}

			if leaderNode == nil {
				return fmt.Errorf("leader %s not in cluster", c.Leader)
			}

			if err := startNode(c, leaderNode); err != nil {
				return err
			}

			time.Sleep(2 * time.Second)

			for _, n := range c.Cluster {
				if n.Alias == c.Leader {
					continue
				}

				if err := startNode(c, n); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
