package mapper

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/figment-networks/indexing-engine/structs"
	"github.com/figment-networks/ni-cosmoslib/figment/api/util"

	"github.com/cosmos/cosmos-sdk/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
)

// delegate undelegate redelegate, -> addresses
// delegate undelegate redelegate + withdraw delegator rewards -> delagator rewards
// withdraw validator commision -> validator rewards

// unbound address cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgWithdrawValidatorCommission sdk messages to SubsetEvent
func (mapper *Mapper) MsgWithdrawValidatorCommission(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &distribution.MsgWithdrawValidatorCommission{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	// TODO  MsgSetWithdrawAddress ???
	// https://docs.cosmos.network/master/modules/distribution/04_messages.html

	// mapped fields
	rev.ValidatorSrcAddress = wvc.ValidatorAddress
	// "cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7"
	// https://atomscan.com/transactions/0108F7BCF51BA2CF0BFFA9D5DDBCB963D0EF6C6C57707A95ADB81ED3DE689F86
	// 2022/05/02 16:51:20 type coin_received attr receiver cosmos1dwq55ln00s7lv72jnxk2zdvz9cdz2yh9vdf85d
	// 2022/05/02 16:51:20 type coin_received attr amount 74302uatom
	// 2022/05/02 16:51:20 type coin_spent attr spender cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
	// 2022/05/02 16:51:20 type coin_spent attr amount 74302uatom
	// 2022/05/02 16:51:20 type message attr action /cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission
	// 2022/05/02 16:51:20 type message attr sender cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
	// 2022/05/02 16:51:20 type message attr module distribution
	// 2022/05/02 16:51:20 type message attr sender cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7
	// 2022/05/02 16:51:20 type transfer attr recipient cosmos1dwq55ln00s7lv72jnxk2zdvz9cdz2yh9vdf85d
	// 2022/05/02 16:51:20 type transfer attr sender cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
	// 2022/05/02 16:51:20 type transfer attr amount 74302uatom
	// 2022/05/02 16:51:20 type withdraw_commission attr amount 74302uatom

	for _, ev := range lg.GetEvents() {
		for _, attr := range ev.GetAttributes() {
			log.Println("type", ev.GetType(), "attr", attr.Key, attr.Value)
		}

	}
	log.Println("dupa")
	return nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgWithdrawDelegatorReward sdk messages to SubsetEvent
func (mapper *Mapper) MsgWithdrawDelegatorReward(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &distribution.MsgWithdrawDelegatorReward{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	// mapped fields
	rev.DelegatorAddress = wvc.DelegatorAddress
	rev.ValidatorSrcAddress = wvc.ValidatorAddress

	// TODO add amont/reward
	for _, ev := range lg.GetEvents() {
		for _, attr := range ev.GetAttributes() {
			log.Println("type", ev.GetType(), "attr", attr.Key, attr.Value)
			if ev.GetType() == "withdraw_rewards" && attr.Key == "amount" {
				fAmounts("reward", strings.Split(attr.Value, ","), rev)
			}
		}
	}

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

	// https://atomscan.com/blocks/10337615
	// ValidatorSrcAddress "cosmosvaloper1hjadhj9nqzpye2vkmkz4thahhd0z8dh3udhq74"
	// DelegatorAddress "cosmos1w2zw9ngef2zwhssmafpaew9czkxxtluf3shv5t"
	// Amounts "60176517uatom"
	// Rewards "156082uatom"

	// coin "receiver"  "cosmos1w2zw9ngef2zwhssmafpaew9czkxxtluf3shv5t" got "156082uatom"
	// coin "receiver"  "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r" got  "60176517uatom"
	// coin "spender"  "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"  spent "156082uatom"
	// coin "spender" "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh" spent "60176517uatom"

	// unbond "validator" "cosmosvaloper1hjadhj9nqzpye2vkmkz4thahhd0z8dh3udhq74"
	// unbond "amount" "60176517uatom"
	// unbond "completion_time" "2022-05-23T13:44:49Z"

	// transfer "recipient" "cosmos1w2zw9ngef2zwhssmafpaew9czkxxtluf3shv5t"
	// transfer "sender" "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"
	// transfer "amount" "156082uatom"

	// transfer "recipient" "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r"
	// transfer "sender" "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
	// transfer "amount" "60176516uatom"

	log.Print("------ wvc", wvc)
	for _, ev := range lg.GetEvents() {
		attr := ev.GetAttributes()
		if ev.GetType() == "coin_received" {
			log.Println("coin_received", len(attr)/2)
			for i := 0; i < len(attr); i = i + 2 {
				log.Println("pair?", i, attr[i].Key, attr[i].Value, attr[i+1].Key, attr[i+1].Value)
				if attr[i].Key == "receiver" && attr[i+1].Key == "amount" {
					if attr[i].Value == rev.DelegatorAddress {
						fAmounts("reward", strings.Split(attr[i+1].Value, ","), rev)
					} else {
						fAmounts("amount", strings.Split(attr[i+1].Value, ","), rev)
					}
				}
			}
			log.Println("amount")
		}
		log.Println("amount33")
	}

	return nil
}

// DistributionWithdrawValidatorCommissionToSub transforms distribution.MsgDelegate sdk messages to SubsetEvent
func (mapper *Mapper) MsgDelegate(msg []byte, lg types.ABCIMessageLog, rev *structs.RewardEvent) (err error) {
	wvc := &staking.MsgDelegate{}
	if err := proto.Unmarshal(msg, wvc); err != nil {
		return fmt.Errorf("not a distribution type: %w", err)
	}

	// mapped fields
	rev.DelegatorAddress = wvc.DelegatorAddress
	rev.ValidatorDstAddress = wvc.ValidatorAddress

	// debug :
	// https://atomscan.com/transactions/71A644FF56F3A06048DC621F16B05FB948CC5AD1F8FD0DB629B0EDE2689F9709

	// ValidatorDstAddress "cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc"
	// DelegatorAddress "cosmos14nh2dy06lvfh4fm65z5k4l44h5cwggefj70y3w"
	// Amounts "65000000uatom"
	// Rewards "2021065uatom"
	// coin "receiver"  "cosmos14nh2dy06lvfh4fm65z5k4l44h5cwggefj70y3w" got "2021065uatom"
	// coin "receiver"  "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh" got  "65000000uatom"
	// coin "spender"  "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"  spent "2021065uatom"
	// coin "spender" "cosmos14nh2dy06lvfh4fm65z5k4l44h5cwggefj70y3w" spent "65000000uatom"
	// deletage "validator" "cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc"
	// deletage "amount" "65000000uatom"
	// delegate "new_shares" "65000000uatom"
	// transfer "recipient" "cosmos14nh2dy06lvfh4fm65z5k4l44h5cwggefj70y3w"
	// transfer "sender" "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"
	// transfer "amount" "2021065uatom"

	log.Print("------ wvc", wvc)
	for _, ev := range lg.GetEvents() {
		attr := ev.GetAttributes()
		if ev.GetType() == "coin_received" {
			log.Println("coin_received", len(attr)/2)
			for i := 0; i < len(attr); i = i + 2 {
				log.Println("pair?", i, attr[i].Key, attr[i].Value, attr[i+1].Key, attr[i+1].Value)
				if attr[i].Key == "receiver" && attr[i+1].Key == "amount" {
					if attr[i].Value == rev.DelegatorAddress {
						fAmounts("reward", strings.Split(attr[i+1].Value, ","), rev)
					} else {
						fAmounts("amount", strings.Split(attr[i+1].Value, ","), rev)
					}
				}
			}
			log.Println("amount")
		}
		log.Println("amount33")
	}

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

	// debug
	// https://atomscan.com/blocks/10337936
	// https://atomscan.com/transactions/3D9B7799A8CFD627903E201B7F1290217D8872C497C1AB270056DB8D6E7588A7

	// ValidatorSrcAddress "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c"
	// ValidatorDstAddress "cosmosvaloper17mggn4znyeyg25wd7498qxl7r2jhgue8u4qjcq"
	// DelegatorAddress "cosmos18d66t5p3x2tf5mjml58uchp22n3axuaxlh082l"
	// Amounts 221000000uatom
	// Rewards 115uatom ????

	// 2022/05/02 16:21:38 type coin_received attr receiver cosmos18d66t5p3x2tf5mjml58uchp22n3axuaxlh082l
	// 2022/05/02 16:21:38 type coin_received attr amount 115uatom
	// 2022/05/02 16:21:38 type coin_spent attr spender cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
	// 2022/05/02 16:21:38 type coin_spent attr amount 115uatom
	// 2022/05/02 16:21:38 type message attr action /cosmos.staking.v1beta1.MsgBeginRedelegate
	// 2022/05/02 16:21:38 type message attr sender cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
	// 2022/05/02 16:21:38 type message attr module staking
	// 2022/05/02 16:21:38 type message attr sender cosmos18d66t5p3x2tf5mjml58uchp22n3axuaxlh082l
	// 2022/05/02 16:21:38 type redelegate attr source_validator cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c
	// 2022/05/02 16:21:38 type redelegate attr destination_validator cosmosvaloper17mggn4znyeyg25wd7498qxl7r2jhgue8u4qjcq
	// 2022/05/02 16:21:38 type redelegate attr amount 221000000uatom
	// 2022/05/02 16:21:38 type redelegate attr completion_time 2022-05-23T14:21:28Z
	// 2022/05/02 16:21:38 type transfer attr recipient cosmos18d66t5p3x2tf5mjml58uchp22n3axuaxlh082l
	// 2022/05/02 16:21:38 type transfer attr sender cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
	// 2022/05/02 16:21:38 type transfer attr amount 115uatom

	log.Print("------ wvc", wvc)
	for _, ev := range lg.GetEvents() {

		if ev.GetType() == "redelegate" {
			for _, attr := range ev.GetAttributes() {
				if attr.Key == "amount" {
					fAmounts("amount", strings.Split(attr.Value, ","), rev)
				}
			}

		}
		log.Println("amount33")
	}
	log.Println("end")
	return nil
}

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
