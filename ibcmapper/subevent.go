package ibcmapper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/figment-networks/indexing-engine/structs"
)

// AddIBCSubEvent converts an ibc event from the log to a Subevent type and adds it to the provided TransactionEvent struct
func AddIBCSubEvent(tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog) (err error) {
	var ev structs.SubsetEvent
	// TypeUrl must be in the format "/ibc.core.client.v1.MsgCreateClient"
	tPath := strings.Split(m.TypeUrl, ".")
	if len(tPath) != 5 {
		return fmt.Errorf("problem with ibc event ibc event %s: %w", m.TypeUrl, ErrUnknownMessageType)
	}

	msgType := tPath[4]
	msgRoute := tPath[2]

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
