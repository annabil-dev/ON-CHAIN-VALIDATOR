package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cobra"
)

type treasuryGenesis struct {
	ChainID  string `json:"chain_id"`
	AppState struct {
		Auth struct {
			Accounts []struct {
				Type    string `json:"@type"`
				Address string `json:"address"`
			} `json:"accounts"`
		} `json:"auth"`
		Bank struct {
			Balances []struct {
				Address string     `json:"address"`
				Coins   []sdk.Coin `json:"coins"`
			} `json:"balances"`
			Supply []sdk.Coin `json:"supply"`
		} `json:"bank"`
		Distribution struct {
			FeePool struct {
				CommunityPool sdk.DecCoins `json:"community_pool"`
			} `json:"fee_pool"`
		} `json:"distribution"`
	} `json:"app_state"`
}

type TreasuryAudit struct {
	ChainID          string
	TreasuryAddress  string
	TreasuryBalance  math.Int
	CommunityPool    sdk.DecCoins
	FounderAddress   string
	FounderBalance   math.Int
	TotalSupply      math.Int
	TreasuryAuthType string
	Control          string
}

// AuditTreasuryGenesis checks that the founder allocation and protocol treasury
// are disjoint and that the treasury remains in the governance-controlled
// distribution module account/community pool.
func AuditTreasuryGenesis(data []byte) (TreasuryAudit, error) {
	var genesis treasuryGenesis
	if err := json.Unmarshal(data, &genesis); err != nil {
		return TreasuryAudit{}, fmt.Errorf("decode genesis: %w", err)
	}
	if genesis.ChainID == "" {
		return TreasuryAudit{}, errors.New("genesis is missing chain_id")
	}
	treasuryAddress := sdk.AccAddress(authtypes.NewModuleAddress(distrtypes.ModuleName)).String()
	balances := make(map[string]sdk.Coins, len(genesis.AppState.Bank.Balances))
	for _, entry := range genesis.AppState.Bank.Balances {
		balances[entry.Address] = entry.Coins
	}
	treasuryBalance := balances[treasuryAddress].AmountOf("umtc")
	founders := make([]string, 0, 1)
	for _, account := range genesis.AppState.Auth.Accounts {
		if account.Address != treasuryAddress && balances[account.Address].AmountOf("umtc").IsPositive() {
			founders = append(founders, account.Address)
		}
	}
	if len(founders) != 1 {
		return TreasuryAudit{}, fmt.Errorf("expected exactly one funded non-treasury genesis auth account, found %d", len(founders))
	}
	founderAddress := founders[0]
	founderBalance := balances[founderAddress].AmountOf("umtc")
	totalSupply := math.ZeroInt()
	for _, coin := range genesis.AppState.Bank.Supply {
		if coin.Denom == "umtc" {
			totalSupply = coin.Amount
		}
	}
	if treasuryAddress == founderAddress {
		return TreasuryAudit{}, errors.New("treasury address must be separate from founder address")
	}
	return TreasuryAudit{
		ChainID:          genesis.ChainID,
		TreasuryAddress:  treasuryAddress,
		TreasuryBalance:  treasuryBalance,
		CommunityPool:    genesis.AppState.Distribution.FeePool.CommunityPool,
		FounderAddress:   founderAddress,
		FounderBalance:   founderBalance,
		TotalSupply:      totalSupply,
		TreasuryAuthType: "distribution module account (not founder BaseAccount)",
		Control:          "x/distribution module; MsgCommunityPoolSpend authority defaults to x/gov",
	}, nil
}

func ValidateTreasuryAudit(report TreasuryAudit) error {
	if report.TreasuryAddress == report.FounderAddress {
		return errors.New("treasury address must be separate from founder address")
	}
	if !report.TreasuryBalance.Equal(math.NewInt(20_000_000_000_000)) {
		return fmt.Errorf("treasury balance is %s umtc; expected 20000000000000", report.TreasuryBalance)
	}
	if !report.FounderBalance.Equal(math.NewInt(1_000_000_000_000)) {
		return fmt.Errorf("founder balance is %s umtc; expected 1000000000000", report.FounderBalance)
	}
	if !report.TotalSupply.Equal(math.NewInt(21_000_000_000_000)) {
		return fmt.Errorf("total supply is %s umtc; expected 21000000000000", report.TotalSupply)
	}
	if !report.CommunityPool.AmountOf("umtc").Equal(math.LegacyNewDec(20_000_000_000_000)) {
		return fmt.Errorf("Community Pool contains %s; expected 20000000000000 umtc", report.CommunityPool)
	}
	return nil
}

func NewTreasuryAuditCmd() *cobra.Command {
	var genesisPath string
	cmd := &cobra.Command{
		Use:   "treasury-audit",
		Short: "Audit founder and protocol treasury allocation in a genesis file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := genesisPath
			if path == "" {
				resolved, err := resolveGenesisPath("")
				if err != nil {
					return err
				}
				path = resolved
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			report, err := AuditTreasuryGenesis(data)
			if err != nil {
				return err
			}
			if err := ValidateTreasuryAudit(report); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Chain ID: %s\nTreasury address: %s\nTreasury type: %s\nTreasury balance: %s umtc\nCommunity Pool: %s\nFounder address: %s\nFounder balance: %s umtc\nTotal supply: %s umtc\nControl: %s\n", report.ChainID, report.TreasuryAddress, report.TreasuryAuthType, report.TreasuryBalance, report.CommunityPool, report.FounderAddress, report.FounderBalance, report.TotalSupply, report.Control)
			return nil
		},
	}
	cmd.Flags().StringVar(&genesisPath, "genesis", "", "path to the genesis.json to audit")
	return cmd
}
