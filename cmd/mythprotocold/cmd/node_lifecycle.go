package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// This checksum is the trust anchor distributed with the Phase 1 release.
// A downloaded genesis is not trusted merely because its adjacent checksum agrees.
var officialGenesisSHA256 = "0e9c91fbfdef9e42f27a3f64326ad5b4ef0afb45e086b4c6c945ba8a85d6e3cf"

type genesisIdentity struct {
	ChainID string `json:"chain_id"`
}

type localTestnetManifest struct {
	ChainID       string `json:"chain_id"`
	GenesisSHA256 string `json:"genesis_sha256"`
}

func NewInitNodeCmd(defaultHome string) *cobra.Command {
	var genesisPath string
	cmd := &cobra.Command{
		Use:   "init-node",
		Short: "Initialize a MythChain node from the official genesis",
		RunE: func(cmd *cobra.Command, _ []string) error {
			home, err := nodeHome(cmd, defaultHome)
			if err != nil {
				return err
			}
			source, err := resolveGenesisPath(genesisPath)
			if err != nil {
				return err
			}
			if err := InitializeNode(home, source); err != nil {
				return err
			}
			cmd.Printf("Node initialized at %s\nNext step — join the network:\n  mythprotocold join --home \"%s\" --chain-id <CHAIN_ID> --persistent-peers <ID@host:26656>\nThen start the node:\n  mythprotocold start --home \"%s\"\n", home, home, home)
			return nil
		},
	}
	cmd.Flags().StringVar(&genesisPath, "genesis", "", "path to the official genesis.json (defaults to release/genesis.json)")
	return cmd
}

func nodeHome(cmd *cobra.Command, defaultHome string) (string, error) {
	home, err := cmd.Flags().GetString("home")
	if err != nil || home == "" {
		home = defaultHome
	}
	if home == "" {
		return "", errors.New("node home is empty; provide --home")
	}
	return filepath.Abs(home)
}

func resolveGenesisPath(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	if executable, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executable)
		for _, candidate := range []string{
			filepath.Join(executableDir, "release", "genesis.json"),
			filepath.Clean(filepath.Join(executableDir, "..", "share", "mythprotocold", "release", "genesis.json")),
		} {
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}
	candidate := filepath.Join("release", "genesis.json")
	if _, err := os.Stat(candidate); err != nil {
		return "", fmt.Errorf("official genesis not found at %s; provide --genesis", candidate)
	}
	return filepath.Abs(candidate)
}

// InitializeNode prepares an isolated node home, verifies the release genesis,
// installs configuration, and creates/loads the persistent CometBFT node key.
func InitializeNode(home, sourceGenesis string) error {
	home, err := filepath.Abs(home)
	if err != nil {
		return err
	}
	genesis, err := os.ReadFile(sourceGenesis)
	if err != nil {
		return fmt.Errorf("read official genesis: %w", err)
	}
	if err := verifyOfficialGenesis(genesis); err != nil {
		return err
	}
	var identity genesisIdentity
	if err := json.Unmarshal(genesis, &identity); err != nil || identity.ChainID == "" {
		return errors.New("official genesis has no valid chain_id")
	}

	configDir := filepath.Join(home, "config")
	for _, dir := range []string{configDir, filepath.Join(home, "data"), filepath.Join(home, "keys")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create node directory %s: %w", dir, err)
		}
	}
	genesisDest := filepath.Join(configDir, "genesis.json")
	if err := writeExclusiveOrSame(genesisDest, genesis); err != nil {
		return fmt.Errorf("install genesis: %w", err)
	}
	checksum := []byte(officialGenesisSHA256 + "\n")
	if err := writeExclusiveOrSame(filepath.Join(configDir, "genesis.sha256"), checksum); err != nil {
		return fmt.Errorf("install trusted genesis checksum: %w", err)
	}

	cfg := initCometBFTConfig()
	cfg.SetRoot(home)
	cfg.Moniker = filepath.Base(home)
	cfg.FilterPeers = true
	cfg.P2P.SeedMode = false
	cfg.P2P.PexReactor = true
	configPath := filepath.Join(configDir, cmtcfg.DefaultConfigFileName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		cmtcfg.WriteConfigFile(configPath, cfg)
	} else if err != nil {
		return err
	}
	appConfigPath := filepath.Join(configDir, "app.toml")
	if _, err := os.Stat(appConfigPath); errors.Is(err, os.ErrNotExist) {
		appCfg := config.DefaultConfig()
		appCfg.MinGasPrices = "0.001umtc"
		tpl, err := template.New("app.toml").Parse(config.DefaultConfigTemplate)
		if err != nil {
			return fmt.Errorf("parse app.toml template: %w", err)
		}
		var rendered bytes.Buffer
		if err := tpl.Execute(&rendered, appCfg); err != nil {
			return fmt.Errorf("render app.toml: %w", err)
		}
		if err := os.WriteFile(appConfigPath, rendered.Bytes(), 0o600); err != nil {
			return fmt.Errorf("write app.toml: %w", err)
		}
	} else if err != nil {
		return err
	}

	if _, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile()); err != nil {
		return fmt.Errorf("load or create persistent node identity: %w", err)
	}
	return nil
}

