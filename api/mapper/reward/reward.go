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
	"github.com/figment-networks/ni-cosmoslib/figment/api/util"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

type Mapper struct {
	Logger          *zap.Logger
	DefaultCurrency string
}

var currencyRegexp = regexp.MustCompile(`^\d+$`)

type amounts struct {
	rewards map[string]*rewstruct.Amount
	amounts map[string]*rewstruct.Amount
}

func newAmounts() *amounts {
	return &amounts{
		rewards: make(map[string]*rewstruct.Amount),
		amounts: make(map[string]*rewstruct.Amount),
	}
}

func mergeAmounts(m map[string]*rewstruct.Amount, amts []*rewstruct.Amount) {
	for _, amt := range amts {
		key := fmt.Sprintf("%s:%d", amt.Currency, amt.Exp)
		amtStruct, ok := m[key]
		if !ok {
			// copy amount
			cpamt := *amt
			amtStruct = &cpamt
			m[key] = amtStruct
		} else {
			combinedAmt := new(big.Int).Add(new(big.Int).SetBytes(amtStruct.Numeric), new(big.Int).SetBytes(amt.Numeric))
			amtStruct.Numeric = combinedAmt.Bytes()
			amtStruct.Text = fmt.Sprintf("%s%s", combinedAmt, amtStruct.Currency)
		}
	}
}

func (a *amounts) addAmounts(amts []*rewstruct.Amount) {
	mergeAmounts(a.amounts, amts)
}

func (a *amounts) addRewards(amts []*rewstruct.Amount) {
	mergeAmounts(a.rewards, amts)
}

func toAmmounts(m map[string]*rewstruct.Amount) (res []*rewstruct.Amount) {
	for _, v := range m {
		res = append(res, v)
	}
	return res
}

func (a *amounts) toRewards() (res []*rewstruct.Amount) {
	return toAmmounts(a.rewards)
}

func (a *amounts) toAmounts() (res []*rewstruct.Amount) {
	return toAmmounts(a.amounts)
}

type recipient struct {
	rcpts map[string]*amounts
}

func newRecipient() *recipient {
	return &recipient{rcpts: make(map[string]*amounts)}
}

func (r *recipient) populate(t *rewstruct.Tx) {
	for rcpt, amts := range r.rcpts {
		t.Recipient = append(t.Recipient, rcpt)
		t.Rewards = amts.toRewards()
		t.Amounts = amts.toAmounts()
	}
}

func (r *recipient) newRecipientToAmounts(recipient string) *amounts {
	amts, ok := r.rcpts[recipient]
	if !ok {
		amts = newAmounts()
		r.rcpts[recipient] = amts
	}
	return amts
}

func (r *recipient) addReward(recipient string, amt []*rewstruct.Amount) {
	amts := r.newRecipientToAmounts(recipient)
	amts.addRewards(amt)
}

func (r *recipient) addAmount(recipient string, amt []*rewstruct.Amount) {
	amts := r.newRecipientToAmounts(recipient)
	amts.addAmounts(amt)
}

type spender struct {
	senders map[string]struct{}
}

func newSpenders() *spender {
	return &spender{senders: make(map[string]struct{})}
}

func (s *spender) populate(rev *rewstruct.Tx) {
	for k := range s.senders {
		rev.Sender = append(rev.Sender, k)
	}
}

func (s *spender) addSender(sender string) {
	s.senders[sender] = struct{}{}
}

// delegate undelegate redelegate, -> addresses
// delegate undelegate redelegate + withdraw delegator rewards -> delagator rewards
// withdraw validator commision -> validator rewards

