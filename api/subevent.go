package api

import (
	"fmt"
	"strings"

	codec_types "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/figment-networks/indexing-engine/structs"

	"github.com/figment-networks/ni-cosmoslib/util"

	"github.com/figment-networks/ni-cosmoslib/api/mapper"
	"github.com/figment-networks/ni-cosmoslib/api/tendermint_mapper"
)

var defaultMapper = &mapper.Mapper{}

// AddSubEvent converts a cosmos event from the log to a Subevent type and adds it to the provided TransactionEvent struct
func AddSubEvent(tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog, ma *mapper.Mapper) (err error) {
	// TypeUrl must be in the format "/cosmos.bank.v1beta1.MsgSend"
	tPath := strings.Split(m.TypeUrl, ".")
	if len(tPath) != 4 {
		return fmt.Errorf("problem with cosmos event cosmos event %s: %w", m.TypeUrl, util.ErrUnknownMessageType)
	}
	// for mapper = nil use the default
	if ma == nil {
		ma = defaultMapper
	}

	msgType := tPath[3]
	msgRoute := tPath[1]

	var ev structs.SubsetEvent
	switch msgRoute {
	case "authz":
		switch msgType {
		case "MsgGrant":
			ev, err = ma.AuthzGrantToSub(m.Value)
		case "MsgExecResponse":
			ev, err = ma.AuthzExecResponseToSub(m.Value)
		case "MsgExec":
			ev, err = ma.AuthzExecToSub(m.Value)
		case "MsgGrantResponse":
			ev, err = ma.AuthzGrantResponseToSub(m.Value)
		case "MsgRevoke":
			ev, err = ma.AuthzMsgRevokeToSub(m.Value)
		case "MsgRevokeResponse":
			ev, err = ma.AuthzMsgRevokeResponseToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "bank":
		switch msgType {
		case "MsgSend":
			ev, err = ma.BankSendToSub(m.Value, lg)
		case "MsgMultiSend":
			ev, err = ma.BankMultisendToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "crisis":
		switch msgType {
		case "MsgVerifyInvariant":
			ev, err = ma.CrisisVerifyInvariantToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "distribution":
		switch msgType {
		case "MsgWithdrawValidatorCommission":
			ev, err = ma.DistributionWithdrawValidatorCommissionToSub(m.Value, lg)
		case "MsgSetWithdrawAddress":
			ev, err = ma.DistributionSetWithdrawAddressToSub(m.Value)
		case "MsgWithdrawDelegatorReward":
			ev, err = ma.DistributionWithdrawDelegatorRewardToSub(m.Value, lg)
		case "MsgFundCommunityPool":
			ev, err = ma.DistributionFundCommunityPoolToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "evidence":
		switch msgType {
		case "MsgSubmitEvidence":
			ev, err = ma.EvidenceSubmitEvidenceToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "feegrant":
		switch msgType {
		case "MsgGrantAllowance":
			ev, err = ma.FeegrantGrantAllowance(m.Value)
		case "MsgGrantAllowanceResponse":
			ev, err = ma.FeegrantGrantAllowanceResponse(m.Value)
		case "MsgRevokeAllowance":
			ev, err = ma.FeegrantRevokeAllowance(m.Value)
		case "MsgRevokeAllowanceResponse":
			ev, err = ma.FeegrantRevokeAllowanceResponse(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "gov":
		switch msgType {
		case "MsgDeposit":
			ev, err = ma.GovDepositToSub(m.Value, lg)
		case "MsgVote":
			ev, err = ma.GovVoteToSub(m.Value)
		case "MsgSubmitProposal":
			ev, err = ma.GovSubmitProposalToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "slashing":
		switch msgType {
		case "MsgUnjail":
			ev, err = ma.SlashingUnjailToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "vesting":
		switch msgType {
		case "MsgCreateVestingAccount":
			ev, err = ma.VestingMsgCreateVestingAccountToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	case "staking":
		switch msgType {
		case "MsgUndelegate":
			ev, err = ma.StakingUndelegateToSub(m.Value, lg)
		case "MsgEditValidator":
			ev, err = ma.StakingEditValidatorToSub(m.Value)
		case "MsgCreateValidator":
			ev, err = ma.StakingCreateValidatorToSub(m.Value)
		case "MsgDelegate":
			ev, err = ma.StakingDelegateToSub(m.Value, lg)
		case "MsgBeginRedelegate":
			ev, err = ma.StakingBeginRedelegateToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	default:
		err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
	}

	if len(ev.Type) > 0 {
		tev.Sub = append(tev.Sub, ev)
		tev.Kind = ev.Type[0]
	}
	return err
}

func AddTendermintSubEvent(tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog) (err error) {
	var ev structs.SubsetEvent
	// TypeUrl must be in the format "/tendermint.liquidity.v1beta1.MsgSwapWithinBatch"
	tPath := strings.Split(m.TypeUrl, ".")
	if len(tPath) != 4 {
		return fmt.Errorf("problem with tendermint event %s (wrong number of members): %w", m.TypeUrl, util.ErrUnknownMessageType)
	}

	msgType := tPath[3]
	msgRoute := tPath[1]

	switch msgRoute {
	case "liquidity":
		switch msgType {
		case "MsgCreatePool":
			ev, err = tendermint_mapper.TendermintCreatePool(m.Value)
		case "MsgDepositWithinBatch":
			ev, err = tendermint_mapper.TendermintDepositWithinBatch(m.Value)
		case "MsgWithdrawWithinBatch":
			ev, err = tendermint_mapper.TendermintWithdrawWithinBatch(m.Value)
		case "MsgSwapWithinBatch":
			ev, err = tendermint_mapper.TendermintSwapWithinBatch(m.Value)
		default:
			err = fmt.Errorf("problem with tendermint liquidity event %s - %s: %w", msgRoute, msgType, util.ErrUnknownMessageType)
		}
	default:
		err = fmt.Errorf("problem with tendermint liquidity event %s - %s:  %w", msgRoute, msgType, util.ErrUnknownMessageType)
	}

	if len(ev.Type) > 0 {
		tev.Sub = append(tev.Sub, ev)
		tev.Kind = ev.Type[0]
	}

	return err
}
