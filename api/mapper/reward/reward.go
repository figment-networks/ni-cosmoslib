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

	"github.com/figment-networks/ni-cosmoslib/client/cosmosgrpc"

	"github.com/figment-networks/ni-cosmoslib/api/util"
)

type Mapper struct {
	Logger              *zap.Logger
	DefaultCurrency     string
	BondedTokensPool    string
	NotBondedTokensPool string
}

// osmo BondedTokensPool    (delegate) osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09    https://www.mintscan.io/osmosis/txs/29039372E1308EFC7118B83E53BB88B03D7A877A200829150CA27338F77C405B
// osmo NotBondedTokensPool (undelegate) osmo1tygms3xhhs3yv487phx3dw4a95jn7t7lfqxwe3  https://www.mintscan.io/osmosis/txs/D7BF9ECFF1135B7D088EC8DE2685F98248924F66EF8083E97E076A2BA1C51420

// cosmos BondedTokensPool (delegate)      "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh" https://www.mintscan.io/cosmos/txs/39AA4A671D9436878000C64EFC6E12527EFE412A84A2A347BAB591C0415D4966
// cosmos NotBondedTokensPool (undelegate) "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r" https://www.mintscan.io/cosmos/txs/8AFC27C7DEC448DE0DFD9E419C11269601401A10CB054C6BB3BE4C1A45CE9C5D

// kava BondedTokensPool (delegate) "kava1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3fwaj0s"       https://www.mintscan.io/kava/txs/17B3C52FB4F876EC53FB18B4B2592F47B2CFB81BDD76E375F166036FB9DE59AA
// kava NotBondedTokensPool (undelegate) "kava1tygms3xhhs3yv487phx3dw4a95jn7t7lawprey"  https://www.mintscan.io/kava/txs/B06C50A36FB3F01584F190FB68BA28FF9C3CCC2E286CBDDECFBD5A6D3DD7C1A6

var currencyRegexp = regexp.MustCompile(`^\d+$`)

// delegate undelegate redelegate, -> addresses
// delegate undelegate redelegate + withdraw delegator rewards -> delagator rewards
// withdraw validator commision -> validator rewards

// ValidatorFromTx maps the resolved `ValidatorSrc` and `ValidatorDst` from processing `ParseRewardEvent` into a common function
func ValidatorFromTx(tx *rewstruct.RewardTx) string {
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

// MsgWithdrawValidatorCommission transforms distribution.MsgWithdrawValidatorCommission sdk messages and related events to RewardTx
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
			break
		}
	}

	return rev, nil
}

// MsgWithdrawDelegatorReward transforms distribution.MsgWithdrawDelegatorReward sdk messages and related events to RewardTx
func (m *Mapper) MsgWithdrawDelegatorReward(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &distribution.MsgWithdrawDelegatorReward{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:         "MsgWithdrawDelegatorReward",
		Delegator:    wvc.DelegatorAddress,
		ValidatorSrc: wvc.ValidatorAddress,
	}

	for _, ev := range lg.GetEvents() {
		if ev.GetType() == "coin_received" {
			parsed, err := m.groupRSEvents(lg.GetEvents())
			if err != nil {
				return rev, err
			}
			for _, p := range parsed {

				am, err := fAmounts(m.DefaultCurrency, strings.Split(p["amount"], ","))
				if err != nil {
					return rev, err
				}
				reward := &rewstruct.RewardAmount{
					Amounts:   am,
					Validator: wvc.ValidatorAddress,
				}
				if wvc.DelegatorAddress != p["receiver"] {
					rev.RewardRecipients = append(rev.RewardRecipients, p["receiver"])
				}
				rev.Rewards = append(rev.Rewards, reward)
			}
		}
		break
	}

	return rev, nil
}

