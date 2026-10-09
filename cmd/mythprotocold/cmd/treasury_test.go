package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestAuditTreasuryGenesisSeparatesFounderAndCommunityPool(t *testing.T) {
	genesisPath := filepath.Join("..", "..", "..", "release", "genesis.json")
	genesis, err := os.ReadFile(genesisPath)
	if err != nil {
		t.Fatal(err)
	}
	report, err := AuditTreasuryGenesis(genesis)
	if err != nil {
		t.Fatalf("audit treasury genesis: %v", err)
	}
	if report.TreasuryAddress != "myth1jv65s3grqf6v6jl3dp4t6c9t9rk99cd86qepld" {
		t.Fatalf("treasury address = %s", report.TreasuryAddress)
	}
	if report.TreasuryAddress == report.FounderAddress {
		t.Fatal("treasury is not separate from founder account")
	}
	if !report.TreasuryBalance.Equal(math.NewInt(20_000_000_000_000)) {
		t.Fatalf("treasury balance = %s", report.TreasuryBalance)
	}
	if !report.FounderBalance.Equal(math.NewInt(1_000_000_000_000)) {
		t.Fatalf("founder balance = %s", report.FounderBalance)
	}
	if !report.TotalSupply.Equal(math.NewInt(21_000_000_000_000)) {
		t.Fatalf("total supply = %s", report.TotalSupply)
	}
	if !report.CommunityPool.AmountOf("umtc").Equal(math.LegacyNewDec(20_000_000_000_000)) {
		t.Fatalf("community pool = %s", report.CommunityPool)
	}
}

func TestAuditTreasuryGenesisRejectsUnexpectedAccountDistribution(t *testing.T) {
	genesisPath := filepath.Join("..", "..", "..", "release", "genesis.json")
	genesis, err := os.ReadFile(genesisPath)
	if err != nil {
		t.Fatal(err)
	}
	// Reduce only the reported total supply. The genesis should fail the audited
	// allocation invariants instead of presenting an incomplete treasury report.
	mutated := []byte(string(genesis))
	old := []byte(`"amount": "21000000000000"`)
	for i := range mutated {
		if len(mutated)-i >= len(old) && string(mutated[i:i+len(old)]) == string(old) {
			copy(mutated[i:i+len(old)], []byte(`"amount": "20000000000000"`))
			break
		}
	}
	report, err := AuditTreasuryGenesis(mutated)
	if err != nil {
		t.Fatalf("audit parser should still read valid JSON: %v", err)
	}
	if err := ValidateTreasuryAudit(report); err == nil {
		t.Fatal("invalid genesis supply passed treasury audit validation")
	}
}

func TestTreasuryCommunityPoolSpendRequiresGovernanceAuthority(t *testing.T) {
	ctrl := gomock.NewController(t)
	storeKey := storetypes.NewKVStoreKey(distrtypes.StoreKey)
	transientKey := storetypes.NewTransientStoreKey("treasury_test")
	storeService := runtime.NewKVStoreService(storeKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, transientKey)
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)
	distributionAccount := authtypes.NewEmptyModuleAccount(distrtypes.ModuleName)
	govAuthority := authtypes.NewModuleAddress("gov").String()
	accountKeeper.EXPECT().GetModuleAddress(distrtypes.ModuleName).Return(distributionAccount.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("myth")).AnyTimes()

	keeper := distrkeeper.NewKeeper(encCfg.Codec, storeService, accountKeeper, bankKeeper, stakingKeeper, authtypes.FeeCollectorName, govAuthority)
	if err := keeper.FeePool.Set(ctx, distrtypes.FeePool{CommunityPool: sdk.NewDecCoinsFromCoins(sdk.NewCoin("umtc", math.NewInt(20_000_000_000_000)))}); err != nil {
		t.Fatalf("initialize treasury pool: %v", err)
	}
	msgServer := distrkeeper.NewMsgServerImpl(keeper)
	transfer := &distrtypes.MsgCommunityPoolSpend{
		Authority: govAuthority,
		Recipient: sdk.AccAddress([]byte("grant-recipient-address")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("umtc", math.NewInt(1_000_000_000))),
	}

	// Founder-key authority is not accepted, and this failed request makes no
	// transfer call to the bank keeper.
	unauthorized := *transfer
	unauthorized.Authority = sdk.AccAddress([]byte("founder-key-address___")).String()
	if _, err := msgServer.CommunityPoolSpend(sdk.WrapSDKContext(ctx), &unauthorized); err == nil {
		t.Fatal("non-governance treasury spend was accepted")
	}

	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, distrtypes.ModuleName, gomock.Any(), transfer.Amount).Return(nil)
	if _, err := msgServer.CommunityPoolSpend(sdk.WrapSDKContext(ctx), transfer); err != nil {
		t.Fatalf("governance-authorized treasury spend: %v", err)
	}
	updatedPool, err := keeper.FeePool.Get(ctx)
	if err != nil {
		t.Fatalf("read updated treasury pool: %v", err)
	}
	if got := updatedPool.CommunityPool.AmountOf("umtc"); !got.Equal(math.LegacyNewDec(19_999_000_000_000)) {
		t.Fatalf("Community Pool after authorized transfer = %s umtc", got)
	}
}

func TestInvalidTreasuryTransactionSignatureIsRejected(t *testing.T) {
	key := secp256k1.GenPrivKey()
	payload := []byte("governance-authorized treasury spend")
	signature, err := key.Sign(payload)
	if err != nil {
		t.Fatalf("sign treasury transaction payload: %v", err)
	}
	if !key.PubKey().VerifySignature(payload, signature) {
		t.Fatal("valid transaction signature was rejected")
	}
	signature[0] ^= 0xff
	if key.PubKey().VerifySignature(payload, signature) {
		t.Fatal("invalid transaction signature was accepted")
	}
}
