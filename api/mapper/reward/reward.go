package reward

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/figment-networks/indexing-engine/proto/rewstruct"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"

	"github.com/figment-networks/ni-cosmoslib/api/util"
)

type Mapper struct {
	Logger          *zap.Logger
	DefaultCurrency string
}

var currencyRegexp = regexp.MustCompile(`^\d+$`)

// delegate undelegate redelegate, -> addresses
// delegate undelegate redelegate + withdraw delegator rewards -> delagator rewards
// withdraw validator commision -> validator rewards

// ValidatorFromTx maps the resolved `ValidatorSrc` and `ValidatorDst` from processing `ParseRewardEvent` into a common function
func ValidatorFromTx(tx *rewstruct.Tx) string {
	var validator string
	switch tx.Type {
	case "MsgWithdrawDelegatorReward":
		validator = tx.ValidatorSrc
	case "MsgUndelegate":
		validator = tx.ValidatorSrc
	case "MsgDelegate":
		validator = tx.ValidatorDst
	case "MsgBeginRedelegate":
		validator = tx.ValidatorSrc
	}
	return validator
}

// ParseRewardEvent converts a cosmos event from the log to a Subevent type and adds it to the provided RewardEvent struct
func ParseRewardEvent(module, msgType string, raw []byte, lg types.ABCIMessageLog, ma *Mapper) (rev *rewstruct.RewardTx, err error) {

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

/*
func (m *Mapper) grouper(lg types.ABCIMessageLog, delegator string, amount_by string) (rev *rewstruct.RewardTx, err error) {
	rev = &rewstruct.RewardTx{}
	if len(lg.GetEvents()) > 5 {
		m.Logger.Warn("unexpected events length", zap.Any("events", lg.GetEvents())) // It would be good to test that kind of event
	}

	for _, ev := range lg.GetEvents() {
		parsed, err := m.groupEvents(ev)
		if err != nil {
			return rev, err
		}
		// if len(parsed) > 1 {
		// 	m.Logger.Warn("multiple event", zap.String("type", ev.GetType())) // is that possible?
		// }

		switch ev.GetType() {
		case "coin_received":

			if amount_by == "coin_received" {
				for _, p := range parsed {
					rev.Recipient = append(rev.Recipient, p["receiver"])

					switch p["receiver"] {
					case delegator:
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						rev.Rewards = append(rev.Rewards, am...)
					default:
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						rev.Amounts = append(rev.Amounts, am...)
					}
				}
			} else if amount_by == "redelegate" {
				for _, p := range parsed {
					rev.Recipient = append(rev.Recipient, p["receiver"])
					if p["receiver"] == delegator {
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						rev.Rewards = append(rev.Rewards, am...)
					}
				}
			} else if amount_by == "withdraw_rewards" {
				for _, p := range parsed {
					rev.Recipient = append(rev.Recipient, p["receiver"])
					am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Rewards = append(rev.Rewards, am...)
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
					am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Rewards = append(rev.Rewards, am...)
				}
			}
		case "withdraw_rewards":
			// MsgWithdrawDelegatorReward
			continue
		case "delegate":
			// MsgDelegate
			if amount_by == "delegate" {
				for _, p := range parsed {
					am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Amounts = append(rev.Amounts, am...)
				}
			}
		case "redelegate":
			// MsgBeginRedelegate
			for _, p := range parsed {
				if amount_by == "redelegate" {
					am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Amounts = append(rev.Amounts, am...)
				}
			}
		case "unbond":
			// MsgUndelegate
			continue
		default:
			// err = fmt.Errorf("unsupported event: %s", ev.GetType())
			// logger.Error(err)
			// other events for log purpouse
			continue
		}
	}
	return rev, err
}
*/

func (m *Mapper) MsgWithdrawValidatorCommission(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &distribution.MsgWithdrawValidatorCommission{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:         "MsgWithdrawValidatorCommission",
		ValidatorDst: wvc.ValidatorAddress,
	}

	for _, ev := range lg.GetEvents() {
		if ev.GetType() == "withdraw_commission" {
			parsed, err := m.groupEvents(ev)
			if err != nil {
				return rev, err
			}
			if val, ok := parsed[0]["amount"]; ok {
				am, err := fAmounts(m.DefaultCurrency, strings.Split(val, ","))
				if err != nil {
					return rev, err
				}
				rev.Amounts = append(rev.Amounts, am...)
			}

		} else {
			continue
		}
	}

	return rev, nil
}

func (m *Mapper) MsgWithdrawDelegatorReward(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &distribution.MsgWithdrawDelegatorReward{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	// Regardless of where the tokens went, the reward counts for the delegator
	//	rev, err = m.grouper(lg, wvc.DelegatorAddress, "withdraw_rewards")
	rev = &rewstruct.RewardTx{}
	if err != nil {
		return rev, err
	}
	rev.Type = "MsgWithdrawDelegatorReward"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorAddress

	return rev, nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgUndelegate sdk messages to SubsetEvent
func (m *Mapper) MsgUndelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgUndelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	//	rev, err = m.grouper(lg, wvc.DelegatorAddress, "coin_received")
	rev = &rewstruct.RewardTx{}
	if err != nil {
		return rev, err
	}
	rev.Type = "MsgUndelegate"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorAddress

	return rev, nil
}

// m transforms distribution.MsgDelegate sdk messages to SubsetEvent
func (m *Mapper) MsgDelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgDelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:         "MsgDelegate",
		Delegator:    wvc.DelegatorAddress,
		ValidatorDst: wvc.ValidatorAddress,
	}

	delegateAmount := ""
	for _, e := range []string{"delegate", "coin_received"} {
		for _, ev := range lg.GetEvents() {
			if e == ev.GetType() {
				switch ev.GetType() {
				case "delegate":
					parsed, err := m.groupEvents(ev)
					if err != nil {
						return rev, err
					}
					if val, ok := parsed[0]["amount"]; ok {
						delegateAmount = parsed[0]["amount"]
						am, err := fAmounts(m.DefaultCurrency, strings.Split(val, ","))
						if err != nil {
							return rev, err
						}
						rev.Amounts = append(rev.Amounts, am...)
					}
				case "coin_received":
					parsed, err := m.groupRSEvents(lg.GetEvents())
					if err != nil {
						return rev, err
					}
					for _, p := range parsed {
						if p["amount"] == delegateAmount {
							continue
						}
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						reward := &rewstruct.RewardAmount{
							Amounts:   am,
							Validator: wvc.ValidatorAddress,
						}
						rev.Rewards = append(rev.Rewards, reward)
					}
				}
			} else {
				continue
			}

		}
	}

	return rev, nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgBeginRedelegate sdk messages to SubsetEvent
func (m *Mapper) MsgBeginRedelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgBeginRedelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:         "MsgBeginRedelegate",
		ValidatorSrc: wvc.ValidatorSrcAddress,
		ValidatorDst: wvc.ValidatorDstAddress,
		Delegator:    wvc.DelegatorAddress,
	}

	for _, e := range []string{"redelegate", "coin_received"} {
		for _, ev := range lg.GetEvents() {
			if e == ev.GetType() {
				switch ev.GetType() {
				case "redelegate":
					parsed, err := m.groupEvents(ev)
					if err != nil {
						return rev, err
					}
					if val, ok := parsed[0]["amount"]; ok {
						//						redelegateAmount = parsed[0]["amount"]
						am, err := fAmounts(m.DefaultCurrency, strings.Split(val, ","))
						if err != nil {
							return rev, err
						}
						rev.Amounts = append(rev.Amounts, am...)
					}
				case "coin_received":
					parsed, err := m.groupRSEvents(lg.GetEvents())
					if err != nil {
						return rev, err
					}
					for i, p := range parsed {
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						reward := &rewstruct.RewardAmount{
							Amounts: am,
						}
						if i%2 == 0 { // TODO can we even do sth like this?
							reward.Validator = wvc.ValidatorSrcAddress
						} else {
							reward.Validator = wvc.ValidatorDstAddress
						}
						rev.Rewards = append(rev.Rewards, reward)
					}
				}
			} else {
				continue
			}

		}
	}

	return rev, nil
}

