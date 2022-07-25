package rewards

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	ttypes "github.com/tendermint/tendermint/proto/tendermint/types"
	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/figment-networks/indexing-engine/proto/rewstruct"
	"github.com/figment-networks/indexing-engine/structs"

	pb "github.com/figment-networks/indexing-engine/proto/datastore"

	"github.com/figment-networks/ni-cosmoslib/client/cosmosgrpc"

	"go.uber.org/zap"
)

var (
	ErrNoRows = errors.New("No records")
)

type RewardsNonce struct {
	Height   uint64
	Sequence uint64
}

type Crossing struct {
	Height         uint64
	Sequence       uint64
	PreviousHeight uint64
	BlockTimes     map[uint64]time.Time
}

func (c *Crossing) GetHeight() uint64 {
	return c.Height
}

func (c *Crossing) GetSequence() uint64 {
	return c.Sequence
}

type HeightTime struct {
	Height uint64
	Time   time.Time
}

type HeightError struct {
	Height uint64
	Error  error
}

type DelegatorOP int64

const (
	DelegatorOPAdd DelegatorOP = iota
	DelegatorOPRemove
	DelegatorOPBoth
)

type DelegatorValidator struct {
	Op        DelegatorOP
	Delegator string
	Validator string
	Amounts   []*rewstruct.Amount
}

type RewardProducer interface {
	GetRewards(*rewstruct.RewardTx) []structs.ClaimedReward
	GetDelegations(tx *rewstruct.RewardTx) (accounts []DelegatorValidator)
	MapTransactions(txs []*tx.Tx, txResponses []*types.TxResponse, t time.Time) (retTxs *rewstruct.RewardTxs, err error)
	PostMsgBeginRedelegate(tx *rewstruct.RewardTx, dels []cosmosgrpc.Delegators) (*rewstruct.RewardTx, error)
}

type Client interface {
	GetBlock(ctx context.Context, height uint64) (block *ttypes.Block, blockID *ttypes.BlockID, err error)
	GetRawTxs(ctx context.Context, height uint64, perPage uint64) (txs []*tx.Tx, txResponses []*types.TxResponse, err error)

	GetHeightValidators(ctx context.Context, height, limit, page uint64) (vals []cosmosgrpc.Validator, err error)
	GetDelegators(ctx context.Context, height uint64, operatorAddress string, limit, page uint64) (vals []cosmosgrpc.DelegationResponse, err error)
	GetDelegatorDelegations(ctx context.Context, height uint64, delegatorAddress string, limit, page uint64) (vals []cosmosgrpc.DelegationResponse, err error)
	GetDelegations(ctx context.Context, height uint64, delegatorAddress string) (dels []cosmosgrpc.Delegators, err error)
}

type RewardsExtractionConfig struct {
	ValidatorFetchPage uint64
	DelegatorFetchPage uint64
	DatastorePrefix    string
	ChainID            string
	Network            string
	MaxChainHeight     uint64
}

type RewardsExtraction struct {
	logger *zap.Logger
	Cfg    RewardsExtractionConfig

	client   Client
	dsClient pb.DatastoreServiceClient

	orp RewardProducer
}

func NewRewardsExtraction(logger *zap.Logger, cfg RewardsExtractionConfig, client Client, dsClient pb.DatastoreServiceClient, orp RewardProducer) *RewardsExtraction {
	return &RewardsExtraction{
		client:   client,
		dsClient: dsClient,
		orp:      orp,
		logger:   logger,
		Cfg:      cfg,
	}
}

type RewardAmountView struct {
	// Textual representation of Amount
	Text string `json:"text,omitempty"`
	// The currency in what amount is returned (if applies)
	Currency string `json:"currency,omitempty"`

	// Numeric part of the amount
	Numeric *big.Int `json:"numeric,omitempty"`
	// Exponential part of amount obviously 0 by default
	Exp int32 `json:"exp,omitempty"`
}

type ClaimedRewardView struct {
	// Delegator address
	Account string `json:"account,omitempty"`
	// Alternate account addresses for rewards recipients
	RewardRecipients []string `json:"reward_recipient,omitempty"`
	// Reward amounts
	Amounts []RewardAmountView `json:"amounts,omitempty"`
	// Sequence, unix hour of the block time
	Sequence uint64 `json:"sequence,omitempty"`
	// Reward height
	Height uint64 `json:"height,omitempty"`
	// Address of validator
	Validator string `json:"validator,omitempty"`
	// Block time
	Time time.Time `json:"time,omitempty"`
	// Type specifies the "type" of reward.
	Type string `json:"type,omitempty"`
	// Transaction Hash
	TxHash string `json:"txhash,omitempty"`
}

