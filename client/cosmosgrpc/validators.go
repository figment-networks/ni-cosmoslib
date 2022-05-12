package cosmosgrpc

import (
	"context"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const errorThreshold = 5

func (c *Client) GetHeightValidators(ctx context.Context, height, limit, page uint64) (vals []Validator, err error) {
	var (
		consecutiveErrors uint64
		total             uint64
	)
	pagination := &query.PageRequest{Limit: page}

	for {
		vs, err := c.stakingClient.Validators(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(height, 10)),
			&stakingTypes.QueryValidatorsRequest{
				//Status:     stakingTypes.Bonded.String(),
				Pagination: pagination,
			})

		if err != nil {
			consecutiveErrors++
			if consecutiveErrors < errorThreshold {
				<-time.After(1 * time.Second)
				continue
			}
			return vals, err
		}
		total += uint64(len(vs.Validators))
		consecutiveErrors = 0
		for _, val := range vs.Validators {
			or, err := c.distributionClient.ValidatorOutstandingRewards(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(height, 10)),
				&distributionTypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: val.OperatorAddress})
			if err != nil {
				return nil, err
			}

			v := Validator{
				OperatorAddress: val.OperatorAddress,
				Jailed:          val.Jailed,
				Status:          stakingTypes.BondStatus_name[int32(val.Status)],
				Tokens:          val.Tokens.BigInt(),
				DelegatorShares: val.DelegatorShares.BigInt(),
				Description: ValidatorDescription{
					Moniker:         val.Description.Moniker,
					Identity:        val.Description.Identity,
					Website:         val.Description.Website,
					SecurityContact: val.Description.SecurityContact,
					Details:         val.Description.Details,
				},
				UnbondingHeight: val.UnbondingHeight,
				UnbondingTime:   val.UnbondingTime,
				Commission: Commission{
					Rate:          val.Commission.Rate.BigInt(),
					MaxRate:       val.Commission.MaxRate.BigInt(),
					MaxChangeRate: val.Commission.MaxChangeRate.BigInt(),
					UpdateTime:    val.Commission.UpdateTime,
				},
				MinSelfDelegation: val.MinSelfDelegation.BigInt(),
			}

			for _, rew := range or.Rewards.Rewards {
				v.Rewards = append(v.Rewards,
					TransactionAmount{
						Text:     rew.Amount.String(),
						Numeric:  rew.Amount.BigInt(), // probably wrong
						Currency: rew.Denom,
					})
			}
			vals = append(vals, v)
		}

		if vs.Pagination.NextKey == nil {
			return vals, err
		}
		pagination.Key = vs.Pagination.NextKey

		if limit > 0 && total >= limit {
			return vals, err
		}
	}
}

func (c *Client) GetDelegators(ctx context.Context, height uint64, operatorAddress string, limit, page uint64) (vals []DelegationResponse, err error) {
	var (
		consecutiveErrors uint64
		total             uint64
	)
	pagination := &query.PageRequest{Limit: page}
	for {
		vd, err := c.stakingClient.ValidatorDelegations(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(height, 10)),
			&stakingTypes.QueryValidatorDelegationsRequest{ValidatorAddr: operatorAddress, Pagination: pagination})
		if err != nil {
			// this is a legit error, no need to retry.
			if strings.Contains(err.Error(), "validator does not exist") {
				return vals, err
			}
			consecutiveErrors++
			if consecutiveErrors < errorThreshold {
				<-time.After(1 * time.Second)
				continue
			}
			return vals, err
		}
		consecutiveErrors = 0
		total += uint64(len(vd.DelegationResponses))
		for _, dr := range vd.DelegationResponses {
			vals = append(vals,
				DelegationResponse{Delegation: Delegation{
					DelegatorAddress: dr.Delegation.DelegatorAddress,
					ValidatorAddress: dr.Delegation.ValidatorAddress,
					Shares:           dr.Delegation.Shares.BigInt(),
				}, Balance: TransactionAmount{
					Text:     dr.Balance.Amount.String(),
					Numeric:  dr.Balance.Amount.BigInt(),
					Currency: dr.Balance.Denom,
				}},
			)
		}

		if vd.Pagination.NextKey == nil {
			return vals, err
		}
		pagination.Key = vd.Pagination.NextKey

		if limit > 0 && total >= limit {
			return vals, err
		}
	}
}

func (c *Client) GetDelegatorDelegations(ctx context.Context, height uint64, delegatorAddress string, limit, page uint64) (vals []DelegationResponse, err error) {
	var (
		consecutiveErrors uint64
		total             uint64
	)
	pagination := &query.PageRequest{Limit: page}
	for {
		vd, err := c.stakingClient.DelegatorDelegations(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(height, 10)),
			&stakingTypes.QueryDelegatorDelegationsRequest{DelegatorAddr: delegatorAddress, Pagination: pagination})
		if err != nil {
			consecutiveErrors++
			if consecutiveErrors < errorThreshold {
				<-time.After(1 * time.Second)
				continue
			}
			return vals, err
		}
		consecutiveErrors = 0
		total += uint64(len(vd.DelegationResponses))
		for _, dr := range vd.DelegationResponses {
			vals = append(vals,
				DelegationResponse{Delegation: Delegation{
					DelegatorAddress: dr.Delegation.DelegatorAddress,
					ValidatorAddress: dr.Delegation.ValidatorAddress,
					Shares:           dr.Delegation.Shares.BigInt(),
				}, Balance: TransactionAmount{
					Text:     dr.Balance.Amount.String(),
					Numeric:  dr.Balance.Amount.BigInt(),
					Currency: dr.Balance.Denom,
				}},
			)
		}

		if vd.Pagination.NextKey == nil {
			return vals, err
		}
		pagination.Key = vd.Pagination.NextKey

		if limit > 0 && total >= limit {
			return vals, err
		}
	}
}

func (c *Client) GetDelegations(ctx context.Context, height uint64, delegatorAddress string) (dels []Delegators, err error) {
	rew, err := c.distributionClient.DelegationTotalRewards(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(height, 10)),
		&distributionTypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: delegatorAddress})
	if err != nil {
		return nil, err
	}
	for _, r := range rew.Rewards {
		unc := []TransactionAmount{}

		for _, c := range r.Reward {
			unc = append(unc, TransactionAmount{
				Text:     c.Amount.String(),
				Numeric:  c.Amount.BigInt(),
				Currency: c.Denom,
			})
		}
		dels = append(dels, Delegators{
			DelegatorAddress: delegatorAddress,
			Unclaimed: []DelegatorsUnclaimed{
				{
					ValidatorAddress: r.ValidatorAddress,
					Unclaimed:        unc,
				},
			},
		})
	}
	return dels, nil
}
