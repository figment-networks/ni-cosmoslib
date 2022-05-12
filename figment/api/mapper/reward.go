package mapper

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/figment-networks/indexing-engine/proto/rewstruct"
	"github.com/figment-networks/indexing-engine/worker/logger"
	"github.com/figment-networks/ni-cosmoslib/figment/api/util"
	"github.com/gogo/protobuf/proto"
)

// delegate undelegate redelegate, -> addresses
// delegate undelegate redelegate + withdraw delegator rewards -> delagator rewards
// withdraw validator commision -> validator rewards

// ParseRewardEvent converts a cosmos event from the log to a Subevent type and adds it to the provided RewardEvent struct
func ParseRewardEvent(module, msgType string, raw []byte, lg types.ABCIMessageLog, ma Mapper) (rev *rewstruct.Tx, err error) {

	switch module {
	case "distribution":
		switch msgType {
		case "MsgSetWithdrawAddress":
			return ma.MsgSetWithdrawAddress(raw, lg)
		case "MsgWithdrawValidatorCommission":
			return ma.MsgWithdrawValidatorCommission(raw, lg)
		case "MsgWithdrawDelegatorReward":
			return ma.MsgWithdrawDelegatorReward(raw, lg)
		case "MsgFundCommunityPool":
			return ma.MsgFundCommunityPool(raw, lg)
		}
	case "staking":
		switch msgType {
		case "MsgUndelegate":
			return ma.MsgUndelegate(raw, lg)
		case "MsgDelegate":
			return ma.MsgDelegate(raw, lg)
		case "MsgBeginRedelegate":
			return ma.MsgBeginRedelegate(raw, lg)
		case "MsgEditValidator":
			return ma.MsgEditValidator(raw, lg)
		case "MsgCreateValidator":
			return ma.MsgCreateValidator(raw, lg)
		}

	}

	return
}

