package reward

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/figment-networks/indexing-engine/proto/rewstruct"
	"go.uber.org/zap/zaptest"
)

func TestMapper_MsgDelegate_Osmosis(t *testing.T) {
	type args struct {
		msg []byte
		lg  types.ABCIMessageLog
	}
	tests := []struct {
		name    string
		args    args
		wantRev *rewstruct.RewardTx
		wantErr bool
	}{
		{
			name: "MsgDelegate_from_height_4284791",
			args: args{
				msg: []byte("\n+osmo1jm909k0z768sxacj05heqjc7c76ca43rstjxcs\x122osmovaloper1gy0nyn2hscxxayj2pdyu8axmfvv75nnvhc079s\x1a\f\n\x05uosmo\x12\x03120"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09"}, {Key: "amount", Value: "120uosmo"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "osmo1jm909k0z768sxacj05heqjc7c76ca43rstjxcs"}, {Key: "amount", Value: "120uosmo"}}},
					{Type: "delegate", Attributes: []types.Attribute{{Key: "validator", Value: "osmovaloper1gy0nyn2hscxxayj2pdyu8axmfvv75nnvhc079s"}, {Key: "amount", Value: "120uosmo"}, {Key: "new_shares", Value: "120.000000000000000000"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.staking.v1beta1.MsgDelegate"}, {Key: "module", Value: "staking"}, {Key: "sender", Value: "osmo1jm909k0z768sxacj05heqjc7c76ca43rstjxcs"}}},
				},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgDelegate",
				ValidatorDst: "osmovaloper1gy0nyn2hscxxayj2pdyu8axmfvv75nnvhc079s",
				Delegator:    "osmo1jm909k0z768sxacj05heqjc7c76ca43rstjxcs",
				Amounts:      []*rewstruct.Amount{{Text: "120uosmo", Currency: "uosmo", Numeric: []byte("x")}},
			},
		},
		{
			// test height 36 for MsgDelegate event backwards compatibility. Different events and amount units
			name: "MsgDelegate_from_height_36",
			args: args{
				msg: []byte("\n+osmo1de7qx00pz2j6gn9k88ntxxylelkazfk39gwddy\x122osmovaloper1de7qx00pz2j6gn9k88ntxxylelkazfk3llxw6r\x1a\x12\n\x05uosmo\x12\t500000000"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "delegate", Attributes: []types.Attribute{{Key: "validator", Value: "osmovaloper1de7qx00pz2j6gn9k88ntxxylelkazfk3llxw6r"}, {Key: "amount", Value: "500000000"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "delegate"}, {Key: "module", Value: "staking"}, {Key: "sender", Value: "osmo1de7qx00pz2j6gn9k88ntxxylelkazfk39gwddy"}}},
				},
				},
			},
			// TODO  amount issues / missing accounts
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgDelegate",
				ValidatorDst: "osmovaloper1de7qx00pz2j6gn9k88ntxxylelkazfk3llxw6r",
				Delegator:    "osmo1de7qx00pz2j6gn9k88ntxxylelkazfk39gwddy",
				Amounts:      []*rewstruct.Amount{{Text: "500000000uosmo", Currency: "uosmo", Numeric: []byte("\x1d\xcde\x00")}},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uosmo",
			}
			gotRev, err := m.MsgDelegate(tt.args.msg, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mapper.MsgDelegate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.MsgDelegate() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}

func TestMapper_MsgBeginRedelegate_Osmosis(t *testing.T) {
	type args struct {
		msg []byte
		lg  types.ABCIMessageLog
	}
	tests := []struct {
		name    string
		args    args
		wantRev *rewstruct.RewardTx
		wantErr bool
	}{
		{
			name: "MsgBeginRedelegate_from_height_4284791",
			args: args{
				msg: []byte("\n+osmo1faqfuwg5cfev205swpr53ewfjqzxn6e2e8u60f\x122osmovaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y4\x1a2osmovaloper1t8qckan2yrygq7kl9apwhzfalwzgc2429p8f0s\"\x11\n\x05uosmo\x12\b30000000"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "osmo1faqfuwg5cfev205swpr53ewfjqzxn6e2e8u60f"}, {Key: "amount", Value: "1059808uosmo"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "1059808uosmo"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.staking.v1beta1.MsgBeginRedelegate"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "module", Value: "staking"}, {Key: "sender", Value: "osmo1faqfuwg5cfev205swpr53ewfjqzxn6e2e8u60f"}}},
					{Type: "redelegate", Attributes: []types.Attribute{{Key: "source_validator", Value: "osmovaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y4"}, {Key: "destination_validator", Value: "osmovaloper1t8qckan2yrygq7kl9apwhzfalwzgc2429p8f0s"}, {Key: "amount", Value: "30000000uosmo"}, {Key: "completion_time", Value: "2022-05-20T12:47:35Z"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "osmo1faqfuwg5cfev205swpr53ewfjqzxn6e2e8u60f"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "1059808uosmo"}}},
				},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgBeginRedelegate",
				Delegator:    "osmo1faqfuwg5cfev205swpr53ewfjqzxn6e2e8u60f",
				ValidatorSrc: "osmovaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y4",
				ValidatorDst: "osmovaloper1t8qckan2yrygq7kl9apwhzfalwzgc2429p8f0s",
				Amounts:      []*rewstruct.Amount{{Text: "30000000uosmo", Currency: "uosmo", Numeric: []byte("\x01\xc9À")}},
				Rewards: []*rewstruct.RewardAmount{
					{Amounts: []*rewstruct.Amount{{Text: "1059808uosmo", Currency: "uosmo", Numeric: []byte("\x10+\xe0")}}, Validator: "osmovaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y4"},
				},
			}},
		{
			name: "MsgBeginRedelegate_from_height_3932419",
			args: args{
				msg: []byte("\n+osmo1m3esxfpc2qlk75jyjhfgyhma3std4cmvuwhrwt\x122osmovaloper12rzd5qr2wmpseypvkjl0spusts0eruw2g35lkn\x1a2osmovaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5w37wr9t\"\x0f\n\x05uosmo\x12\x06954871"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "osmo1m3esxfpc2qlk75jyjhfgyhma3std4cmvuwhrwt"}, {Key: "amount", Value: "3111uosmo"}, {Key: "receiver", Value: "osmo1m3esxfpc2qlk75jyjhfgyhma3std4cmvuwhrwt"}, {Key: "amount", Value: "3910uosmo"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "3111uosmo"}, {Key: "spender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "3910uosmo"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.staking.v1beta1.MsgBeginRedelegate"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "module", Value: "staking"}, {Key: "sender", Value: "osmo1m3esxfpc2qlk75jyjhfgyhma3std4cmvuwhrwt"}}},
					{Type: "redelegate", Attributes: []types.Attribute{{Key: "source_validator", Value: "osmovaloper12rzd5qr2wmpseypvkjl0spusts0eruw2g35lkn"}, {Key: "destination_validator", Value: "osmovaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5w37wr9t"}, {Key: "amount", Value: "954871uosmo"}, {Key: "completion_time", Value: "2022-04-23T15:12:24Z"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "osmo1m3esxfpc2qlk75jyjhfgyhma3std4cmvuwhrwt"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "3111uosmo"}, {Key: "recipient", Value: "osmo1m3esxfpc2qlk75jyjhfgyhma3std4cmvuwhrwt"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "3910uosmo"}}},
				},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgBeginRedelegate",
				Delegator:    "osmo1m3esxfpc2qlk75jyjhfgyhma3std4cmvuwhrwt",
				ValidatorSrc: "osmovaloper12rzd5qr2wmpseypvkjl0spusts0eruw2g35lkn",
				ValidatorDst: "osmovaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5w37wr9t",
				Amounts:      []*rewstruct.Amount{{Text: "954871uosmo", Currency: "uosmo", Numeric: []byte("\x0e\x91\xf7")}},
				// Rewards from each validator was checked by requesting delegations from height 3932418
				Rewards: []*rewstruct.RewardAmount{
					{Amounts: []*rewstruct.Amount{{Text: "3111uosmo", Currency: "uosmo", Numeric: []byte("\x0c'")}}, Validator: "osmovaloper12rzd5qr2wmpseypvkjl0spusts0eruw2g35lkn"},
					{Amounts: []*rewstruct.Amount{{Text: "3910uosmo", Currency: "uosmo", Numeric: []byte("\x0fF")}}, Validator: "osmovaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5w37wr9t"},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uosmo",
			}
			gotRev, err := m.MsgBeginRedelegate(tt.args.msg, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mapper.MsgBeginRedelegate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.MsgBeginRedelegate() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}

func TestMapper_MsgWithdrawDelegatorReward_Osmosis(t *testing.T) {
	type args struct {
		msg []byte
		lg  types.ABCIMessageLog
	}
	tests := []struct {
		name    string
		args    args
		wantRev *rewstruct.RewardTx
		wantErr bool
	}{
		{
			name: "MsgWithdrawDelegatorReward_from_height_4373725",
			args: args{
				msg: []byte("\n+osmo14w50thy5jlank0c8vdzdlzehmf6hcjg89wkzmg\x122osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "osmo14w50thy5jlank0c8vdzdlzehmf6hcjg89wkzmg"}, {Key: "amount", Value: "113522uosmo"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "113522uosmo"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "module", Value: "distribution"}, {Key: "sender", Value: "osmo14w50thy5jlank0c8vdzdlzehmf6hcjg89wkzmg"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "osmo14w50thy5jlank0c8vdzdlzehmf6hcjg89wkzmg"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "113522uosmo"}}},
					{Type: "withdraw_rewards", Attributes: []types.Attribute{{Key: "amount", Value: "113522uosmo"}, {Key: "validator", Value: "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n"}}},
				},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgWithdrawDelegatorReward",
				Delegator:    "osmo14w50thy5jlank0c8vdzdlzehmf6hcjg89wkzmg",
				ValidatorSrc: "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n",
				Rewards:      []*rewstruct.RewardAmount{{Amounts: []*rewstruct.Amount{{Text: "113522uosmo", Currency: "uosmo", Numeric: []byte("\x01\xbbr")}}, Validator: "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n"}},
			},
		},
		{
			// Modify Withdraw Address https://www.mintscan.io/osmosis/txs/01844C9906EC71D72EB50F2968D5ADCA4EFA4B861722E66A835AC532DC1101C3
			// Get Reward https://www.mintscan.io/osmosis/txs/C25F9C53231799E3A75729B2C08BE88BD1A8E88207C36BC82A7308BD2700CDB9
			name: "MsgWithdrawDelegatorReward_from_height_3951103_afer_Modify_Withdraw_Address_from_height_3878913",
			args: args{
				msg: []byte("\n+osmo1udv3dqftlrkc5yxqn9tnu7atyg8wnl4e4wm84a\x122osmovaloper1thsw3n94lzxy0knhss9n554zqp4dnfzx78j7sq"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "osmo153geppn7520n2kvsl850wxkz5r576n3wxnj0yj"}, {Key: "amount", Value: "1647849uosmo"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "1647849uosmo"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "module", Value: "distribution"}, {Key: "sender", Value: "osmo1udv3dqftlrkc5yxqn9tnu7atyg8wnl4e4wm84a"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "osmo153geppn7520n2kvsl850wxkz5r576n3wxnj0yj"}, {Key: "sender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "1647849uosmo"}}},
					{Type: "withdraw_rewards", Attributes: []types.Attribute{{Key: "amount", Value: "1647849uosmo"}, {Key: "validator", Value: "osmovaloper1thsw3n94lzxy0knhss9n554zqp4dnfzx78j7sq"}}},
				},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:            "MsgWithdrawDelegatorReward",
				Delegator:       "osmo1udv3dqftlrkc5yxqn9tnu7atyg8wnl4e4wm84a",
				ValidatorSrc:    "osmovaloper1thsw3n94lzxy0knhss9n554zqp4dnfzx78j7sq",
				Rewards:         []*rewstruct.RewardAmount{{Amounts: []*rewstruct.Amount{{Text: "1647849uosmo", Currency: "uosmo", Numeric: []byte("\x19$\xe9")}}, Validator: "osmovaloper1thsw3n94lzxy0knhss9n554zqp4dnfzx78j7sq"}},
				RewardRecipient: "osmo153geppn7520n2kvsl850wxkz5r576n3wxnj0yj",
			},
		},
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uosmo",
			}
			gotRev, err := m.MsgWithdrawDelegatorReward(tt.args.msg, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mapper.MsgWithdrawDelegatorReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.MsgWithdrawDelegatorReward() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}

func TestMapper_MsgUndelegate_Osmosis(t *testing.T) {
	type args struct {
		msg []byte
		lg  types.ABCIMessageLog
	}
	tests := []struct {
		name    string
		args    args
		wantRev *rewstruct.RewardTx
		wantErr bool
	}{
		{
			name: "MsgUndelegate_from_height_4373725",
			args: args{
				msg: []byte("\n+osmo12d9lvpyfcq9cjcn6chty2yv27w22nye2qark47\x122osmovaloper12zwq8pcmmgwsl95rueqsf65avfg5zcj047ucw6\x1a\x11\n\x05uosmo\x12\b10100000"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "osmo12d9lvpyfcq9cjcn6chty2yv27w22nye2qark47"}, {Key: "amount", Value: "1ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC,1373536uosmo"}, {Key: "receiver", Value: "osmo1tygms3xhhs3yv487phx3dw4a95jn7t7lfqxwe3"}, {Key: "amount", Value: "10100000uosmo"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, {Key: "amount", Value: "1ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC,1373536uosmo"}, {Key: "spender", Value: "osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09"}, {Key: "amount", Value: "10100000uosmo"}}},
					{Type: "unbond", Attributes: []types.Attribute{{Key: "validator", Value: "osmovaloper12zwq8pcmmgwsl95rueqsf65avfg5zcj047ucw6"}, {Key: "amount", Value: "10100000uosmo"}, {Key: "completion_time", Value: "2022-05-27T10:53:55Z"}}},
				},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgUndelegate",
				ValidatorSrc: "osmovaloper12zwq8pcmmgwsl95rueqsf65avfg5zcj047ucw6",
				Delegator:    "osmo12d9lvpyfcq9cjcn6chty2yv27w22nye2qark47",
				Amounts:      []*rewstruct.Amount{{Text: "10100000uosmo", Currency: "uosmo", Numeric: []byte("\x9a\x1d ")}},
				Rewards:      []*rewstruct.RewardAmount{{Amounts: []*rewstruct.Amount{{Text: "1ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", Currency: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", Numeric: []byte("\x01")}, {Text: "1373536uosmo", Currency: "uosmo", Numeric: []byte("\x14\xf5`")}}, Validator: "osmovaloper12zwq8pcmmgwsl95rueqsf65avfg5zcj047ucw6"}},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uosmo",
			}
			gotRev, err := m.MsgUndelegate(tt.args.msg, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mapper.MsgUndelegate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.MsgUndelegate() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}
