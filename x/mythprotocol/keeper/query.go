package keeper

import (
	"context"

	"mythprotocol/x/mythprotocol/types"
)

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

func (q queryServer) TaskLease(ctx context.Context, req *types.QueryTaskLeaseRequest) (*types.QueryTaskLeaseResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidTaskID
	}
	lease, found, err := q.k.GetTaskLease(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}
	return &types.QueryTaskLeaseResponse{Lease: lease, Found: found}, nil
}
