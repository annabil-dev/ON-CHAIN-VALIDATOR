package cmd

import (
	"github.com/spf13/cobra"

	"mythprotocol/app"
)

// NewJoinCmd creates the normal node onboarding flow.
// Joining nodes consume the official genesis instead of generating one.
func NewJoinCmd() *cobra.Command {
	return NewJoinNetworkCmd(app.DefaultNodeHome)
}
