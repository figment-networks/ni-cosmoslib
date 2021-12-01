package mapper

import (
	"fmt"

	"github.com/figment-networks/indexing-engine/structs"

	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/gogo/protobuf/proto"
)

// FeegrantGrantAllowance transforms feegrant.MsgGrantAllowance sdk messages to SubsetEvent
func (mapper *Mapper) FeegrantGrantAllowance(msg []byte) (se structs.SubsetEvent, err error) {
	m := &feegrant.MsgGrantAllowance{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a grant_allowance type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"grant_allowance"},
		Module: "feegrant",
		Additional: map[string][]string{
			"granter":   {m.Granter},
			"grantee":   {m.Grantee},
			"allowance": {string(m.Allowance.Value)},
		},
	}, nil
}

// FeegrantGrantAllowanceResponse transforms feegrant.MsgGrantAllowanceResponse sdk messages to SubsetEvent
func (mapper *Mapper) FeegrantGrantAllowanceResponse(msg []byte) (se structs.SubsetEvent, err error) {
	m := &feegrant.MsgGrantAllowanceResponse{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a grant_allowance_response type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"grant_allowance_response"},
		Module: "feegrant",
	}, nil
}

// FeegrantRevokeAllowance transforms feegrant.MsgRevokeAllowance sdk messages to SubsetEvent
func (mapper *Mapper) FeegrantRevokeAllowance(msg []byte) (se structs.SubsetEvent, err error) {
	m := &feegrant.MsgRevokeAllowance{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a revoke_allowance type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"revoke_allowance"},
		Module: "feegrant",
		Additional: map[string][]string{
			"granter": {m.Granter},
			"grantee": {m.Grantee},
		},
	}, nil
}

// FeegrantRevokeAllowanceResponse transforms feegrant.MsgRevokeAllowanceResponse sdk messages to SubsetEvent
func (mapper *Mapper) FeegrantRevokeAllowanceResponse(msg []byte) (se structs.SubsetEvent, err error) {
	m := &feegrant.MsgRevokeAllowanceResponse{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a revoke_allowance_response type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"revoke_allowance_response"},
		Module: "feegrant",
	}, nil
}
