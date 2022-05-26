package ibcmapper

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	channel "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
)

func TestIBCChannelRecvPacketToSub(t *testing.T) {

	packetData := &types.FungibleTokenPacketData{
		Receiver: "receiver",
		Sender:   "sender",
		Amount:   "1",
		Denom:    "denom",
	}
	packetDataBytes, _ := json.Marshal(packetData)
	// Just test that the proof commitment is actually being encoded as base64.
	proofCommitment := []byte("aa\u0000\u0000bb")
	msg := channel.MsgRecvPacket{
		ProofCommitment: proofCommitment,
		Packet: channel.Packet{
			Data: packetDataBytes,
		},
	}
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
