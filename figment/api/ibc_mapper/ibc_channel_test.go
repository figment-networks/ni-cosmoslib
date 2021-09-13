package mapper_test

import (
	"testing"

	channel "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	mapper "github.com/figment-networks/ni-cosmoslib/figment/api/ibc_mapper"
)

func TestIBCChannelRecvPacketToSub(t *testing.T) {

	// Just test that the proof commitment is actually being encoded as base64.
	proofCommitment := "aa\u0000\u0000bb"
	msg := channel.MsgRecvPacket{ProofCommitment: []byte(proofCommitment)}
	bMsg, err := msg.Marshal()
	if err != nil {
		t.Errorf("unexpected marshal err: %s", err.Error())
		return
	}

	subsetEvent, err := mapper.IBCChannelRecvPacketToSub(bMsg)
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
	}

	value, ok := subsetEvent.Additional["proof_commitment"]
	if !ok || len(value) == 0 {
		t.Error("missing proof_commitment value")
	}

	expValue := "YWEAAGJi"
	if value[0] != expValue {
		t.Errorf("expected %s, received %s", expValue, value[0])
	}
}
