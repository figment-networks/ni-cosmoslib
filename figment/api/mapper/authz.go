package mapper

import (
	"fmt"

	"github.com/figment-networks/indexing-engine/structs"

	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/gogo/protobuf/proto"
)

// AuthzGrantToSub transforms authz.MsgGrant sdk messages to SubsetEvent
func AuthzGrantToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &authz.MsgGrant{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a grant type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"grant"},
		Module: "authz",
		Additional: map[string][]string{
			"granter":             {m.Granter},
			"grantee":             {m.Grantee},
			"grant_expiration":    {m.Grant.Expiration.String()},
			"grant_authorization": {string(m.Grant.Authorization.Value)},
		},
	}, nil
}

// AuthzExecResponseToSub transforms authz.MsgExecResponse sdk messages to SubsetEvent
func AuthzExecResponseToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &authz.MsgExecResponse{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a exec_response type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"exec_response"},
		Module: "authz",
	}

	for _, result := range m.Results {
		se.Additional["results"] = append(se.Additional["results"], string(result))
	}

	return se, nil
}

// AuthzExecToSub transforms authz.MsgExec sdk messages to SubsetEvent
func AuthzExecToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &authz.MsgExec{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a exec type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"exec"},
		Module: "authz",
		Additional: map[string][]string{
			"grantee": {m.Grantee},
		},
	}

	for _, mg := range m.Msgs {
		se.Additional["msgs"] = append(se.Additional["msgs"], string(mg.Value))
	}

	return se, nil
}

// AuthzGrantResponseToSub transforms authz.MsgGrantResponse sdk messages to SubsetEvent
func AuthzGrantResponseToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &authz.MsgGrantResponse{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a grant_response type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"grant_response"},
		Module: "authz",
	}, nil
}

// AuthzMsgRevokeToSub transforms authz.MsgRevoke sdk messages to SubsetEvent
func AuthzMsgRevokeToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &authz.MsgRevoke{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a revoke type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"revoke"},
		Module: "authz",
		Additional: map[string][]string{
			"granter":      {m.Granter},
			"grantee":      {m.Grantee},
			"msg_type_url": {m.MsgTypeUrl},
		},
	}, nil
}
