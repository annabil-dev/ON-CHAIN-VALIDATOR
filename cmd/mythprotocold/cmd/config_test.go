package cmd

import (
	"testing"
	"time"
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
