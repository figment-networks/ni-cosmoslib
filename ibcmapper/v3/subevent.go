package ibcmapper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/figment-networks/indexing-engine/structs"

	codec_types "github.com/cosmos/cosmos-sdk/codec/types"

	api "github.com/figment-networks/ni-cosmoslib/api"
)

var invalidTypeErrFmt string = "Not a %s type: %w"

// AddIBCSubEvent converts an ibc event from the log to a Subevent type and adds it to the provided TransactionEvent struct
func AddIBCSubEvent(tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog) (err error) {
	var ev structs.SubsetEvent
	// TypeUrl must be in the format "/ibc.core.client.v1.MsgCreateClient"
	tPath := strings.Split(m.TypeUrl, ".")
	if len(tPath) != 5 {
		return fmt.Errorf("problem with ibc event ibc event %s: %w", m.TypeUrl, api.ErrUnknownMessageType)
	}

	msgType := tPath[4]
	msgRoute := tPath[2]

	switch msgRoute {
	case "client":
		switch msgType {
		case "MsgCreateClient":
			ev, err = IBCCreateClientToSub(m.Value)
		case "MsgUpdateClient":
			ev, err = IBCUpdateClientToSub(m.Value)
		case "MsgUpgradeClient":
			ev, err = IBCUpgradeClientToSub(m.Value)
		case "MsgSubmitMisbehaviour":
			ev, err = IBCSubmitMisbehaviourToSub(m.Value)
		default:
			err = fmt.Errorf("problem with ibc event %s - %s: %w", msgRoute, msgType, api.ErrUnknownMessageType)
		}
	case "connection":
		switch msgType {
		case "MsgConnectionOpenInit":
			ev, err = IBCConnectionOpenInitToSub(m.Value)
		case "MsgConnectionOpenConfirm":
			ev, err = IBCConnectionOpenConfirmToSub(m.Value)
		case "MsgConnectionOpenAck":
			ev, err = IBCConnectionOpenAckToSub(m.Value)
		case "MsgConnectionOpenTry":
			ev, err = IBCConnectionOpenTryToSub(m.Value)
		default:
			err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, api.ErrUnknownMessageType)
		}
	case "channel":
		// types defined in https://github.com/cosmos/ibc-go/blob/9f70a070d773f8bfdb62d4205d8878f3149f351a/modules/core/04-channel/types/tx.pb.go#L896-L915
		switch msgType {
		case "MsgChannelOpenInit":
			ev, err = IBCChannelOpenInitToSub(m.Value)
		case "MsgChannelOpenTry":
			ev, err = IBCChannelOpenTryToSub(m.Value)
		case "MsgChannelOpenConfirm":
			ev, err = IBCChannelOpenConfirmToSub(m.Value)
		case "MsgChannelOpenAck":
			ev, err = IBCChannelOpenAckToSub(m.Value)
		case "MsgChannelCloseInit":
			ev, err = IBCChannelCloseInitToSub(m.Value)
		case "MsgChannelCloseConfirm":
			ev, err = IBCChannelCloseConfirmToSub(m.Value)
		case "MsgRecvPacket":
			ev, err = IBCChannelRecvPacketToSub(m.Value)
		case "MsgTimeout":
			ev, err = IBCChannelTimeoutToSub(m.Value)
		case "MsgTimeoutOnClose":
			ev, err = IBCChannelTimeoutOnCloseToSub(m.Value)
		case "MsgAcknowledgement":
			ev, err = IBCChannelAcknowledgementToSub(m.Value)

		default:
			err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, api.ErrUnknownMessageType)
		}
	case "transfer":
		switch msgType {
		case "MsgTransfer":
			ev, err = IBCTransferToSub(m.Value)
		default:
			err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, api.ErrUnknownMessageType)
		}
	default:
		err = fmt.Errorf("problem with ibc event %s - %s:  %w", msgRoute, msgType, api.ErrUnknownMessageType)
	}

	if len(ev.Type) > 0 {
		tev.Sub = append(tev.Sub, ev)
		tev.Kind = ev.Type[0]
	}

	return err
}
