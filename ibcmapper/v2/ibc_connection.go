package ibcmapper

import (
	"fmt"
	"strconv"

	"github.com/figment-networks/indexing-engine/structs"
	"github.com/figment-networks/ni-cosmoslib/util"

	connection "github.com/cosmos/ibc-go/v2/modules/core/03-connection/types"
	"github.com/gogo/protobuf/proto"
)

// IBCConnectionOpenInitToSub transforms ibc.MsgConnectionOpenInit sdk messages to SubsetEvent
func IBCConnectionOpenInitToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenInit{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_init type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"connection_open_init"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"client_id":                  {m.ClientId},
			"version_identifier":         {versionIdentifier(m.Version)},
			"version_features":           versionFeatures(m.Version),
			"delay_period":               {strconv.FormatUint(m.DelayPeriod, 10)},
			"counterparty_client_id":     {m.Counterparty.ClientId},
			"counterparty_connection_id": {m.Counterparty.ConnectionId},
			"counterparty_prefix":        {string(m.Counterparty.Prefix.String())},
		},
	}, nil
}

// IBCConnectionOpenConfirmToSub transforms ibc.MsgConnectionOpenConfirm sdk messages to SubsetEvent
func IBCConnectionOpenConfirmToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenConfirm{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_confirm type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofAck, err := util.EncodeToB64(m.ProofAck, "proof_ack")
	if err != nil {
		return se, err
	}

	return structs.SubsetEvent{
		Type:   []string{"connection_open_confirm"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"connection_id":                {m.ConnectionId},
			"proof_ack":                    {proofAck},
			"proof_height_revision_number": {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height": {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
		},
	}, nil
}

// IBCConnectionOpenAckToSub transforms ibc.MsgConnectionOpenAck sdk messages to SubsetEvent
func IBCConnectionOpenAckToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenAck{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_ack type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofTry, err := util.EncodeToB64(m.ProofTry, "proof_try")
	if err != nil {
		return se, err
	}
	proofClient, err := util.EncodeToB64(m.ProofClient, "proof_client")
	if err != nil {
		return se, err
	}
	proofConsensus, err := util.EncodeToB64(m.ProofConsensus, "proof_consensus")
	if err != nil {
		return se, err
	}

	return structs.SubsetEvent{
		Type:   []string{"connection_open_ack"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"connection_id":                    {m.ConnectionId},
			"counterparty_connection_id":       {m.CounterpartyConnectionId},
			"version_identifier":               {versionIdentifier(m.Version)},
			"version_features":                 versionFeatures(m.Version),
			"client_state":                     {m.ClientState.String()},
			"proof_height_revision_number":     {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height":     {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
			"proof_try":                        {proofTry},
			"proof_client":                     {proofClient},
			"proof_consensus":                  {proofConsensus},
			"consensus_height_revision_number": {strconv.FormatUint(m.ConsensusHeight.RevisionNumber, 10)},
			"consensus_height_revision_height": {strconv.FormatUint(m.ConsensusHeight.RevisionHeight, 10)},
		},
	}, nil
}

// IBCConnectionOpenTryToSub transforms ibc.MsgConnectionOpenTry sdk messages to SubsetEvent
func IBCConnectionOpenTryToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenTry{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_try type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofInit, err := util.EncodeToB64(m.ProofInit, "proof_init")
	if err != nil {
		return se, err
	}
	proofClient, err := util.EncodeToB64(m.ProofClient, "proof_client")
	if err != nil {
		return se, err
	}
	proofConsensus, err := util.EncodeToB64(m.ProofConsensus, "proof_consensus")
	if err != nil {
		return se, err
	}

	se = structs.SubsetEvent{
		Type:   []string{"connection_open_try"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"client_id":                        {m.ClientId},
			"previous_connection_id":           {m.PreviousConnectionId},
			"client_state":                     {m.ClientState.String()},
			"counterparty_client_id":           {m.Counterparty.ClientId},
			"counterparty_connection_id":       {m.Counterparty.ConnectionId},
			"counterparty_prefix":              {string(m.Counterparty.Prefix.String())},
			"delay_period":                     {strconv.FormatUint(m.DelayPeriod, 10)},
			"counterparty_versions":            {},
			"proof_height_revision_number":     {strconv.FormatUint(m.ProofHeight.RevisionNumber, 10)},
			"proof_height_revision_height":     {strconv.FormatUint(m.ProofHeight.RevisionHeight, 10)},
			"proof_init":                       {proofInit},
			"proof_client":                     {proofClient},
			"proof_consensus":                  {proofConsensus},
			"consensus_height_revision_number": {strconv.FormatUint(m.ConsensusHeight.RevisionNumber, 10)},
			"consensus_height_revision_height": {strconv.FormatUint(m.ConsensusHeight.RevisionHeight, 10)},
		},
	}

	for i, cpv := range m.CounterpartyVersions {
		se.Additional[fmt.Sprintf("counterparty_version_identifier_%d", i)] = []string{cpv.Identifier}
		se.Additional[fmt.Sprintf("counterparty_version_features_%d", i)] = cpv.Features
	}

	return se, nil
}

func versionIdentifier(version *connection.Version) string {

	if version == nil {
		return ""
	}

	return version.Identifier
}

func versionFeatures(version *connection.Version) []string {

	if version == nil {
		return nil
	}

	return version.Features
}
