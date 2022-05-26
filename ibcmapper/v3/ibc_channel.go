package ibcmapper

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/figment-networks/indexing-engine/structs"
	shared "github.com/figment-networks/indexing-engine/structs"
	"github.com/figment-networks/ni-cosmoslib/api/util"

	channel "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/gogo/protobuf/proto"
)

var (
	// the constant is missing from https://github.com/cosmos/ibc-go/blob/main/modules/core/04-channel/types/events.go
	// and tx.go is missing in https://github.com/cosmos/ibc-go/tree/main/modules/core/04-channel/types
	// so manually use the name defined in tx.pb.go
	constChannelTimeoutOnClose = "proof_close"
	bigZero                    = new(big.Int).SetInt64(0)
)

// IBCChannelOpenInitToSub transforms ibc.MsgChannelOpenInit sdk messages to SubsetEvent
func IBCChannelOpenInitToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenInit{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_init type: %w", err)
	}

	return shared.SubsetEvent{
		Type:   []string{"channel_open_init"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"port_id":                         {m.PortId},
			"channel_state":                   {strconv.FormatInt(int64(m.Channel.State), 10)},
			"channel_ordering":                {strconv.FormatInt(int64(m.Channel.Ordering), 10)},
			"channel_counterparty_port_id":    {m.Channel.Counterparty.PortId},
			"channel_counterparty_channel_id": {m.Channel.Counterparty.ChannelId},
			"channel_connection_hops":         m.Channel.ConnectionHops,
			"channel_version":                 {m.Channel.Version},
		},
	}, nil
}

// IBCChannelOpenConfirmToSub transforms ibc.MsgChannelOpenConfirm sdk messages to SubsetEvent
func IBCChannelOpenConfirmToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenConfirm{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_confirm type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofAck, err := util.EncodeToB64(m.ProofAck, "proof_ack")
	if err != nil {
		return se, err
	}

	return shared.SubsetEvent{
		Type:   []string{"channel_open_confirm"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"port_id":                      {m.PortId},
			"channel_id":                   {m.ChannelId},
			"proof_ack":                    {proofAck},
			"proof_height_revision_number": {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height": {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
		},
	}, nil
}

// IBCChannelOpenAckToSub transforms ibc.MsgChannelOpenAck sdk messages to SubsetEvent
func IBCChannelOpenAckToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenAck{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_ack type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofTry, err := util.EncodeToB64(m.ProofTry, "proof_try")
	if err != nil {
		return se, err
	}

	return shared.SubsetEvent{
		Type:   []string{"channel_open_ack"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"port_id":                      {m.PortId},
			"channel_id":                   {m.ChannelId},
			"counterparty_channel_id":      {m.CounterpartyChannelId},
			"counterparty_version":         {m.CounterpartyVersion},
			"proof_try":                    {proofTry},
			"proof_height_revision_number": {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height": {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
		},
	}, nil
}

// IBCChannelOpenTryToSub transforms ibc.MsgChannelOpenTry sdk messages to SubsetEvent
func IBCChannelOpenTryToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenTry{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_try type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofInit, err := util.EncodeToB64(m.ProofInit, "proof_init")
	if err != nil {
		return se, err
	}

	return shared.SubsetEvent{
		Type:   []string{"channel_open_try"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"port_id":                         {m.PortId},
			"previous_channel_id":             {m.PreviousChannelId},
			"channel_state":                   {strconv.FormatInt(int64(m.Channel.State), 10)},
			"channel_ordering":                {strconv.FormatInt(int64(m.Channel.Ordering), 10)},
			"channel_counterparty_port_id":    {m.Channel.Counterparty.PortId},
			"channel_counterparty_channel_id": {m.Channel.Counterparty.ChannelId},
			"channel_connection_hops":         m.Channel.ConnectionHops,
			"channel_version":                 {m.Channel.Version},
			"counterparty_version":            {m.CounterpartyVersion},
			"proof_init":                      {proofInit},
			"proof_height_revision_number":    {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height":    {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
		},
	}, nil
}

// IBCChannelCloseInitToSub transforms ibc.MsgChannelCloseInit sdk messages to SubsetEvent
func IBCChannelCloseInitToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgChannelCloseInit{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_close_init type: %w", err)
	}

	return shared.SubsetEvent{
		Type:   []string{"channel_close_init"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"port_id":    {m.PortId},
			"channel_id": {m.ChannelId},
		},
	}, nil
}

// IBCChannelCloseConfirmToSub transforms ibc.MsgChannelCloseConfirm sdk messages to SubsetEvent
func IBCChannelCloseConfirmToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgChannelCloseConfirm{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_close_confirm type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofInit, err := util.EncodeToB64(m.ProofInit, "proof_init")
	if err != nil {
		return se, err
	}

	return shared.SubsetEvent{
		Type:   []string{"channel_close_confirm"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"port_id":                      {m.PortId},
			"channel_id":                   {m.ChannelId},
			"proof_init":                   {proofInit},
			"proof_height_revision_number": {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height": {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
		},
	}, nil
}

// IBCChannelRecvPacketToSub transforms ibc.MsgRecvPacket sdk messages to SubsetEvent
func IBCChannelRecvPacketToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgRecvPacket{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a recv_packet type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofCommitment, err := util.EncodeToB64(m.ProofCommitment, "proof_commitment")
	if err != nil {
		return se, err
	}

	event := shared.SubsetEvent{
		Type:   []string{"recv_packet"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"packet_sequence":                       {strconv.FormatUint(m.Packet.Sequence, 10)},
			"packet_source_port":                    {m.Packet.SourcePort},
			"packet_source_channel":                 {m.Packet.SourceChannel},
			"packet_destination_port":               {m.Packet.DestinationPort},
			"packet_destination_channel":            {m.Packet.DestinationChannel},
			"packet_data":                           {string(m.Packet.Data)},
			"packet_timeout_height_revision_number": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionNumber, 10)},
			"packet_timeout_height_revision_height": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionHeight, 10)},
			"packet_timeout_stamp":                  {strconv.FormatUint(m.Packet.TimeoutTimestamp, 10)},
			"proof_commitment":                      {proofCommitment},
			"proof_height_revision_number":          {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height":          {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
		},
	}
	err = ParsePacket(m.Packet.Data, &event)
	return event, err
}