// MsgUndelegate transforms distribution.MsgUndelegate sdk messages and related events to RewardTx
func (m *Mapper) MsgUndelegate(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgUndelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:         "MsgUndelegate",
		Delegator:    wvc.DelegatorAddress,
		ValidatorSrc: wvc.ValidatorAddress,
	}

	for _, e := range []string{"unbond", "coin_received"} {
	innerLoop:
		for _, ev := range lg.GetEvents() {
			if e == ev.GetType() {
				switch ev.GetType() {
				case "unbond":
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
					break innerLoop
				case "coin_received":
					parsed, err := m.groupRSEvents(lg.GetEvents())
					if err != nil {
						return rev, err
					}
					for _, p := range parsed {
						if m.NotBondedTokensPool != "" && p["receiver"] == m.NotBondedTokensPool {
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
						if wvc.DelegatorAddress != p["receiver"] {
							rev.RewardRecipients = append(rev.RewardRecipients, p["receiver"])
						}
						rev.Rewards = append(rev.Rewards, reward)
					}
					break innerLoop
				}
			}
		}
	}

	return rev, nil
}

// MsgDelegate transforms distribution.MsgDelegate sdk messages and related events to RewardTx
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

	for _, e := range []string{"delegate", "coin_received"} {
	innerLoop:
		for _, ev := range lg.GetEvents() {
			if e == ev.GetType() {
				switch ev.GetType() {
				case "delegate":
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
					break innerLoop
				case "coin_received":
					parsed, err := m.groupRSEvents(lg.GetEvents())
					if err != nil {
						return rev, err
					}
					for _, p := range parsed {
						if m.BondedTokensPool != "" && p["receiver"] == m.BondedTokensPool {
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
						if wvc.DelegatorAddress != p["receiver"] {
							rev.RewardRecipients = append(rev.RewardRecipients, p["receiver"])
						}
						rev.Rewards = append(rev.Rewards, reward)
					}
					break innerLoop
				}
			}
		}
	}

	return rev, nil
}

// MsgBeginRedelegate transforms distribution.MsgBeginRedelegate sdk messages and related events to RewardTx
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
	innerLoop:
		for _, ev := range lg.GetEvents() {
			if e == ev.GetType() {
				switch ev.GetType() {
				case "redelegate":
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
					break innerLoop
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
						/*	To be sure which transaction belongs to which validator, we need to check the delegation for the previous block.
							This code is temporary and may not work in some cases.
						*/
						if i%2 == 0 { // TODO in orde
							reward.Validator = wvc.ValidatorSrcAddress
						} else {
							reward.Validator = wvc.ValidatorDstAddress
						}
						if wvc.DelegatorAddress != p["receiver"] {
							rev.RewardRecipients = append(rev.RewardRecipients, p["receiver"])
						}
						rev.Rewards = append(rev.Rewards, reward)
					}
					break innerLoop
				}
			}
		}
	}

	return rev, nil
}

func toDec(numeric *big.Int, exp int32) types.Dec {
	if exp < 0 {
		return types.NewDecFromBigIntWithPrec(numeric, int64(-1*exp))
	}
	powerTen := big.NewInt(10)
	powerTen = powerTen.Exp(powerTen, big.NewInt(int64(exp)), nil)
	return types.NewDecFromBigIntWithPrec(powerTen.Mul(numeric, powerTen), 0)
}

