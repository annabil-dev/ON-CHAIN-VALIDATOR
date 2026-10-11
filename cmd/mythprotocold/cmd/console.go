package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// enableFriendlyStartConsole applies only to an interactive 'start' command.
// Services, piped automation, and operators explicitly choosing --log_level
// keep the existing Cosmos SDK logging behavior unchanged.
func enableFriendlyStartConsole(cmd *cobra.Command) error {
	if cmd.Name() != "start" || !isatty.IsTerminal(os.Stderr.Fd()) {
		return nil
	}
	if flag := cmd.Flag("log_level"); flag == nil || flag.Changed {
		return nil
	}
	if os.Getenv("MYTHPROTOCOL_LOG_LEVEL") != "" {
		return nil
	}
	if err := cmd.Flags().Set("log_level", "error"); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Mythchain node | compact console (full logs: --log_level info)")
	fmt.Fprintln(os.Stderr, "Starting and connecting to peers... Ctrl+C to stop.")
	go followNodeProgress(cmd.Context(), os.Stderr, "http://127.0.0.1:26657")
	return nil
}

type nodeProgress struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
			CatchingUp        bool   `json:"catching_up"`
		} `json:"sync_info"`
	} `json:"result"`
}

type nodePeers struct {
	Result struct {
		NPeers string `json:"n_peers"`
	} `json:"result"`
}

func getNodeStatus(ctx context.Context, client *http.Client, origin string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/status", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("RPC returned HTTP %d", resp.StatusCode)
	}
	var status nodeProgress
	if err := json.NewDecoder(io.LimitReader(resp.Body, 512*1024)).Decode(&status); err != nil {
		return "", err
	}
	height := status.Result.SyncInfo.LatestBlockHeight
	if height == "" {
		return "", fmt.Errorf("RPC status missing height")
	}
	mode := "SYNCING"
	if !status.Result.SyncInfo.CatchingUp {
		mode = "SYNCED"
	}
	return fmt.Sprintf("%s | height %s", mode, height), nil
}

func getPeerCount(ctx context.Context, client *http.Client, origin string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/net_info", nil)
	if err != nil {
		return "-"
	}
	resp, err := client.Do(req)
	if err != nil {
		return "-"
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "-"
	}
	var peers nodePeers
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(&peers); err != nil {
		return "-"
	}
	if peers.Result.NPeers == "" {
		return "-"
	}
	return peers.Result.NPeers
}

func followNodeProgress(ctx context.Context, writer io.Writer, origin string) {
	client := &http.Client{Timeout: 3 * time.Second}
	ticker := time.NewTicker(8 * time.Second)
	defer ticker.Stop()
	// The app server is not immediately ready after the startup pre-run.
	ready := time.NewTimer(4 * time.Second)
	defer ready.Stop()
	select {
	case <-ctx.Done():
		return
	case <-ready.C:
	}
	lastLine := ""
	for {
		message, err := getNodeStatus(ctx, client, origin)
		switch {
		case err != nil:
			message = "WAITING FOR RPC | node starting or custom RPC address"
		default:
			message += " | peers " + getPeerCount(ctx, client, origin)
		}
		message = strings.TrimSpace(message)
		if message != lastLine {
			// Short padded CR update stays on one line in standard 80-column
			// consoles. No cursor controls or ANSI support are required.
			fmt.Fprintf(writer, "\r%-72.72s", message)
			lastLine = message
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