func grouper(lg types.ABCIMessageLog, amount_by string) (rev *rewstruct.Tx, err error) {
	if len(lg.GetEvents()) > 5 {
		err = fmt.Errorf("unexpected events length: %s", lg.GetEvents())
		logger.Error(err) // It would be good to test that kind of event
	}
	for _, ev := range lg.GetEvents() {
		parsed, err := groupEvents(ev)
		if err != nil {
			return rev, err
		}
		if len(parsed) > 1 {
			err = fmt.Errorf("multiple %s events", ev.GetType())
			logger.Error(err) // is that possible?
		}

		switch ev.GetType() {
		case "coin_received":

			if amount_by == "coin_received" {
				for _, p := range parsed {
					rev.Recipient = append(rev.Recipient, p["receiver"])

					switch p["receiver"] {
					case rev.Delegator:
						am, err := fAmounts(strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						rev.Rewards = append(rev.Rewards, am...)
					default:
						am, err := fAmounts(strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						rev.Amounts = append(rev.Amounts, am...)
					}
				}
			} else if amount_by == "redelegate" {
				for _, p := range parsed {
					rev.Recipient = append(rev.Recipient, p["receiver"])
					if p["receiver"] == rev.Delegator {
						am, err := fAmounts(strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						rev.Rewards = append(rev.Rewards, am...)
					}
				}
			}

		case "coin_spent":
			for _, p := range parsed {
				rev.Sender = append(rev.Sender, p["spender"])
			}
		case "withdraw_commission":
			// MsgWithdrawValidatorCommission
			if amount_by == "withdraw_commission" {
				for _, p := range parsed {
					am, err := fAmounts(strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Amounts = append(rev.Amounts, am...)
				}
			}
		case "withdraw_rewards":
			// MsgWithdrawDelegatorReward
			continue
			// for _, p := range parsed {
			// 	fAmounts("reward", strings.Split(p["amount"], ","), rev)
			// }
		case "delegate":
			// MsgDelegate
			continue
			// counted already in coin_received
			// for _, p := range parsed {
			// 	fAmounts("amount", strings.Split(p["amount"], ","), rev)
			// }
		case "redelegate":
			// MsgBeginRedelegate
			for _, p := range parsed {
				if amount_by == "redelegate" { // TODO do not need it here ;)
					am, err := fAmounts(strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Amounts = append(rev.Amounts, am...)
				}
			}
		case "unbond":
			// MsgUndelegate

			continue
			// for _, p := range parsed {
			// 	fAmounts("amount", strings.Split(p["amount"], ","), rev)
			// }
		default:
			err = fmt.Errorf("unsupported event: %s", ev.GetType())
			logger.Warn(err.Error())
			// other events for log purpouse
			_, err := groupEvents(ev)
			if err != nil {
				return rev, err
			}
		}
	}
	return rev, err
}

func (mapper *Mapper) MsgWithdrawValidatorCommission(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	//func (mapper *Mapper) MsgWithdrawValidatorCommission(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &distribution.MsgWithdrawValidatorCommission{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgWithdrawValidatorCommission"
	rev.ValidatorDst = wvc.ValidatorAddress

	rev, err = grouper(lg, "withdraw_commission")
	if err != nil {
		return rev, err
	}

	return rev, nil
}

func (mapper *Mapper) MsgWithdrawDelegatorReward(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &distribution.MsgWithdrawDelegatorReward{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgWithdrawDelegatorReward"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorAddress

	rev, err = grouper(lg, "coin_received")
	if err != nil {
		return rev, err
	}
	// ZBADAC CO SI ETU DZEJE

	return rev, nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgUndelegate sdk messages to SubsetEvent
func (mapper *Mapper) MsgUndelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgUndelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgUndelegate"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorAddress

	rev, err = grouper(lg, "coin_received")
	if err != nil {
		return rev, err
	}

	return rev, nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgDelegate sdk messages to SubsetEvent
func (mapper *Mapper) MsgDelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgDelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgDelegate"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorDst = wvc.ValidatorAddress

	rev, err = grouper(lg, "coin_received")
	if err != nil {
		return rev, err
	}

	return rev, nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgBeginRedelegate sdk messages to SubsetEvent
func (mapper *Mapper) MsgBeginRedelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgBeginRedelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgBeginRedelegate"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorSrcAddress
	rev.ValidatorDst = wvc.ValidatorDstAddress

	rev, err = grouper(lg, "redelegate")
	if err != nil {
		return rev, err
	}

	return rev, nil
}

func (mapper *Mapper) MsgEditValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgEditValidator{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgEditValidator"

	// rev, err = grouper(lg, "xxx")
	// if err != nil {
	// 	return rev, err
	// }

	return rev, nil
}

func (mapper *Mapper) MsgCreateValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgCreateValidator{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Delegator = wvc.DelegatorAddress
	rev.Validator = []string{wvc.ValidatorAddress}
	rev.Type = "MsgCreateValidator"

	// rev, err = grouper(lg, "xxx")
	// if err != nil {
	// 	return rev, err
	// }

	return rev, nil
}

func (mapper *Mapper) MsgSetWithdrawAddress(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &distribution.MsgSetWithdrawAddress{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgWithdrawDelegatorReward"

	// rev, err = grouper(lg, "xxx")
	// if err != nil {
	// 	return rev, err
	// }

	return rev, nil
}

func (mapper *Mapper) MsgFundCommunityPool(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &distribution.MsgFundCommunityPool{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Sender = []string{wvc.Depositor}
	rev.Type = "MsgFundCommunityPool"

	// rev, err = grouper(lg, "xxx")
	// if err != nil {
	// 	return rev, err
	// }

	return rev, nil
}

func fAmounts(amounts []string) (am []*rewstruct.Amount, err error) {
	for _, amt := range amounts {
		attrAmt := &rewstruct.Amount{}
		sliced := util.GetCurrency(amt)
		var (
			c       *big.Int
			exp     int32
			coinErr error
		)
		if len(sliced) == 3 {
			attrAmt.Currency = sliced[2]
			c, exp, coinErr = util.GetCoin(sliced[1])
		} else {
			c, exp, coinErr = util.GetCoin(amt)
		}
		if coinErr != nil {
			return nil, fmt.Errorf("[COSMOS-API] Error parsing amount '%s': %s ", amt, coinErr)
		}

		attrAmt.Numeric = c.Bytes()
		attrAmt.Text = amt
		attrAmt.Exp = exp

		am = append(am, attrAmt)
	}
	return am, nil
}

func groupEvents(ev types.StringEvent) (result [](map[string]string), err error) {
	attr := ev.GetAttributes()
	etype := ev.GetType()
	// elm - events length map
	elm := map[string]int{
		"coin_received":       2,         // multiple events
		"coin_spent":          2,         // multiple events
		"message":             len(attr), // 3 or 4 keys :/
		"transfer":            3,
		"withdraw_commission": 1, // MsgWithdrawValidatorCommission
		"withdraw_rewards":    2, // MsgWithdrawDelegatorReward
		"redelegate":          4, // MsgBeginRedelegate
		"delegate":            3, // MsgDelegate
		"unbond":              3, // MsgUndelegate
	}

	elen, exists := elm[etype]
	if !exists {
		return result, fmt.Errorf("missing in events length map: %s", etype)
	}

	for i := 0; i < len(attr); i = i + elen {
		emap := make(map[string]string)
		for j := 0; j < elen; j++ {
			emap[attr[i+j].Key] = attr[i+j].Value
			log.Println("type", etype, "key", attr[i+j].Key, "Value", attr[i+j].Value)
		}
		if len(emap) < elen {
			// logs ->0 ->events ->type message
			// https://www.mintscan.io/cosmos/txs/0BA41D804ED0195CAD8D65BDFA80202F6C0267CBDBCBD3E71CC3BB78DE40BACC
			log.Println("duplicated keys in event list, which may contain different data", attr)
		}
		result = append(result, emap)
	}

	return result, nil
}