// IBCChannelTimeoutToSub transforms ibc.MsgTimeout sdk messages to SubsetEvent
func IBCChannelTimeoutToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgTimeout{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a timeout type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofUnreceived, err := util.EncodeToB64(m.ProofUnreceived, "proof_unreceived")
	if err != nil {
		return se, err
	}

	event := shared.SubsetEvent{
		Type:   []string{"timeout"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"packet_sequence":                       {strconv.FormatUint(m.Packet.Sequence, 10)},
			"packet_source_port":                    {m.Packet.SourcePort},
			"packet_source_channel":                 {m.Packet.SourceChannel},
			"packet_destination_port":               {m.Packet.DestinationPort},
			"packet_destination_channel":            {m.Packet.DestinationChannel},
			"packet_data":                           {string(m.Packet.Data)},
			"packet_timeout_height_revision_number": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionNumber, 10)},
			"packet_timeout_height_revision_height": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionHeight, 10)},
			"packet_timeout_stamp":                  {strconv.FormatUint(m.Packet.TimeoutTimestamp, 10)},
			"proof_unreceived":                      {proofUnreceived},
			"proof_height_revision_number":          {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height":          {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
			"next_sequence_recv":                    {strconv.FormatUint(m.NextSequenceRecv, 10)},
		},
	}
	err = ParsePacket(m.Packet.Data, &event)
	return event, err
}

