package cosmosgrpc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	errUnknownMessageType = fmt.Errorf("unknown message type")
)

// TxLogError Error message
type TxLogError struct {
	Codespace string  `json:"codespace"`
	Code      float64 `json:"code"`
	Message   string  `json:"message"`
}

func (c *Client) GetRawTxs(ctx context.Context, height, perPage uint64) (txs []*tx.Tx, txResponses []*types.TxResponse, err error) {
	pag := &query.PageRequest{
		CountTotal: true,
		Limit:      perPage,
	}

	var skipped uint64
	var page = uint64(1)
	for {
		pag.Offset = (perPage * page) - perPage
		now := time.Now()

		nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutSearchTxCall)
		grpcRes, err := c.txServiceClient.GetTxsEvent(nctx, &tx.GetTxsEventRequest{
			Events:     []string{"tx.height=" + strconv.FormatUint(height, 10)},
			Pagination: pag,
		}, grpc.WaitForReady(true))
		cancel()

		c.logger.Debug("[OSMOSIS-API] Request Time (GetTxsEvent)", zap.Duration("duration", time.Now().Sub(now)))
		if err != nil {
			ngrpcRes, nskipped, err := c.skipIfUnresolvable(ctx, height, pag, uint64(len(txs)), err)
			if err != nil {
				return nil, nil, err
			}
			grpcRes = ngrpcRes
			skipped = skipped + nskipped
		}

		txs = append(txs, grpcRes.Txs...)
		txResponses = append(txResponses, grpcRes.TxResponses...)

		if grpcRes.Pagination.GetTotal() <= uint64(len(txs))+skipped {
			break
		}
		page++
	}

	return txs, txResponses, nil
}

// skipIfUnresolvable fetch the tx one at a time skipping any unresolvable types.
func (c *Client) skipIfUnresolvable(ctx context.Context, height uint64, pag *query.PageRequest, currentTxCnt uint64, err error) (grpcRes *tx.GetTxsEventResponse, nskipped uint64, errOut error) {

	if !c.errorResolve.Check(err) {
		return nil, 0, err
	}

	grpcRes = &tx.GetTxsEventResponse{}

	c.logger.Error("Skipping unresolvable height", zap.Error(err), zap.Uint64("height", height))
	grpcRes.Txs = make([]*tx.Tx, 0, pag.Limit)
	grpcRes.TxResponses = make([]*types.TxResponse, 0, pag.Limit)

	// transactionTotal represents the total number of tx's known on the node.
	var transactionTotal uint64
	var transactionTotalFound bool
	// skippedTxCount represents the number of tx's we skipped b/c we couldn't parse.
	var skippedTxCount uint64

	var offset uint64
	for offset < pag.Limit {
		ngrpcRes, err := c.getOneTransaction(ctx, height, pag.Offset+offset)

		if err != nil {
			// the tx total count isn't known at least 1 tx was fetched (and was skipped because it was unparseable)
			// and we see page should be within [1, X] range, given Y
			// implies that there was only offset txs and we can stop processing.
			if offset > 0 && !transactionTotalFound &&
				strings.Contains(err.Error(), "page should be within") &&
				strings.Contains(err.Error(), "range, given") {
				transactionTotal = offset
				break
			}
			// if the tx at this height is an un-parseable message type, skip.
			if !c.errorResolve.Check(err) {
				offset++
				skippedTxCount++
				if transactionTotalFound {
					if transactionTotal <= currentTxCnt+offset {
						break
					}
				}
				continue
			}
			return grpcRes, skippedTxCount, err
		}

		grpcRes.Txs = append(grpcRes.Txs, ngrpcRes.Txs...)
		grpcRes.TxResponses = append(grpcRes.TxResponses, ngrpcRes.TxResponses...)

		transactionTotal = ngrpcRes.GetPagination().GetTotal()
		transactionTotalFound = true

		offset++
		if transactionTotal <= currentTxCnt+offset {
			break
		}
	}

	grpcRes.Pagination = &query.PageResponse{Total: transactionTotal}
	return grpcRes, skippedTxCount, nil
}

func (c *Client) getOneTransaction(ctx context.Context, height, offset uint64) (*tx.GetTxsEventResponse, error) {
	nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutSearchTxCall)
	ngrpcRes, err := c.txServiceClient.GetTxsEvent(nctx, &tx.GetTxsEventRequest{
		Events: []string{"tx.height=" + strconv.FormatUint(height, 10)},
		Pagination: &query.PageRequest{
			CountTotal: true,
			Offset:     offset,
			Limit:      1,
		},
	}, grpc.WaitForReady(true))
	cancel()
	return ngrpcRes, err
}
