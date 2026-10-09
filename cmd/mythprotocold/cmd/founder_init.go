package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
)

const founderInitMarker = ".founder-init-incomplete"

type founderManifest struct {
	ChainID           string `json:"chain_id"`
	InitialHeight     int64  `json:"initial_height"`
	GenesisSHA256     string `json:"genesis_sha256"`
	FounderAddress    string `json:"founder_address"`
	FounderAllocation string `json:"founder_allocation"`
	TreasuryAddress   string `json:"treasury_address"`
	TreasuryControl   string `json:"treasury_control"`
	NodeID            string `json:"node_id"`
}

type founderGenesisHeader struct {
	ChainID       string `json:"chain_id"`
	InitialHeight int64  `json:"initial_height"`
}

func NewFounderInitCmd(mbm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	var chainID, confirmedChainID, outputDir, genesisOutput string
	var externalAddress string
	cmd := &cobra.Command{
		Use:   "founder-init",
		Short: "Create the founder's initial MythChain genesis and validator home",
		Long: `Creates a one-validator genesis using the configured chain ID, the fixed Mythchain
token allocations, and a local encrypted file keyring. The chain ID must be entered
twice. This command writes into a new output directory and never overwrites a node home.

The generated genesis is a candidate until its SHA-256 and production parameters are
approved and a release binary is rebuilt with that exact genesis checksum.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateFounderChainID(chainID, confirmedChainID); err != nil {
				return err
			}
			if outputDir == "" {
				return errors.New("--output-dir is required; choose a new, dedicated founder directory")
			}
			externalHost, err := validateFounderEndpoint(externalAddress)
			if err != nil {
				return err
			}
			home, err := filepath.Abs(outputDir)
			if err != nil {
				return err
			}
			if _, err := os.Stat(home); err == nil {
				return fmt.Errorf("founder output directory already exists: %s; choose a fresh path, existing files will not be overwritten", home)
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(home), 0o700); err != nil {
				return err
			}
			if err := os.Mkdir(home, 0o700); err != nil {
				return err
			}
			markerPath := filepath.Join(home, founderInitMarker)
			if err := os.WriteFile(markerPath, []byte("Genesis generation incomplete; inspect before resuming.\n"), 0o600); err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return fmt.Errorf("initialize client context: %w", err)
			}
			serverCtx := server.GetServerContextFromCmd(cmd)
			commissionRate, err := getCommissionFlag(cmd, flagCommissionRate)
			if err != nil {
				return err
			}
			commissionMaxRate, err := getCommissionFlag(cmd, flagCommissionMaxRate)
			if err != nil {
				return err
			}
			commissionMaxChange, err := getCommissionFlag(cmd, flagCommissionMaxChange)
			if err != nil {
				return err
			}
			minGasPrices, err := cmd.Flags().GetString(server.FlagMinGasPrices)
			if err != nil {
				return err
			}
			enablePoUW, err := cmd.Flags().GetBool(flagEnablePoUWEmissions)
			if err != nil {
				return err
			}

			args := initArgs{
				algo:                   string(hd.Secp256k1Type),
				chainID:                chainID,
				keyringBackend:         "file",
				minGasPrices:           minGasPrices,
				nodeDirPrefix:          "validator",
				numValidators:          1,
				outputDir:              home,
				startingIPAddress:      externalHost,
				externalAddress:        strings.TrimSpace(externalAddress),
				validatorsStakesAmount: map[int]sdk.Coin{},
				ports: map[int]string{
					0: "26657",
				},
				enablePoUWEmissions: enablePoUW,
				commissionRate:      commissionRate,
				commissionMaxRate:   commissionMaxRate,
				commissionMaxChange: commissionMaxChange,
			}
			if err := initTestnetFiles(clientCtx, cmd, serverCtx.Config, mbm, genBalIterator, args); err != nil {
				return fmt.Errorf("founder genesis generation failed; incomplete output retained at %s: %w", home, err)
			}

			nodeHome := filepath.Join(home, "validator0")
			genesisPath := filepath.Join(nodeHome, "config", cmtcfg.DefaultGenesisJSONName)
			genesis, err := os.ReadFile(genesisPath)
			if err != nil {
				return fmt.Errorf("read generated founder genesis: %w", err)
			}
			var doc founderGenesisHeader
			if err := json.Unmarshal(genesis, &doc); err != nil {
				return fmt.Errorf("decode generated founder genesis header: %w", err)
			}
			if doc.ChainID != chainID || doc.InitialHeight < 1 {
				return fmt.Errorf("generated genesis identity mismatch: chain_id=%q initial_height=%d", doc.ChainID, doc.InitialHeight)
			}
			audit, err := AuditTreasuryGenesis(genesis)
			if err != nil {
				return fmt.Errorf("audit generated genesis: %w", err)
			}
			if err := ValidateTreasuryAudit(audit); err != nil {
				return fmt.Errorf("generated genesis tokenomics rejected: %w", err)
			}
			if err := writeGenesisChecksumFile(genesisPath); err != nil {
				return fmt.Errorf("write node-home genesis checksum: %w", err)
			}

			if genesisOutput == "" {
				genesisOutput = filepath.Join(home, "genesis.json")
			}
			genesisOutput, err = filepath.Abs(genesisOutput)
			if err != nil {
				return err
			}
			if err := copyFileExclusive(genesisPath, genesisOutput, 0o644); err != nil {
				return fmt.Errorf("write public genesis artifact: %w", err)
			}
			hash := sha256.Sum256(genesis)
			genesisHash := hex.EncodeToString(hash[:])
			checksumOutput := filepath.Join(filepath.Dir(genesisOutput), "genesis.sha256")
			if err := writeExclusive(checksumOutput, []byte(genesisHash+"\n"), 0o644); err != nil {
				return fmt.Errorf("write public genesis checksum: %w", err)
			}
			nodeKey, err := os.ReadFile(filepath.Join(nodeHome, "config", cmtcfg.DefaultNodeKeyName))
			if err != nil {
				return fmt.Errorf("read generated node identity: %w", err)
			}
			var identity struct {
				PrivKey json.RawMessage `json:"priv_key"`
			}
			if err := json.Unmarshal(nodeKey, &identity); err != nil || len(identity.PrivKey) == 0 {
				return errors.New("generated founder node identity is invalid")
			}
			loadedKey, err := p2p.LoadNodeKey(filepath.Join(nodeHome, "config", cmtcfg.DefaultNodeKeyName))
			if err != nil {
				return err
			}
			manifest := founderManifest{
				ChainID:           chainID,
				InitialHeight:     doc.InitialHeight,
				GenesisSHA256:     genesisHash,
				FounderAddress:    audit.FounderAddress,
				FounderAllocation: "1000000 MTC",
				TreasuryAddress:   audit.TreasuryAddress,
				TreasuryControl:   audit.Control,
				NodeID:            string(loadedKey.ID()),
			}
			manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(home, "founder-manifest.json"), append(manifestBytes, '\n'), 0o644); err != nil {
				return err
			}
			if err := os.Remove(markerPath); err != nil {
				return err
			}
			cmd.Printf("Founder genesis candidate created.\nNode home: %s\nChain ID: %s\nInitial height: %d\nGenesis: %s\nGenesis SHA-256: %s\nFounder address: %s\nFounder allocation: 1000000 MTC\nTreasury: %s (20000000 MTC; governance controlled)\nFounder Node ID: %s\n", nodeHome, chainID, doc.InitialHeight, genesisOutput, genesisHash, audit.FounderAddress, audit.TreasuryAddress, manifest.NodeID)
			cmd.Println("Do not start this candidate yet: build/install a release whose linked genesis checksum matches the candidate, and complete the production approval checklist.")
			cmd.Println("The validator/operator keyring is stored in the founder node home. Back up its encrypted keyring and validator keys offline; do not copy them into release artifacts.")
			return nil
		},
	}
	cmd.Flags().StringVar(&chainID, flags.FlagChainID, "", "approved non-testnet chain ID")
	cmd.Flags().StringVar(&confirmedChainID, "confirm-chain-id", "", "re-enter --chain-id exactly to confirm the genesis network")
	cmd.Flags().StringVar(&outputDir, flagOutputDir, "", "new output directory for the founder node (must not already exist)")
	cmd.Flags().StringVar(&genesisOutput, "genesis-output", "", "path for the public genesis.json (default: <output-dir>/genesis.json)")
	cmd.Flags().StringVar(&externalAddress, "external-address", "", "public host:port for the founder node's CometBFT P2P endpoint (required)")
	cmd.Flags().String(server.FlagMinGasPrices, "0.001umtc,0.001uzyra", "minimum gas prices written to app.toml")
	cmd.Flags().Bool(flagEnablePoUWEmissions, false, "Enable PoUW emissions in the generated genesis")
	cmd.Flags().String(flagCommissionRate, "0.05", "founder validator initial commission rate")
	cmd.Flags().String(flagCommissionMaxRate, "0.06", "founder validator maximum commission rate")
	cmd.Flags().String(flagCommissionMaxChange, "0.01", "founder validator maximum commission change per update")
	_ = cmd.MarkFlagRequired(flags.FlagChainID)
	_ = cmd.MarkFlagRequired("confirm-chain-id")
	_ = cmd.MarkFlagRequired(flagOutputDir)
	return cmd
}

var productionChainIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,49}$`)

func validateFounderEndpoint(address string) (string, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(address))
	if err != nil || host == "" {
		return "", errors.New("--external-address must be a real host:port, for example founder.example.org:26656")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return "", errors.New("--external-address port must be a number between 1 and 65535")
	}
	return host, nil
}