// IBCChannelTimeoutOnCloseToSub transforms ibc.MsgTimeout sdk messages to SubsetEvent
func IBCChannelTimeoutOnCloseToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgTimeoutOnClose{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf(invalidTypeErrFmt, constChannelTimeoutOnClose, err)
	}

	// Encode fields that can contain null bytes.
	proofUnreceived, err := util.EncodeToB64(m.ProofUnreceived, "proof_unreceived")
	if err != nil {
		return se, err
	}

	proofClose, err := util.EncodeToB64(m.ProofClose, "proof_close")
	if err != nil {
		return se, err
	}

	event := shared.SubsetEvent{
		Type:   []string{constChannelTimeoutOnClose},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"packet_sequence":                       {strconv.FormatUint(m.Packet.Sequence, 10)},
			"packet_source_port":                    {m.Packet.SourcePort},
			"packet_source_channel":                 {m.Packet.SourceChannel},
			"packet_destination_port":               {m.Packet.DestinationPort},
			"packet_destination_channel":            {m.Packet.DestinationChannel},
			"packet_data":                           {string(m.Packet.Data)},
			"packet_timeout_height_revision_number": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionNumber, 10)},
			"packet_timeout_height_revision_height": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionHeight, 10)},
			"packet_timeout_stamp":                  {strconv.FormatUint(m.Packet.TimeoutTimestamp, 10)},
			"proof_unreceived":                      {proofUnreceived},
			"proof_close":                           {proofClose},
			"proof_height_revision_number":          {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height":          {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
			"next_sequence_recv":                    {strconv.FormatUint(m.NextSequenceRecv, 10)},
		},
	}
	err = ParsePacket(m.Packet.Data, &event)
	return event, err
}

// IBCChannelAcknowledgementToSub transforms ibc.MsgAcknowledgement sdk messages to SubsetEvent
func IBCChannelAcknowledgementToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &channel.MsgAcknowledgement{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_acknowledgement type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofAcked, err := util.EncodeToB64(m.ProofAcked, "proof_acked")
	if err != nil {
		return se, err
	}

	event := shared.SubsetEvent{
		Type:   []string{"channel_acknowledgement"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"packet_sequence":                       {strconv.FormatUint(m.Packet.Sequence, 10)},
			"packet_source_port":                    {m.Packet.SourcePort},
			"packet_source_channel":                 {m.Packet.SourceChannel},
			"packet_destination_port":               {m.Packet.DestinationPort},
			"packet_destination_channel":            {m.Packet.DestinationChannel},
			"packet_data":                           {string(m.Packet.Data)},
			"packet_timeout_height_revision_number": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionNumber, 10)},
			"packet_timeout_height_revision_height": {strconv.FormatUint(m.Packet.TimeoutHeight.RevisionHeight, 10)},
			"packet_timeout_stamp":                  {strconv.FormatUint(m.Packet.TimeoutTimestamp, 10)},
			"acknowledgement":                       {string(m.Acknowledgement)},
			"proof_acked":                           {proofAcked},
			"proof_height_revision_number":          {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height":          {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
		},
	}
	err = ParsePacket(m.Packet.Data, &event)
	return event, err
}

func ParsePacket(data []byte, event *structs.SubsetEvent) error {
	var packetData *types.FungibleTokenPacketData
	err := json.Unmarshal(data, &packetData)
	if err != nil {
		return fmt.Errorf("packet malformed: %w", err)
	}
	amt, ok := new(big.Int).SetString(packetData.Amount, 10)
	if !ok {
		return fmt.Errorf("packet amount not a string: %v", packetData)
	}
	if amt.Cmp(bigZero) < 0 || len(packetData.Denom) == 0 || len(packetData.Sender) == 0 || len(packetData.Receiver) == 0 {
		return fmt.Errorf("packet malformed: %v", packetData)
	}
	// adding the Amount on the receiver.
	event.Sender = []structs.EventTransfer{
		{
			Account: structs.Account{ID: packetData.Sender},
		},
	}
	event.Recipient = []structs.EventTransfer{
		{
			Account: structs.Account{ID: packetData.Receiver},
			Amounts: []structs.TransactionAmount{
				{
					Text:     amt.String(),
					Numeric:  amt,
					Currency: packetData.Denom,
				},
			},
		},
	}
	return nil
}
