package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// VerifyGenesisChecksum validates a genesis file against its sha256 checksum.
func VerifyGenesisChecksum(genesisPath string, checksumPath string) error {
	data, err := os.ReadFile(genesisPath)
	if err != nil {
		return err
	}

	expectedBytes, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	actual := hex.EncodeToString(hash[:])
	expected := string(expectedBytes)

	if len(expected) > 64 {
		expected = expected[:64]
	}

	if actual != expected {
		return fmt.Errorf("genesis checksum mismatch")
	}

	return nil
}