// ParseRewardEvent converts a cosmos event from the log to a Subevent type and adds it to the provided RewardEvent struct
func ParseRewardEvent(module, msgType string, raw []byte, lg types.ABCIMessageLog, ma *Mapper) (rev *rewstruct.Tx, err error) {

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

func (m *Mapper) grouper(lg types.ABCIMessageLog, delegator string, amount_by string) (rev *rewstruct.Tx, err error) {
	rev = &rewstruct.Tx{}

	recipients := newRecipient()
	spenders := newSpenders()

	if len(lg.GetEvents()) > 5 {
		m.Logger.Warn("unexpected events length", zap.Any("events", lg.GetEvents())) // It would be good to test that kind of event
	}

	for _, ev := range lg.GetEvents() {
		parsed, err := m.groupEvents(ev)
		if err != nil {
			return rev, err
		}
		if len(parsed) > 1 {
			m.Logger.Warn("multiple event", zap.String("type", ev.GetType())) // is that possible?
		}

		switch ev.GetType() {
		case "coin_received":
			if amount_by == "coin_received" {
				for _, p := range parsed {
					switch p["receiver"] {
					case delegator:
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						recipients.addReward(p["receiver"], am)
					default:
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						recipients.addAmount(p["receiver"], am)
					}
				}
			} else if amount_by == "redelegate" {
				for _, p := range parsed {
					if p["receiver"] == delegator {
						am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
						if err != nil {
							return rev, err
						}
						recipients.addReward(p["receiver"], am)
					}
				}
			}

		case "coin_spent":
			for _, p := range parsed {
				spenders.addSender(p["spender"])
			}
		case "withdraw_commission":
			// MsgWithdrawValidatorCommission
			if amount_by == "withdraw_commission" {
				for _, p := range parsed {
					am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
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
			if amount_by == "delegate" {
				for _, p := range parsed {
					am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Amounts = append(rev.Amounts, am...)
				}
			}
			// counted already in coin_received
			// for _, p := range parsed {
			// 	fAmounts("amount", strings.Split(p["amount"], ","), rev)
			// }
		case "redelegate":
			// MsgBeginRedelegate
			for _, p := range parsed {
				if amount_by == "redelegate" { // TODO do not need it here ;)
					am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
					if err != nil {
						return rev, err
					}
					rev.Amounts = append(rev.Amounts, am...)
				}
			}
		case "unbond":
			// MsgUndelegate
			// for _, p := range parsed {
			// 	if amount_by == "unbond" { // TODO do not need it here ;)
			// 		am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
			// 		if err != nil {
			// 			return rev, err
			// 		}
			// 		rev.Amounts = append(rev.Amounts, am...)
			// 	}
			// }
			continue
		default:
			// err = fmt.Errorf("unsupported event: %s", ev.GetType())
			// logger.Error(err)
			// other events for log purpouse
			continue
		}
	}
	recipients.populate(rev)
	spenders.populate(rev)
	return rev, err
}

func (m *Mapper) MsgWithdrawValidatorCommission(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	//func (mapper *Mapper) MsgWithdrawValidatorCommission(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &distribution.MsgWithdrawValidatorCommission{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev, err = m.grouper(lg, "", "withdraw_commission")
	if err != nil {
		return rev, err
	}
	rev.Type = "MsgWithdrawValidatorCommission"
	rev.ValidatorDst = wvc.ValidatorAddress

	return rev, nil
}

func (m *Mapper) MsgWithdrawDelegatorReward(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &distribution.MsgWithdrawDelegatorReward{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev, err = m.grouper(lg, wvc.DelegatorAddress, "coin_received")
	if err != nil {
		return rev, err
	}
	rev.Type = "MsgWithdrawDelegatorReward"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorAddress

	// validate it

	return rev, nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgUndelegate sdk messages to SubsetEvent
func (m *Mapper) MsgUndelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgUndelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev, err = m.grouper(lg, wvc.DelegatorAddress, "coin_received")
	if err != nil {
		return rev, err
	}
	rev.Type = "MsgUndelegate"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorAddress

	return rev, nil
}

// m transforms distribution.MsgDelegate sdk messages to SubsetEvent
func (m *Mapper) MsgDelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgDelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev, err = m.grouper(lg, wvc.DelegatorAddress, "delegate")
	if err != nil {
		return rev, err
	}
	rev.Type = "MsgDelegate"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorDst = wvc.ValidatorAddress

	return rev, nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgBeginRedelegate sdk messages to SubsetEvent
func (m *Mapper) MsgBeginRedelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &staking.MsgBeginRedelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev, err = m.grouper(lg, wvc.DelegatorAddress, "redelegate")
	if err != nil {
		return rev, err
	}
	rev.Type = "MsgBeginRedelegate"
	rev.Delegator = wvc.DelegatorAddress
	rev.ValidatorSrc = wvc.ValidatorSrcAddress
	rev.ValidatorDst = wvc.ValidatorDstAddress

	return rev, nil
}

func (m *Mapper) MsgEditValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
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

func (m *Mapper) MsgCreateValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
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

func (m *Mapper) MsgSetWithdrawAddress(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
	wvc := &distribution.MsgSetWithdrawAddress{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.Tx{}
	rev.Type = "MsgSetWithdrawAddress"
	rev.Delegator = wvc.DelegatorAddress

	// TODO categorize  withdraw address 	wvc.WithdrawAddress

	return rev, nil
}

func (m *Mapper) MsgFundCommunityPool(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.Tx, err error) {
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
	"coin_spent":          {"spender"},
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