// PostMsgBeginRedelegate takes a parsed RewardTx and establishes which
// validator should be assigned to which reward given the height - 1 reward balances (dels)
func (m *Mapper) PostMsgBeginRedelegate(rev *rewstruct.RewardTx, dels []cosmosgrpc.Delegators) (*rewstruct.RewardTx, error) {
	if rev.Type != "MsgBeginRedelegate" {
		return nil, fmt.Errorf("incorrect msg type: %s", rev.Type)
	}
	// usedValidatorAmounts to keep track if we used a particular validator
	// reward amount. this prevents the same validator being assigned to
	// two rewards of the same amount.
	usedValidatorAmounts := make(map[string]struct{})

	for _, rew := range rev.Rewards {
		srcDelta, srcExists, srcKey := types.NewDec(0), false, ""
		dstDelta, dstExists, dstKey := types.NewDec(0), false, ""

		for _, del := range dels {
			for _, unc := range del.Unclaimed {
				// assumption: every reward amount should exist
				// in dels (height - 1) if there was a claim,
				// so we can just check the first amount to
				// establish which validator the reward belongs
				// to.
				for _, amt := range unc.Unclaimed {
					if amt.Currency != rew.Amounts[0].Currency {
						continue
					}

					// if we already used once, we cannot use it again.
					key := fmt.Sprintf("%s:%s", unc.ValidatorAddress, amt.Numeric.String())
					if _, ok := usedValidatorAmounts[key]; ok {
						continue
					}

					ra := toDec(big.NewInt(0).SetBytes(rew.Amounts[0].Numeric), rew.Amounts[0].Exp)
					ua := toDec(amt.Numeric, amt.Exp)

					if rev.ValidatorSrc == unc.ValidatorAddress && !srcExists {
						srcDelta = ua.Sub(ra).Abs()
						srcExists = true
						srcKey = key
						break
					}

					if rev.ValidatorDst == unc.ValidatorAddress && !dstExists {
						dstDelta = ua.Sub(ra).Abs()
						dstExists = true
						dstKey = key
						break
					}
				}
			}
		}

		// neither a src or dst exist, this wasn't a reward.
		if !srcExists && !dstExists {
			return nil, fmt.Errorf("invalid reward, validator is neither src nor dst: %v", rew.Amounts[0])
		}

		if !srcExists {
			rew.Validator = rev.ValidatorDst
			usedValidatorAmounts[srcKey] = struct{}{}
			continue
		}

		if !dstExists {
			rew.Validator = rev.ValidatorSrc
			usedValidatorAmounts[dstKey] = struct{}{}
			continue
		}

		if srcDelta.LTE(dstDelta) {
			rew.Validator = rev.ValidatorSrc
			usedValidatorAmounts[srcKey] = struct{}{}
		} else {
			rew.Validator = rev.ValidatorDst
			usedValidatorAmounts[dstKey] = struct{}{}
		}
	}

	return rev, nil
}

// MsgEditValidator transforms distribution.MsgEditValidator sdk messages to RewardTx
func (m *Mapper) MsgEditValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgEditValidator{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:         "MsgEditValidator",
		ValidatorDst: wvc.ValidatorAddress,
	}

	return rev, nil
}

// MsgCreateValidator transforms distribution.MsgCreateValidator sdk messages to RewardTx
func (m *Mapper) MsgCreateValidator(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &staking.MsgCreateValidator{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:         "MsgCreateValidator",
		Delegator:    wvc.DelegatorAddress,
		ValidatorDst: wvc.ValidatorAddress,
	}

	return rev, nil
}

// MsgSetWithdrawAddress transforms distribution.MsgSetWithdrawAddress sdk messages to RewardTx
func (m *Mapper) MsgSetWithdrawAddress(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &distribution.MsgSetWithdrawAddress{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type:             "MsgSetWithdrawAddress",
		Delegator:        wvc.DelegatorAddress,
		RewardRecipients: []string{wvc.WithdrawAddress},
	}

	return rev, nil
}

// MsgFundCommunityPool transforms distribution.MsgFundCommunityPool sdk messages to RewardTx
func (m *Mapper) MsgFundCommunityPool(msg []byte, lg types.ABCIMessageLog) (rev *rewstruct.RewardTx, err error) {
	wvc := &distribution.MsgFundCommunityPool{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return rev, fmt.Errorf("not a distribution type: %w", err)
	}

	rev = &rewstruct.RewardTx{
		Type: "MsgFundCommunityPool",
	}

	return rev, nil
}

// defaultCurrency transforms amounts to Amount structure
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
	"unbond":              {"amount"},
	"redelegate":          {"amount"},
	"withdraw_commission": {"amount"},
	"withdraw_rewards":    {"amount"},
}

// eventsFilter filter events according to rewardEvents map
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
		}
	}

	return result, nil

}

// groupEvents group events into slice of maps
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
	if len(result) == 0 {
		return result, fmt.Errorf("missing events type: %s", etype)
	}

	return result, nil
}

// groupRSEvents group coin_received/coin_spent events into slice of maps
func (m *Mapper) groupRSEvents(ev types.StringEvents) (result []map[string]string, err error) {
	/*
		Both coin_received and coin_spent contains pairs "receiver", "amount" and "spender", "amount".
		The goal is to group these pairs. e.g.
		a = [[a:1], [b:1], [a:2], [b:2], [a:3], [b:3]]
		b = [[a:1], [c:1], [a:2], [c:2], [a:3], [c:3]]

		out = [{a:1,b:1 c:1}, {a:2,b:2, c:2}, {a:3, b:3, c:3}]
	*/
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
