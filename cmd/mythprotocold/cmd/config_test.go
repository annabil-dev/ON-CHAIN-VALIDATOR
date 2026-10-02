package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmtconfig "github.com/cometbft/cometbft/config"
)

func TestInitCometBFTConfigTargetsOneMinuteBlocks(t *testing.T) {
	cfg := initCometBFTConfig()

	if cfg.Consensus.TimeoutCommit != time.Minute {
		t.Fatalf("timeout_commit = %s, want %s", cfg.Consensus.TimeoutCommit, time.Minute)
	}
	if !cfg.Consensus.CreateEmptyBlocks {
		t.Fatal("create_empty_blocks is disabled; idle chain will not produce timed blocks")
	}
	if cfg.Consensus.CreateEmptyBlocksInterval != 0 {
		t.Fatalf("create_empty_blocks_interval = %s, want 0s", cfg.Consensus.CreateEmptyBlocksInterval)
	}
}

func TestMultiNodeConfigWritesOneMinuteTimeoutAndWildcardCORS(t *testing.T) {
	cfg := cmtconfig.DefaultConfig()
	cfg.Consensus.TimeoutCommit = 5 * time.Second
	cfg.Consensus.SkipTimeoutCommit = true
	cfg.Consensus.CreateEmptyBlocks = false
	cfg.RPC.CORSAllowedOrigins = []string{"https://old.example"}

	configPath := filepath.Join(t.TempDir(), "config.toml")
	writeMultiNodeConfig(configPath, cfg)
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	for _, expected := range []string{
		`timeout_commit = "1m0s"`,
		`skip_timeout_commit = false`,
		`create_empty_blocks = true`,
		`create_empty_blocks_interval = "0s"`,
		`cors_allowed_origins = ["*", ]`,
	} {
		if !strings.Contains(string(contents), expected) {
			t.Errorf("generated config does not contain %q", expected)
		}
	}
}
