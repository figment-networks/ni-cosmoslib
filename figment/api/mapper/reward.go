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
	"github.com/figment-networks/indexing-engine/structs"
	"github.com/figment-networks/indexing-engine/worker/logger"
	"github.com/figment-networks/ni-cosmoslib/figment/api/util"
	"github.com/gogo/protobuf/proto"
)

// delegate undelegate redelegate, -> addresses
// delegate undelegate redelegate + withdraw delegator rewards -> delagator rewards
// withdraw validator commision -> validator rewards

func grouper(rev *structs.RewardEvent, lg types.ABCIMessageLog, amount_by string) (err error) {
	if len(lg.GetEvents()) > 5 {
		panic("just checking")
	}
	for _, ev := range lg.GetEvents() {

		switch ev.GetType() {
		case "coin_received":

			parsed, err := groupEvents(ev)
			if err != nil {
				return err
			}
			if amount_by == "coin_received" {
				for _, p := range parsed {
					rev.RecipientAddress = append(rev.RecipientAddress, p["receiver"])

					// switch p["receiver"] {
					// case rev.DelegatorAddress:
					// 	am, rew, err := fAmounts2("reward", strings.Split(p["amount"], ","))
					// 	if err != nil {
					// 		return rev, err
					// 	}
					// 	rev.Amounts = append(rev.Amounts, am...)
					// 	rev.Rewards = append(rev.Rewards, rew...)
					// default:
					// 	am, rew, err := fAmounts2("amount", strings.Split(p["amount"], ","))
					// 	if err != nil {
					// 		return rev, err
					// 	}
					// 	rev.Amounts = append(rev.Amounts, am...)
					// 	rev.Rewards = append(rev.Rewards, rew...)
					// }

					if p["receiver"] == rev.DelegatorAddress {
						fAmounts("reward", strings.Split(p["amount"], ","), rev)
					} else {
						fAmounts("amount", strings.Split(p["amount"], ","), rev)
					}
				}
			} else if amount_by == "redelegate" {
				for _, p := range parsed {
					rev.RecipientAddress = append(rev.RecipientAddress, p["receiver"])
					if p["receiver"] == rev.DelegatorAddress {
						fAmounts("reward", strings.Split(p["amount"], ","), rev)
					}
				}
			}

		case "coin_spent":
			parsed, err := groupEvents(ev)
			if err != nil {
				return err
			}
			for _, p := range parsed {
				rev.SenderAddress = append(rev.SenderAddress, p["spender"])
			}
		case "withdraw_commission":
			// MsgWithdrawValidatorCommission
			parsed, err := groupEvents(ev)
			if err != nil {
				return err
			}
			if len(parsed) > 1 {
				logger.Error(fmt.Errorf("multiple withdraw_commission events")) // is that possible?
				panic("withdraw_commission")

			}
			if amount_by == "withdraw_commission" {
				for _, p := range parsed {
					fAmounts("amount", strings.Split(p["amount"], ","), rev)
				}
			}
		case "withdraw_rewards":
			// MsgWithdrawDelegatorReward
			parsed, err := groupEvents(ev)
			if err != nil {
				return err
			}
			if len(parsed) > 1 {
				logger.Error(fmt.Errorf("multiple withdraw_rewards events")) // is that possible?
				panic("withdraw_rewards")

			}
			// for _, p := range parsed {
			// 	fAmounts("reward", strings.Split(p["amount"], ","), rev)
			// }
		case "delegate":
			// MsgDelegate
			parsed, err := groupEvents(ev)
			if err != nil {
				return err
			}
			if len(parsed) > 1 {
				logger.Error(fmt.Errorf("multiple delegate events")) // is that possible?
				panic("delegate")

			}
			// counted already in coin_received
			// for _, p := range parsed {
			// 	fAmounts("amount", strings.Split(p["amount"], ","), rev)
			// }
		case "redelegate":
			// MsgBeginRedelegate
			parsed, err := groupEvents(ev)
			if err != nil {
				return err
			}
			if len(parsed) > 1 {
				logger.Error(fmt.Errorf("multiple redelegate events")) // is that possible?
				panic("redelegate")
			}
			if amount_by == "redelegate" { // TODO do not need it here ;)
				for _, p := range parsed {
					fAmounts("amount", strings.Split(p["amount"], ","), rev)
				}
			}
		case "unbond":
			// MsgUndelegate
			parsed, err := groupEvents(ev)
			if err != nil {
				return err
			}
			if len(parsed) > 1 {
				logger.Error(fmt.Errorf("multiple unbond events")) // is that possible?
				panic("unbond")
			}
			// for _, p := range parsed {
			// 	fAmounts("amount", strings.Split(p["amount"], ","), rev)
			// }
		default:
			// other events for log purpouse
			_, err := groupEvents(ev)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (mapper *Mapper) MsgWithdrawValidatorCommission(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &distribution.MsgWithdrawValidatorCommission{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	rev.ValidatorDstAddress = wvc.ValidatorAddress
	grouper(rev, lg, "withdraw_commission")

	return nil
}

func (mapper *Mapper) MsgWithdrawDelegatorReward(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &distribution.MsgWithdrawDelegatorReward{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	rev.DelegatorAddress = wvc.DelegatorAddress
	rev.ValidatorSrcAddress = wvc.ValidatorAddress
	grouper(rev, lg, "coin_received")
	// ZBADAC CO SI ETU DZEJE

	return nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgUndelegate sdk messages to SubsetEvent
func (mapper *Mapper) MsgUndelegate(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &staking.MsgUndelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	// mapped fields
	rev.DelegatorAddress = wvc.DelegatorAddress
	rev.ValidatorSrcAddress = wvc.ValidatorAddress
	grouper(rev, lg, "coin_received")

	return nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgDelegate sdk messages to SubsetEvent
func (mapper *Mapper) MsgDelegate(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &staking.MsgDelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	rev.DelegatorAddress = wvc.DelegatorAddress
	rev.ValidatorDstAddress = wvc.ValidatorAddress
	grouper(rev, lg, "coin_received")

	return nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgBeginRedelegate sdk messages to SubsetEvent
func (mapper *Mapper) MsgBeginRedelegate(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &staking.MsgBeginRedelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	rev.DelegatorAddress = wvc.DelegatorAddress
	rev.ValidatorSrcAddress = wvc.ValidatorSrcAddress
	rev.ValidatorDstAddress = wvc.ValidatorDstAddress
	grouper(rev, lg, "redelegate")

	return nil
}

// func (mapper *Mapper) MsgEditValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
// 	wvc := &staking.MsgEditValidator{}
// 	if err := proto.Unmarshal(msg, wvc); err != nil {
// 		return rev, fmt.Errorf("not a distribution type: %w", err)
// 	}

// 	rev = &rewstruct.Tx{}

// 	rev.Type = "MsgEditValidator"
// 	return rev, nil
// }

// func (mapper *Mapper) MsgCreateValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
// 	wvc := &staking.MsgCreateValidator{}
// 	if err := proto.Unmarshal(msg, wvc); err != nil {
// 		return rev, fmt.Errorf("not a distribution type: %w", err)
// 	}

// 	rev = &rewstruct.Tx{}
// 	rev.Delegator = wvc.DelegatorAddress
// 	rev.Validator = []string{wvc.ValidatorAddress}

// 	rev.Type = "MsgCreateValidator"
// 	return rev, nil
// }

// func (mapper *Mapper) MsgSetWithdrawAddress(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
// 	wvc := &distribution.MsgSetWithdrawAddress{}
// 	if err := proto.Unmarshal(msg, wvc); err != nil {
// 		return rev, fmt.Errorf("not a distribution type: %w", err)
// 	}

// 	rev = &rewstruct.Tx{}
// 	// mapped fields
// 	rev.Type = "MsgWithdrawDelegatorReward"
// 	return rev, nil
// }

// func (mapper *Mapper) MsgFundCommunityPool(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
// 	wvc := &distribution.MsgFundCommunityPool{}
// 	if err := proto.Unmarshal(msg, wvc); err != nil {
// 		return rev, fmt.Errorf("not a distribution type: %w", err)
// 	}

// 	rev = &rewstruct.Tx{}
// 	rev.Sender = []string{wvc.Depositor}

// 	rev.Type = "MsgFundCommunityPool"
// 	return rev, nil
// }

func fAmounts(field string, amounts []string, rev *structs.RewardEvent) (err error) {
	for _, amt := range amounts {
		attrAmt := structs.TransactionAmount{Numeric: &big.Int{}}

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
			return fmt.Errorf("[COSMOS-API] Error parsing amount '%s': %s ", amt, coinErr)
		}

		attrAmt.Text = amt
		attrAmt.Exp = exp
		attrAmt.Numeric.Set(c)

		if field == "amount" {
			rev.Amounts = append(rev.Amounts, attrAmt)
		} else if field == "reward" {
			rev.Rewards = append(rev.Rewards, attrAmt)
		}
	}
	return nil
}

func fAmounts2(field string, amounts []string) (am, rew []*rewstruct.Amount, err error) {
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
			return nil, nil, fmt.Errorf("[COSMOS-API] Error parsing amount '%s': %s ", amt, coinErr)
		}

		attrAmt.Numeric = c.Bytes()
		attrAmt.Text = amt
		attrAmt.Exp = exp

		if field == "amount" {
			am = append(am, attrAmt)
		} else if field == "reward" {
			rew = append(rew, attrAmt)
		}
	}
	return am, rew, nil
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
