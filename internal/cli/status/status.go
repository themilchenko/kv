package status

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/themilchenko/kv/internal/config"

	"github.com/spf13/cobra"
)

// statusResponse represents the JSON returned by the node's /status endpoint
type statusResponse struct {
	Me        nodeInfo   `json:"me"`
	Leader    nodeInfo   `json:"leader"`
	Followers []nodeInfo `json:"followers"`
}

type nodeInfo struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

// Cmd returns the status command.
func Cmd(cfg **config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := *cfg

			errs := make([]error, 0)
			var status statusResponse
			inactiveAddrs := make(map[string]struct{})

			for _, n := range c.Cluster {
				url := fmt.Sprintf("http://%s/status", n.HttpAddress)

				resp, err := http.Get(url)
				if err != nil {
					inactiveAddrs[n.Alias] = struct{}{}
					errs = append(errs, fmt.Errorf("failed to query status endpoint %s: %w", url, err))

					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return fmt.Errorf("unable to read response body: %w", err)
					}

					if err := json.Unmarshal(body, &status); err != nil {
						return fmt.Errorf("invalid JSON from status endpoint: %w", err)
					}

					break
				}

				errs = append(errs, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url))
			}

			if len(errs) == len(c.Cluster) {
				return errors.Join(errs...)
			}

			fmt.Printf("Leader:   %s (%s)\n", status.Leader.ID, status.Leader.Address)
			fmt.Println("Followers:")
			for _, f := range status.Followers {
				if _, ok := inactiveAddrs[f.ID]; !ok {
					fmt.Printf("  - %s (%s)\n", f.ID, f.Address)
				}
			}

			return nil
		},
	}
}
