package api

import (
	"fmt"
	"strings"

	codec_types "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/figment-networks/indexing-engine/structs"
	ibc_mapper "github.com/figment-networks/ni-cosmoslib/api/ibc_mapper"
	"github.com/figment-networks/ni-cosmoslib/api/mapper"
)

// AddSubEvent converts a cosmos event from the log to a Subevent type and adds it to the provided TransactionEvent struct
func AddSubEvent(tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog) (err error) {
	// TypeUrl must be in the format "/cosmos.bank.v1beta1.MsgSend"
	tPath := strings.Split(m.TypeUrl, ".")
	if len(tPath) != 4 {
		return fmt.Errorf("problem with cosmos event cosmos event %s: %w", m.TypeUrl, ErrUnknownMessageType)
	}

	msgType := tPath[3]
	msgRoute := tPath[1]

	var ev structs.SubsetEvent
	switch msgRoute {
	case "bank":
		switch msgType {
		case "MsgSend":
			ev, err = mapper.BankSendToSub(m.Value, lg)
		case "MsgMultiSend":
			ev, err = mapper.BankMultisendToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "crisis":
		switch msgType {
		case "MsgVerifyInvariant":
			ev, err = mapper.CrisisVerifyInvariantToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "distribution":
		switch msgType {
		case "MsgWithdrawValidatorCommission":
			ev, err = mapper.DistributionWithdrawValidatorCommissionToSub(m.Value, lg)
		case "MsgSetWithdrawAddress":
			ev, err = mapper.DistributionSetWithdrawAddressToSub(m.Value)
		case "MsgWithdrawDelegatorReward":
			ev, err = mapper.DistributionWithdrawDelegatorRewardToSub(m.Value, lg)
		case "MsgFundCommunityPool":
			ev, err = mapper.DistributionFundCommunityPoolToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "evidence":
		switch msgType {
		case "MsgSubmitEvidence":
			ev, err = mapper.EvidenceSubmitEvidenceToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "gov":
		switch msgType {
		case "MsgDeposit":
			ev, err = mapper.GovDepositToSub(m.Value, lg)
		case "MsgVote":
			ev, err = mapper.GovVoteToSub(m.Value)
		case "MsgSubmitProposal":
			ev, err = mapper.GovSubmitProposalToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "slashing":
		switch msgType {
		case "MsgUnjail":
			ev, err = mapper.SlashingUnjailToSub(m.Value)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "vesting":
		switch msgType {
		case "MsgCreateVestingAccount":
			ev, err = mapper.VestingMsgCreateVestingAccountToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "staking":
		switch msgType {
		case "MsgUndelegate":
			ev, err = mapper.StakingUndelegateToSub(m.Value, lg)
		case "MsgEditValidator":
			ev, err = mapper.StakingEditValidatorToSub(m.Value)
		case "MsgCreateValidator":
			ev, err = mapper.StakingCreateValidatorToSub(m.Value)
		case "MsgDelegate":
			ev, err = mapper.StakingDelegateToSub(m.Value, lg)
		case "MsgBeginRedelegate":
			ev, err = mapper.StakingBeginRedelegateToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	default:
		err = fmt.Errorf("problem with cosmos event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
	}

	if len(ev.Type) > 0 {
		tev.Sub = append(tev.Sub, ev)
		tev.Kind = ev.Type[0]
	}
	return err
}

// AddIBCSubEvent converts an ibc event from the log to a Subevent type and adds it to the provided TransactionEvent struct
func AddIBCSubEvent(tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog) (err error) {
	var ev structs.SubsetEvent
	// TypeUrl must be in the format "/ibc.core.client.v1.MsgCreateClient"
	tPath := strings.Split(m.TypeUrl, ".")
	if len(tPath) != 5 {
		return fmt.Errorf("problem with ibc event ibc event %s: %w", m.TypeUrl, ErrUnknownMessageType)
	}

	msgType := tPath[3]
	msgRoute := tPath[1]

	switch msgRoute {
	case "client":
		switch msgType {
		case "MsgCreateClient":
			ev, err = ibc_mapper.IBCCreateClientToSub(m.Value)
		case "MsgUpdateClient":
			ev, err = ibc_mapper.IBCUpdateClientToSub(m.Value)
		case "MsgUpgradeClient":
			ev, err = ibc_mapper.IBCUpgradeClientToSub(m.Value)
		case "MsgSubmitMisbehaviour":
			ev, err = ibc_mapper.IBCSubmitMisbehaviourToSub(m.Value)
		default:
			err = fmt.Errorf("problem with ibc event %s - %s: %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "connection":
		switch msgType {
		case "MsgConnectionOpenInit":
			ev, err = ibc_mapper.IBCConnectionOpenInitToSub(m.Value)
		case "MsgConnectionOpenConfirm":
			ev, err = ibc_mapper.IBCConnectionOpenConfirmToSub(m.Value)
		case "MsgConnectionOpenAck":
			ev, err = ibc_mapper.IBCConnectionOpenAckToSub(m.Value)
		case "MsgConnectionOpenTry":
			ev, err = ibc_mapper.IBCConnectionOpenTryToSub(m.Value)
		default:
			err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "channel":
		switch msgType {
		case "MsgChannelOpenInit":
			ev, err = ibc_mapper.IBCChannelOpenInitToSub(m.Value)
		case "MsgChannelOpenTry":
			ev, err = ibc_mapper.IBCChannelOpenTryToSub(m.Value)
		case "MsgChannelOpenConfirm":
			ev, err = ibc_mapper.IBCChannelOpenConfirmToSub(m.Value)
		case "MsgChannelOpenAck":
			ev, err = ibc_mapper.IBCChannelOpenAckToSub(m.Value)
		case "MsgChannelCloseInit":
			ev, err = ibc_mapper.IBCChannelCloseInitToSub(m.Value)
		case "MsgChannelCloseConfirm":
			ev, err = ibc_mapper.IBCChannelCloseConfirmToSub(m.Value)
		case "MsgRecvPacket":
			ev, err = ibc_mapper.IBCChannelRecvPacketToSub(m.Value)
		case "MsgTimeout":
			ev, err = ibc_mapper.IBCChannelTimeoutToSub(m.Value)
		case "MsgAcknowledgement":
			ev, err = ibc_mapper.IBCChannelAcknowledgementToSub(m.Value)

		default:
			err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	case "transfer":
		switch msgType {
		case "MsgTransfer":
			ev, err = ibc_mapper.IBCTransferToSub(m.Value)
		default:
			err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, ErrUnknownMessageType)
		}
	default:
		err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, ErrUnknownMessageType)
	}

	if len(ev.Type) > 0 {
		tev.Sub = append(tev.Sub, ev)
		tev.Kind = ev.Type[0]
	}

	return err
}