func (m *Mapper) MsgEditValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgEditValidator{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{}
	rev.Type = "MsgEditValidator"

	// rev, err = grouper(lg, "xxx")
	// if err != nil {
	// 	return rev, err
	// }

	return rev, nil
}

func (m *Mapper) MsgCreateValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgCreateValidator{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{}
	rev.Delegator = wvc.DelegatorAddress
	rev.Type = "MsgCreateValidator"

	return rev, nil
}

func (m *Mapper) MsgSetWithdrawAddress(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &distribution.MsgSetWithdrawAddress{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:      "MsgSetWithdrawAddress",
		Delegator: wvc.DelegatorAddress,
	}

	return rev, nil
}

func (m *Mapper) MsgFundCommunityPool(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &distribution.MsgFundCommunityPool{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type: "MsgFundCommunityPool",
	}
	//	rev.Sender = []string{wvc.Depositor}

	return rev, nil
}

func fAmounts(defaultCurrency string, amounts []string) (am []*rewstruct.Amount, err error) {

	for _, amt := range amounts {
		attrAmt := &rewstruct.Amount{}
		// We add a default currency mainly for old heights where there was only one value (see osmosis test for height 36). Whereas now you can have multiple currencies there.
		if currencyRegexp.MatchString(amt) {
			amt = amt + defaultCurrency
		}

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

var rewardEvents = map[string][]string{
	"coin_received":       {"receiver", "amount"},
	"coin_spent":          {"spender", "amount"},
	"delegate":            {"amount"},
	"redelegate":          {"amount"},
	"withdraw_commission": {"amount"},
}

func (m *Mapper) eventsFilter(ev types.StringEvent) (result []types.Attribute, err error) {

	attr := ev.GetAttributes()
	etype := ev.GetType()
	elen, exists := rewardEvents[etype]
	if !exists {
		// skip if no exists
		return result, nil
	}
	for _, v := range attr {
		for _, k := range elen {
			if k == v.Key {
				localAttr := v
				result = append(result, localAttr)
				m.Logger.Debug("event", zap.String("type", etype), zap.Any("content", v))
			}
			continue
		}
	}

	return result, nil

}

func (m *Mapper) groupEvents(ev types.StringEvent) (result []map[string]string, err error) {
	events, err := m.eventsFilter(ev)
	// end if there was no events
	if err != nil && len(events) == 0 {
		return result, err
	}
	etype := ev.GetType()
	elen, exists := rewardEvents[etype]
	if !exists {
		return result, nil
	}

	eventLen := len(elen)
	for i := 0; i < len(events); i = i + eventLen {
		emap := make(map[string]string)

		for j := 0; j < eventLen; j++ {
			if i+j <= len(events) {
				a := events[i+j]
				emap[a.Key] = a.Value
			} else {
				m.Logger.Error("rewardEvents missconfigured")
			}
		}

		result = append(result, emap)
	}

	return result, nil
}

func (m *Mapper) groupRSEvents(ev types.StringEvents) (result []map[string]string, err error) {
	received := []types.Attribute{}
	send := []types.Attribute{}

	for _, r := range ev {
		switch r.GetType() {
		case "coin_received":
			received = append(received, r.GetAttributes()...)
		case "coin_spent":
			send = append(send, r.GetAttributes()...)
		}
	}
	if len(received) != len(send) {
		return result, fmt.Errorf("lack of consistency in coin_received/coin_spent events")
	}
	eventLen := len(rewardEvents["coin_received"])
	for i := 0; i < len(received); i = i + eventLen {
		emap := make(map[string]string)

		for j := 0; j < eventLen; j++ {
			if i+j <= len(received) {
				r, s := received[i+j], send[i+j]
				emap[r.Key] = r.Value
				emap[s.Key] = s.Value
			}
		}
		result = append(result, emap)
	}

	return result, nil
}

// a = [[a:1], [b:1], [a:2], [b:2], [a:3], [b:3]]
// b = [[a:1], [c:1], [a:2], [c:2], [a:3], [c:3]]

// out = [{a:1,b:1 c:1}, {a:2,b:2, c:2}, {a:3, b:3, c:3}]