type UnclaimedRewardView struct {
	// Delegator address
	Account string `json:"account,omitempty"`
	// Rewards amounts
	Amounts []RewardAmountView `json:"amounts,omitempty"`
	// Sequence, unix hour of the block time
	Sequence uint64 `json:"sequence,omitempty"`
	// Reward Height
	Height uint64 `json:"height,omitempty"`
	// Address of validator
	Validator string `json:"validator,omitempty"`
	// Block time
	Time time.Time `json:"time,omitempty"`
}

func (re *RewardsExtraction) FetchHeights(ctx context.Context, startHeight, endHeight, sequence uint64) (h structs.Heights, crossingHeights []Crossing, err error) {
	blockTimes := map[uint64]time.Time{}
	previousHeight := startHeight
	for height := startHeight; height < endHeight+1; height++ {
		block, _, err := re.client.GetBlock(ctx, height)
		if err != nil {
			h.ErrorAt = append(h.ErrorAt, height)
			return h, crossingHeights, err
		}

		blockTimes[height] = block.Header.Time

		timeID := uint64(math.Floor(float64(block.Header.Time.Truncate(time.Hour).Unix()) / 3600))
		if sequence != timeID {
			crossingHeights = append(crossingHeights, Crossing{
				Height:         height,
				Sequence:       timeID,
				PreviousHeight: previousHeight,
				BlockTimes:     blockTimes,
			})
			sequence = timeID
			previousHeight = height
			blockTimes = map[uint64]time.Time{}
		}

		h.Heights = append(h.Heights, height)
		if h.LatestData.LastHeight < height {
			h.LatestData.LastHeight = height
			h.LatestData.LastMark = height
		}

		fmt.Println("height: ", height)
	}

	for _, ch := range crossingHeights {
		fmt.Printf("crossing heights: %v\n", ch)
	}

	return h, crossingHeights, nil
}

func (re *RewardsExtraction) CalculateRewards(ctx context.Context, crossingHeight Crossing) error {
	accs, err := re.fetchAccounts(ctx, crossingHeight.Height)
	if err != nil {
		return err
	}

	unclaimedRewards, err := re.fetchUnclaimedBalances(ctx, crossingHeight, accs)
	if err != nil {
		return err
	}

	fmt.Println(len(unclaimedRewards))
	// marshal and send to s3

	if crossingHeight.Height > crossingHeight.PreviousHeight {
		claimedRewards, err := re.fetchClaims(ctx, re.orp, crossingHeight)
		if err != nil {
			return err
		}

		fmt.Println(len(claimedRewards))
		// marshal and send to s3
	}

	return nil
}

