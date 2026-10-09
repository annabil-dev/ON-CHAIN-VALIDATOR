package cmd

import (
	"strings"
	"testing"
)

func TestFounderInitRequiresConfirmedProductionChainID(t *testing.T) {
	for _, tc := range []struct {
		name, chainID, confirm string
		wantErr                bool
	}{
		{name: "accepted explicit production ID", chainID: "mythchain-mainnet-1", confirm: "mythchain-mainnet-1"},
		{name: "missing confirmation", chainID: "mythchain-mainnet-1", wantErr: true},
		{name: "confirmation mismatch", chainID: "mythchain-mainnet-1", confirm: "mythchain-mainnet-2", wantErr: true},
		{name: "existing random test chain", chainID: "chain-rlkh6n", confirm: "chain-rlkh6n", wantErr: true},
		{name: "local sync test chain", chainID: "phase2-local-sync", confirm: "phase2-local-sync", wantErr: true},
		{name: "invalid characters", chainID: "myth chain", confirm: "myth chain", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFounderChainID(tc.chainID, tc.confirm)
			if tc.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestGenericSDKInitIsDisabled(t *testing.T) {
	cmd := NewDisabledSDKInitCmd()
	cmd.SetArgs([]string{"testnode"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "generic init is disabled") {
		t.Fatalf("generic init error = %v, want explicit disabled message", err)
	}
}

func TestFounderEndpointRejectsPlaceholderPort(t *testing.T) {
	for _, address := range []string{"TEMP-NGROK-HOST:TEMP-NGROK-PORT", "0.tcp.ap.ngrok.io:0", "0.tcp.ap.ngrok.io:65536", "0.tcp.ap.ngrok.io:"} {
		if _, err := validateFounderEndpoint(address); err == nil {
			t.Fatalf("accepted invalid founder P2P endpoint %q", address)
		}
	}
	host, err := validateFounderEndpoint("0.tcp.ap.ngrok.io:29687")
	if err != nil || host != "0.tcp.ap.ngrok.io" {
		t.Fatalf("valid ngrok endpoint rejected: host=%q err=%v", host, err)
	}
}

func TestPublicGenesisCommandExposesValidationOnly(t *testing.T) {
	genesisCmd := NewSafeGenesisCmd(nil)
	if genesisCmd.Commands() == nil || len(genesisCmd.Commands()) != 1 || genesisCmd.Commands()[0].Name() != "validate" {
		t.Fatalf("public genesis subcommands = %v, want validate-genesis only", genesisCmd.Commands())
	}
	for _, alias := range genesisCmd.Commands()[0].Aliases {
		if alias == "validate-genesis" {
			return
		}
	}
	t.Fatal("genesis validation alias is not registered")
}