func writeExclusiveOrSame(path string, contents []byte) error {
	if existing, err := os.ReadFile(path); err == nil {
		if string(existing) == string(contents) {
			return nil
		}
		return fmt.Errorf("%s already exists with different contents; refusing to overwrite", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err = f.Write(contents); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func verifyOfficialGenesis(data []byte) error {
	hash := sha256.Sum256(data)
	actual := hex.EncodeToString(hash[:])
	if actual != officialGenesisSHA256 {
		return fmt.Errorf("genesis rejected: checksum does not match official MythChain genesis (got %s)", actual)
	}
	return nil
}

// writeLocalTestnetManifest marks genesis files produced by the explicit
// multi-node testnet generator. It does not alter the official release trust
// anchor and the manifest is checked against the exact local genesis bytes.
func writeLocalTestnetManifest(home, genesisPath string) error {
	genesis, err := os.ReadFile(genesisPath)
	if err != nil {
		return err
	}
	var identity genesisIdentity
	if err := json.Unmarshal(genesis, &identity); err != nil || identity.ChainID == "" {
		return errors.New("generated testnet genesis is invalid or missing chain_id")
	}
	hash := sha256.Sum256(genesis)
	manifest := localTestnetManifest{ChainID: identity.ChainID, GenesisSHA256: hex.EncodeToString(hash[:])}
	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, "config", "local-testnet.json"), append(data, '\n'), 0o600)
}

func writeGenesisChecksumFile(genesisPath string) error {
	genesis, err := os.ReadFile(genesisPath)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(genesis)
	checksumPath := filepath.Join(filepath.Dir(genesisPath), "genesis.sha256")
	return os.WriteFile(checksumPath, []byte(hex.EncodeToString(hash[:])+"\n"), 0o644)
}

func verifyNodeGenesis(home string, genesis []byte) error {
	if err := verifyOfficialGenesis(genesis); err == nil {
		return nil
	}
	manifestBytes, err := os.ReadFile(filepath.Join(home, "config", "local-testnet.json"))
	if err != nil {
		return errors.New("genesis rejected: checksum does not match the official release")
	}
	var manifest localTestnetManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return fmt.Errorf("invalid local testnet manifest: %w", err)
	}
	hash := sha256.Sum256(genesis)
	if manifest.GenesisSHA256 != hex.EncodeToString(hash[:]) {
		return errors.New("local testnet genesis hash does not match its manifest")
	}
	var identity genesisIdentity
	if err := json.Unmarshal(genesis, &identity); err != nil || identity.ChainID == "" || manifest.ChainID != identity.ChainID {
		return errors.New("local testnet chain ID does not match its manifest")
	}
	return nil
}

