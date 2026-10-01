package keeper

import (
	"context"

	"mythprotocol/x/mythprotocol/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) RegisterTask(ctx context.Context, msg *types.MsgRegisterTask) (*types.MsgRegisterTaskResponse, error) {
	return k.Keeper.RegisterTask(ctx, msg)
}

func (k msgServer) ClaimTask(ctx context.Context, msg *types.MsgClaimTask) (*types.MsgClaimTaskResponse, error) {
	return k.Keeper.ClaimTask(ctx, msg)
}

func (k msgServer) ReleaseTask(ctx context.Context, msg *types.MsgReleaseTask) (*types.MsgReleaseTaskResponse, error) {
	return k.Keeper.ReleaseTask(ctx, msg)
}

func (k msgServer) SubmitTaskResult(ctx context.Context, msg *types.MsgSubmitTaskResult) (*types.MsgSubmitTaskResultResponse, error) {
	return k.Keeper.SubmitTaskResult(ctx, msg)
}

func (k msgServer) VoteTask(ctx context.Context, msg *types.MsgVoteTask) (*types.MsgVoteTaskResponse, error) {
	return k.Keeper.VoteTask(ctx, msg)
}
