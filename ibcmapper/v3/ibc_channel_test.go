package ibcmapper

import (
	"testing"

	channel "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

func TestIBCChannelRecvPacketToSub(t *testing.T) {

	// Just test that the proof commitment is actually being encoded as base64.
	proofCommitment := []byte("aa\u0000\u0000bb")
	msg := channel.MsgRecvPacket{ProofCommitment: proofCommitment}
	bMsg, err := msg.Marshal()
	if err != nil {
		t.Errorf("unexpected marshal err: %s", err.Error())
		return
	}

	subsetEvent, err := IBCChannelRecvPacketToSub(bMsg)
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
