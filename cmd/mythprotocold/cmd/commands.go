package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log/v2"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythprotocol/app"
)

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	basicManager module.BasicManager,
) {
	rootCmd.AddCommand(
		NewDisabledSDKInitCmd(),
		NewInitNodeCmd(app.DefaultNodeHome),
		NewInPlaceTestnetCmd(),
		NewTestnetMultiNodeCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		NewFounderInitCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		NewJoinCmd(),
		NewTreasuryAuditCmd(),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)

	server.AddCommandsWithStartCmdOptions(rootCmd, app.DefaultNodeHome, newApp, appExport, server.StartCmdOptions{
		AddFlags: addModuleInitFlags,
	})

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		NewSafeGenesisCmd(basicManager),
		server.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)
}

// addModuleInitFlags adds more flags to the start command.
func addModuleInitFlags(startCmd *cobra.Command) {
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.WaitTxCmd(),
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

// newApp creates the application
func newApp(
	logger log.Logger,
	db dbm.DB,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)
	baseappOptions = append(baseappOptions, peerIDFilterOption(appOpts))

	return app.New(
		logger, db, true,
		appOpts,
		baseappOptions...,
	)
}

// peerIDFilterOption restricts peers of an initialized joining node to the
// persistent peers and seeds explicitly configured by the operator.
func peerIDFilterOption(appOpts servertypes.AppOptions) func(*baseapp.BaseApp) {
	allowed := make(map[string]struct{})
	if home, ok := appOpts.Get(flags.FlagHome).(string); ok && home != "" {
		allowed = allowedPeerIDs(filepath.Join(home, "config", cmtcfg.DefaultConfigFileName))
	}

	return func(bApp *baseapp.BaseApp) {
		bApp.SetIDPeerFilter(func(id string) *abci.ResponseQuery {
			if _, ok := allowed[id]; ok {
				return &abci.ResponseQuery{}
			}
			return &abci.ResponseQuery{Code: 1, Log: "peer ID is not configured as a persistent peer or seed"}
		})
	}
}

func allowedPeerIDs(configPath string) map[string]struct{} {
	allowed := make(map[string]struct{})
	cfg, err := readCometConfig(configPath)
	if err != nil || cfg.P2P == nil {
		return allowed
	}
	lists := []string{cfg.P2P.PersistentPeers, cfg.P2P.Seeds, cfg.P2P.UnconditionalPeerIDs}
	for _, list := range lists {
		for _, raw := range strings.Split(list, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			id, err := parsePeerAddress(raw)
			if err == nil {
				allowed[string(id)] = struct{}{}
			}
		}
	}
	return allowed
}

// appExport creates a new app (optionally at a given height) and exports state.
func appExport(
	logger log.Logger,
	db dbm.DB,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var bApp *app.App

	// this check is necessary as we use the flag in x/upgrade.
	// we can exit more gracefully by checking the flag here.
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	appOpts = viperAppOpts
	if height != -1 {
		bApp = app.New(logger, db, false, appOpts)
		if err := bApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		bApp = app.New(logger, db, true, appOpts)
	}

	return bApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
