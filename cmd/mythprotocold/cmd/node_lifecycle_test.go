package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/p2p"
)

func TestInitializeNodePersistsIdentityAndValidatesHome(t *testing.T) {
	genesis := filepath.Join("..", "..", "..", "release", "genesis.json")
	home := filepath.Join(t.TempDir(), ".mythprotocol")
	if err := InitializeNode(home, genesis); err != nil {
		t.Fatalf("initialize node: %v", err)
	}
	keyPath := filepath.Join(home, "config", "node_key.json")
	keyBefore, err := p2p.LoadNodeKey(keyPath)
	if err != nil {
		t.Fatalf("load node key after init: %v", err)
	}
	if err := InitializeNode(home, genesis); err != nil {
		t.Fatalf("repeat initialization: %v", err)
	}
	keyAfter, err := p2p.LoadNodeKey(keyPath)
	if err != nil {
		t.Fatalf("load node key after restart: %v", err)
	}
	if keyBefore.ID() != keyAfter.ID() {
		t.Fatalf("node ID changed across restart: %s != %s", keyBefore.ID(), keyAfter.ID())
	}
	if err := ValidateNodeHome(home); err != nil {
		t.Fatalf("validate initialized node: %v", err)
	}
	cmtConfig, err := readCometConfig(filepath.Join(home, "config", "config.toml"))
	if err != nil {
		t.Fatalf("read initialized CometBFT config: %v", err)
	}
	if !cmtConfig.FilterPeers {
		t.Fatal("peer allowlist filtering is disabled")
	}
	appToml, err := os.ReadFile(filepath.Join(home, "config", "app.toml"))
	if err != nil {
		t.Fatalf("read app.toml: %v", err)
	}
	if !strings.Contains(string(appToml), `minimum-gas-prices = "0.001umtc"`) || strings.Contains(string(appToml), "{{") {
		t.Fatal("app.toml must render usable default gas price, not an unexpanded template")
	}
	for _, path := range []string{
		filepath.Join(home, "config", "genesis.json"),
		filepath.Join(home, "config", "config.toml"),
		filepath.Join(home, "config", "app.toml"),
		keyPath,
		filepath.Join(home, "data"),
		filepath.Join(home, "keys"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected initialized node path %s: %v", path, err)
		}
	}
}

func TestInitializeNodeRejectsModifiedGenesis(t *testing.T) {
	genesis := filepath.Join("..", "..", "..", "release", "genesis.json")
	data, err := os.ReadFile(genesis)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, ' ')
	tampered := filepath.Join(t.TempDir(), "genesis.json")
	if err := os.WriteFile(tampered, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := InitializeNode(filepath.Join(t.TempDir(), "node"), tampered); err == nil {
		t.Fatal("modified official genesis was accepted")
	}
}

func TestLocalTestnetManifestBindsChainIDAndGenesisHash(t *testing.T) {
	home := filepath.Join(t.TempDir(), "node")
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	genesisPath := filepath.Join(t.TempDir(), "genesis.json")
	genesis := []byte(`{"chain_id":"local-test-1","app_state":{}}`)
	if err := os.WriteFile(genesisPath, genesis, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeLocalTestnetManifest(home, genesisPath); err != nil {
		t.Fatalf("write local testnet manifest: %v", err)
	}
	if err := verifyNodeGenesis(home, genesis); err != nil {
		t.Fatalf("verify marked local testnet genesis: %v", err)
	}
	if err := verifyNodeGenesis(home, append(genesis, ' ')); err == nil {
		t.Fatal("modified local testnet genesis was accepted")
	}
}

func TestValidateNodeHomeRejectsInvalidIdentity(t *testing.T) {
	genesis := filepath.Join("..", "..", "..", "release", "genesis.json")
	home := filepath.Join(t.TempDir(), "node")
	if err := InitializeNode(home, genesis); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "config", "node_key.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateNodeHome(home); err == nil {
		t.Fatal("invalid node identity was accepted")
	}
}

func TestValidatePeerListRejectsMalformedAddresses(t *testing.T) {
	if err := validatePeerList("not-a-peer", ""); err == nil {
		t.Fatal("malformed peer address was accepted")
	}
	if err := validatePeerList("", "0123456789abcdef0123456789abcdef01234567@127.0.0.1:26656"); err != nil {
		t.Fatalf("valid seed address rejected: %v", err)
	}
}

func TestJoinRequiresPeerBeforeChangingConfiguration(t *testing.T) {
	genesis := filepath.Join("..", "..", "..", "release", "genesis.json")
	home := filepath.Join(t.TempDir(), "node")
	if err := InitializeNode(home, genesis); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(home, "config", "config.toml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	joinCmd := NewJoinNetworkCmd(home)
	if err := joinCmd.Execute(); err == nil {
		t.Fatal("join without a peer was accepted")
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("failed join modified node configuration")
	}
}

func TestJoinConfiguresAndRestrictsBootstrapPeer(t *testing.T) {
	genesis := filepath.Join("..", "..", "..", "release", "genesis.json")
	home := filepath.Join(t.TempDir(), "node")
	if err := InitializeNode(home, genesis); err != nil {
		t.Fatal(err)
	}
	peer := "0123456789abcdef0123456789abcdef01234567@127.0.0.1:26656"
	joinCmd := NewJoinNetworkCmd(home)
	joinCmd.SetArgs([]string{"--persistent-peers", peer})
	if err := joinCmd.Execute(); err != nil {
		t.Fatalf("configure bootstrap peer: %v", err)
	}
	allowed := allowedPeerIDs(filepath.Join(home, "config", "config.toml"))
	if _, ok := allowed["0123456789abcdef0123456789abcdef01234567"]; !ok {
		t.Fatal("configured persistent peer was not added to the peer allowlist")
	}
	if _, ok := allowed["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"]; ok {
		t.Fatal("unconfigured peer was added to the peer allowlist")
	}
}
