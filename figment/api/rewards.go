package api

import (
	"fmt"
	"log"
	"strings"

	codec_types "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/figment-networks/indexing-engine/structs"
	"github.com/figment-networks/ni-cosmoslib/figment/api/mapper"
)

// AddRewardEvent converts a cosmos event from the log to a Subevent type and adds it to the provided RewardEvent struct
func AddRewardEvent(rev *structs.RewardEvent, m *codec_types.Any, lg types.ABCIMessageLog, ma *mapper.Mapper) (err error) {

	// TypeUrl must be in the format "/cosmos.bank.v1beta1.MsgSend"

	tPath := strings.Split(m.TypeUrl, ".")
	if len(tPath) != 4 {
		return fmt.Errorf("problem with cosmos event cosmos event %s: %w", m.TypeUrl, ErrUnknownMessageType)
	}
	// for mapper = nil use the default
	if ma == nil {
		ma = defaultMapper
	}

	msgType := tPath[3]
	msgRoute := tPath[1]
	rev.Type = msgType
	var is bool

	log.Println("route: ", msgRoute, msgType)
	switch msgRoute {
	case "distribution":
		switch msgType {
		case "MsgWithdrawValidatorCommission":
			err = ma.MsgWithdrawValidatorCommission(m.Value, lg, rev)
		case "MsgWithdrawDelegatorReward":
			err = ma.MsgWithdrawDelegatorReward(m.Value, lg, rev)
		}

	case "staking":
		switch msgType {
		case "MsgUndelegate":
			err = ma.MsgUndelegate(m.Value, lg, rev)
		case "MsgDelegate":
			err = ma.MsgDelegate(m.Value, lg, rev)
		case "MsgBeginRedelegate":
			err = ma.MsgBeginRedelegate(m.Value, lg, rev)
		}
	}

	_ = is

	return nil
}