func (re *RewardsExtraction) fetchClaims(ctx context.Context, rp RewardProducer, crossingHeight Crossing) ([]ClaimedRewardView, error) {
	const concurrentRequests = 24
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentRequests)

	claimedRewardsMutex := sync.Mutex{}
	claimedRewards := []ClaimedRewardView{}
	for height := crossingHeight.PreviousHeight + 1; height <= crossingHeight.Height; height++ {
		height := height
		g.Go(func() error {
			// exit early if any other goroutine has an error
			if err := ctx.Err(); err != nil {
				return ctx.Err()
			}

			rawTxs, rewTxResps, err := re.client.GetRawTxs(ctx, height, 100)
			if err != nil {
				return fmt.Errorf("error getting raw tx (%d): %w ", height, err)
			}
			blockTime, ok := crossingHeight.BlockTimes[height]
			if !ok {
				return fmt.Errorf("error getting block time for tx  (%d)", height)
			}
			txs, err := re.orp.MapTransactions(rawTxs, rewTxResps, blockTime)
			if err != nil {
				return fmt.Errorf("error mapping transaction  (%d): %w ", height, err)
			}

			for i, tx := range txs.Txs {
				if tx.Type == "MsgBeginRedelegate" {
					dels, err := re.client.GetDelegations(ctx, height-1, tx.Delegator)
					if err != nil {
						return fmt.Errorf("error mapping transaction getdelegations (%d): %w ", height, err)
					}
					updatedTx, err := re.orp.PostMsgBeginRedelegate(tx, dels)
					if err != nil {
						return fmt.Errorf("error mapping transaction postmsgbeginredelegate (%d): %w ", height, err)
					}
					txs.Txs[i] = updatedTx
				}

				for _, r := range rp.GetRewards(tx) {
					rewardAmounts := []RewardAmountView{}
					for _, cr := range r.ClaimedReward {
						rewardAmounts = append(rewardAmounts, RewardAmountView{
							Text:     cr.Text,
							Currency: cr.Currency,
							Numeric:  cr.Numeric,
							Exp:      cr.Exp,
						})
					}
					c := ClaimedRewardView{
						Account:          r.Account,
						RewardRecipients: r.RewardRecipients,
						Amounts:          rewardAmounts,
						Sequence:         crossingHeight.Sequence,
						Height:           height,
						Validator:        r.Validator,
						Time:             blockTime,
						Type:             r.Type,
						TxHash:           r.TxHash,
					}
					claimedRewardsMutex.Lock()
					claimedRewards = append(claimedRewards, c)
					claimedRewardsMutex.Unlock()
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return claimedRewards, nil
}

func (re *RewardsExtraction) fetchUnclaimedBalances(ctx context.Context, crossingHeight Crossing, accounts map[string]interface{}) ([]UnclaimedRewardView, error) {
	const (
		errorThreshold     = 5
		concurrentRequests = 24
	)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentRequests)

	blockTime, ok := crossingHeight.BlockTimes[crossingHeight.Height]
	if !ok {
		return nil, fmt.Errorf("error getting block time for tx  (%d)", crossingHeight.Height)
	}

	unclaimedRewardsMutex := sync.Mutex{}
	unclaimedRewards := []UnclaimedRewardView{}
	for acc, _ := range accounts {
		acc := acc
		g.Go(func() error {
			if err := ctx.Err(); err != nil {
				return ctx.Err()
			}
			consecutiveErrors := 0
			for {
				dels, err := re.client.GetDelegations(ctx, crossingHeight.Height, acc)
				if err != nil {
					consecutiveErrors++
					if consecutiveErrors < errorThreshold {
						<-time.After(1 * time.Second)
						continue
					}
					return err
				}

				for _, del := range dels {
					for _, d := range del.Unclaimed {
						ra := []RewardAmountView{}
						for _, a := range d.Unclaimed {
							ra = append(ra, RewardAmountView{
								Text:     a.Text,
								Currency: a.Currency,
								Numeric:  a.Numeric,
								Exp:      a.Exp,
							})
						}
						u := UnclaimedRewardView{
							Account:   acc,
							Amounts:   ra,
							Sequence:  crossingHeight.Sequence,
							Height:    crossingHeight.Height,
							Validator: d.ValidatorAddress,
							Time:      blockTime,
						}
						unclaimedRewardsMutex.Lock()
						unclaimedRewards = append(unclaimedRewards, u)
						unclaimedRewardsMutex.Unlock()

						fmt.Printf("unclaimed balance: %+v\n", u)
					}
				}

				break
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return unclaimedRewards, nil
}

type AccountsHeight struct {
	Sequence uint64
	Height   uint64
	Accounts map[string]interface{}
}

func (re *RewardsExtraction) fetchAccounts(ctx context.Context, height uint64) (accounts map[string]interface{}, err error) {
	validators, err := re.client.GetHeightValidators(ctx, height, 0, re.Cfg.ValidatorFetchPage)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator lists %w", err)
	}

	const concurrentRequests = 50
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentRequests)

	accountsMap := make(map[string]interface{})
	accountsMutex := sync.Mutex{}
	// produce account addresses with bounded concurrency specified by concurrentRequests
	for _, v := range validators {
		v := v // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			// exit early if any other goroutine has an error
			if err := ctx.Err(); err != nil {
				return ctx.Err()
			}
			deleg, err := re.client.GetDelegators(ctx, height, v.OperatorAddress, 0, re.Cfg.DelegatorFetchPage)
			if err != nil {
				// GetHeightValidators can produce Delegators not at this height
				// this is ok b/c we are just trying to fetch an initial list at this height.
				if strings.Contains(err.Error(), "validator does not exist") {
					return nil
				}
				return fmt.Errorf("Error getting delegators %w", err)
			}
			accountsMutex.Lock()
			defer accountsMutex.Unlock()
			for _, d := range deleg {
				accountsMap[d.Delegation.DelegatorAddress] = struct{}{}

			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}

	return accountsMap, nil
}
