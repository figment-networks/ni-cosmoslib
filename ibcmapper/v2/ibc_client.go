package ibcmapper

import (
	"fmt"

	"github.com/figment-networks/indexing-engine/structs"
	shared "github.com/figment-networks/indexing-engine/structs"
	"github.com/figment-networks/ni-cosmoslib/api/util"

	client "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	"github.com/gogo/protobuf/proto"
)

// IBCCreateClientToSub transforms ibc.MsgCreateClient sdk messages to SubsetEvent
func IBCCreateClientToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &client.MsgCreateClient{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a create_client type: %w", err)
	}

	return shared.SubsetEvent{
		Type:   []string{"create_client"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"client_state":    {m.ClientState.String()},
			"consensus_state": {m.ConsensusState.String()},
		},
	}, nil
}

// IBCUpdateClientToSub transforms ibc.MsgUpdateClient sdk messages to SubsetEvent
func IBCUpdateClientToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &client.MsgUpdateClient{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a update_client type: %w", err)
	}

	return shared.SubsetEvent{
		Type:   []string{"update_client"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"client_id": {m.ClientId},
			"header":    {m.Header.String()},
		},
	}, nil
}

// IBCUpgradeClientToSub transforms ibc.MsgUpgradeClient sdk messages to SubsetEvent
func IBCUpgradeClientToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &client.MsgUpgradeClient{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a upgrade_client type: %w", err)
	}

	// Encode fields that can contain null bytes.
	proofUpgradeClient, err := util.EncodeToB64(m.ProofUpgradeClient, "proof_upgrade_client")
	if err != nil {
		return se, err
	}
	proofUpgradeConsensusState, err := util.EncodeToB64(m.ProofUpgradeConsensusState, "proof_upgrade_consensus_state")
	if err != nil {
		return se, err
	}

	return shared.SubsetEvent{
		Type:   []string{"upgrade_client"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"client_id":                     {m.ClientId},
			"client_state":                  {m.ClientState.String()},
			"consensus_state":               {m.ConsensusState.String()},
			"proof_upgrade_client":          {proofUpgradeClient},
			"proof_upgrade_consensus_state": {proofUpgradeConsensusState},
		},
	}, nil
}

// IBCSubmitMisbehaviourToSub transforms ibc.MsgSubmitMisbehaviour sdk messages to SubsetEvent
func IBCSubmitMisbehaviourToSub(msg []byte) (se shared.SubsetEvent, err error) {
	m := &client.MsgSubmitMisbehaviour{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a submit_misbehaviour type: %w", err)
	}

	return shared.SubsetEvent{
		Type:   []string{"submit_misbehaviour"},
		Module: "ibc",
		Node: map[string][]structs.Account{
			"signer": {{ID: m.Signer}},
		},
		Additional: map[string][]string{
			"client_id":    {m.ClientId},
			"misbehaviour": {m.Misbehaviour.String()},
		},
	}, nil
}