func validateFounderChainID(chainID, confirmation string) error {
	if !productionChainIDPattern.MatchString(chainID) {
		return errors.New("chain ID is required and must be 1–50 ASCII letters, digits, '.', '_' or '-'")
	}
	if strings.EqualFold(chainID, "chain-rlkh6n") || strings.EqualFold(chainID, "phase2-local-sync") || strings.HasPrefix(strings.ToLower(chainID), "chain-test") {
		return fmt.Errorf("chain ID %q is reserved for an existing/local test network", chainID)
	}
	if confirmation != chainID {
		return errors.New("--confirm-chain-id must exactly match --chain-id")
	}
	return nil
}

func copyFileExclusive(source, destination string, mode os.FileMode) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	return writeExclusive(destination, data, mode)
}

func writeExclusive(destination string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(destination)
		return err
	}
	return f.Close()
}

func NewDisabledSDKInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [moniker]",
		Short: "Disabled on Mythchain; use init-node or the founder ceremony",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return errors.New("generic init is disabled to protect the official genesis; users must run init-node, founder genesis must use founder-init, local testnets may use multi-node")
		},
	}
}

func NewSafeGenesisCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis",
		Short: "Genesis inspection and validation commands",
		Args:  cobra.NoArgs,
		RunE:  client.ValidateCmd,
	}
	cmd.AddCommand(genutilcli.ValidateGenesisCmd(mbm))
	return cmd
}
