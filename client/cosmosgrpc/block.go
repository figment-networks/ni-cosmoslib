package cosmosgrpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

// GetBlock fetches most recent block from chain
func (c *Client) GetBlock(ctx context.Context, height uint64) (block *types.Block, blockID *types.BlockID, err error) {

	if height == 0 {
		lb, err := c.tmServiceClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
		if err != nil {
			return nil, nil, err
		}
		return lb.Block, lb.BlockId, nil
	}

	bbh, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)})
	if err != nil {
		return nil, nil, err
	}

	return bbh.Block, bbh.BlockId, nil
}
