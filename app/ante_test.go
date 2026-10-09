package app

import (
	"testing"

	"cosmossdk.io/log/v2"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	mythtypes "mythprotocol/x/mythprotocol/types"
)

func TestAppBuildsWithGaslessAnteFeeChecker(t *testing.T) {
	db := dbm.NewMemDB()
	defer db.Close()
	options := sims.AppOptionsMap{}
	options[flags.FlagHome] = t.TempDir()

	app := New(log.NewNopLogger(), db, false, options, baseapp.SetChainID("gasless-ante-test"))
	require.NotNil(t, app)
}

func TestPoUWMessagesAreExcludedFromGasless(t *testing.T) {
	pouw := []types.Msg{
		&mythtypes.MsgRegisterTask{},
		&mythtypes.MsgClaimTask{},
		&mythtypes.MsgSubmitTaskResult{},
		&mythtypes.MsgVoteTask{},
		&mythtypes.MsgReleaseTask{},
	}
	for _, msg := range pouw {
		require.True(t, isPoUWTaskMessage(msg), "expected PoUW message to be fee-gated: %T", msg)
	}

	// Ordinary messages keep the one-trial gasless path for brand-new accounts.
	require.False(t, isPoUWTaskMessage(&banktypes.MsgSend{}))
}

func TestGasFeeDenomActivationFollowsPoUWPhase(t *testing.T) {
	if !feeDenomAllowed(false, "umtc") {
		t.Fatal("umtc fees must be allowed in the MTC-only phase")
	}
	if feeDenomAllowed(false, "umyth") || feeDenomAllowed(true, "umyth") {
		t.Fatal("legacy umyth fees must not be accepted")
	}
	if feeDenomAllowed(false, "uzyra") {
		t.Fatal("uzyra fees must be disabled before PoUW activation")
	}
	if !feeDenomAllowed(true, "uzyra") {
		t.Fatal("uzyra fees must be enabled after PoUW activation")
	}
	if feeDenomAllowed(false, "stake") || feeDenomAllowed(true, "stake") {
		t.Fatal("legacy stake fees must not be accepted")
	}
}