// ValidateNodeHome performs startup integrity checks before CometBFT starts.
func ValidateNodeHome(home string) error {
	configDir := filepath.Join(home, "config")
	genesisPath := filepath.Join(configDir, "genesis.json")
	genesis, err := os.ReadFile(genesisPath)
	if err != nil {
		return fmt.Errorf("node is not initialized: read %s: %w (run init-node)", genesisPath, err)
	}
	if err := verifyNodeGenesis(home, genesis); err != nil {
		return err
	}
	var genesisID genesisIdentity
	if err := json.Unmarshal(genesis, &genesisID); err != nil || genesisID.ChainID == "" {
		return errors.New("node genesis is invalid or missing chain_id")
	}
	cfg := cmtcfg.DefaultConfig()
	cfg.SetRoot(home)
	if _, err := readCometConfig(filepath.Join(configDir, cmtcfg.DefaultConfigFileName)); err != nil {
		return fmt.Errorf("load node configuration: %w", err)
	}
	nodeKey, err := p2p.LoadNodeKey(filepath.Join(configDir, cmtcfg.DefaultNodeKeyName))
	if err != nil {
		return fmt.Errorf("invalid or missing node identity: %w", err)
	}
	if nodeKey == nil || nodeKey.PrivKey == nil || nodeKey.ID() == "" {
		return errors.New("invalid node identity: empty node ID")
	}
	return nil
}

func NewJoinNetworkCmd(defaultHome string) *cobra.Command {
	var peers, seeds, chainID string
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join an existing MythChain network using the official genesis",
		RunE: func(cmd *cobra.Command, _ []string) error {
			home, err := nodeHome(cmd, defaultHome)
			if err != nil {
				return err
			}
			if err := ValidateNodeHome(home); err != nil {
				return fmt.Errorf("join rejected: %w", err)
			}
			genesis, err := os.ReadFile(filepath.Join(home, "config", cmtcfg.DefaultGenesisJSONName))
			if err != nil {
				return err
			}
			var genesisID genesisIdentity
			if err := json.Unmarshal(genesis, &genesisID); err != nil {
				return err
			}
			if chainID != "" && chainID != genesisID.ChainID {
				return fmt.Errorf("join rejected: requested chain ID %q does not match official genesis %q", chainID, genesisID.ChainID)
			}
			if strings.TrimSpace(peers) == "" && strings.TrimSpace(seeds) == "" {
				return errors.New("join requires at least one --persistent-peers or --seeds address")
			}
			if err := validatePeerList(peers, seeds); err != nil {
				return fmt.Errorf("join rejected: %w", err)
			}
			configPath := filepath.Join(home, "config", cmtcfg.DefaultConfigFileName)
			loaded, err := readCometConfig(configPath)
			if err != nil {
				return err
			}
			loaded.P2P.PersistentPeers = strings.TrimSpace(peers)
			loaded.P2P.Seeds = strings.TrimSpace(seeds)
			loaded.P2P.PexReactor = true
			cmtcfg.WriteConfigFile(configPath, loaded)
			cmd.Printf("Node joined the network configuration.\nNext step — start the node:\n  mythprotocold start --home \"%s\"\n", home)
			return nil
		},
	}
	cmd.Flags().StringVar(&peers, "persistent-peers", "", "comma-separated CometBFT persistent peer multiaddrs (ID@host:port)")
	cmd.Flags().StringVar(&seeds, "seeds", "", "comma-separated CometBFT seed peer multiaddrs (ID@host:port)")
	cmd.Flags().StringVar(&chainID, "chain-id", "", "expected chain ID; must match the official genesis")
	return cmd
}

func validatePeerList(peers, seeds string) error {
	for _, list := range []struct{ label, value string }{{"persistent peer", peers}, {"seed", seeds}} {
		for _, addr := range strings.Split(list.value, ",") {
			addr = strings.TrimSpace(addr)
			if addr == "" {
				continue
			}
			parsed, err := p2p.NewNetAddressString(addr)
			if err != nil {
				return fmt.Errorf("invalid %s address %q: %w", list.label, addr, err)
			}
			if parsed.ID == "" {
				return fmt.Errorf("%s address %q must include the peer ID", list.label, addr)
			}
		}
	}
	return nil
}

func readCometConfig(path string) (*cmtcfg.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := cmtcfg.DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
