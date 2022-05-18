package reward

import (
	"reflect"
	"sort"
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
		wantRev *rewstruct.Tx
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
			wantRev: &rewstruct.Tx{Type: "MsgDelegate", Sender: []string{"osmo1jm909k0z768sxacj05heqjc7c76ca43rstjxcs"}, ValidatorDst: "osmovaloper1gy0nyn2hscxxayj2pdyu8axmfvv75nnvhc079s", Delegator: "osmo1jm909k0z768sxacj05heqjc7c76ca43rstjxcs", Amounts: []*rewstruct.Amount{{Text: "120uosmo", Currency: "uosmo", Numeric: []byte("x")}}},
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
			wantRev: &rewstruct.Tx{Type: "MsgDelegate", ValidatorDst: "osmovaloper1de7qx00pz2j6gn9k88ntxxylelkazfk3llxw6r", Delegator: "osmo1de7qx00pz2j6gn9k88ntxxylelkazfk39gwddy", Amounts: []*rewstruct.Amount{{Text: "500000000uosmo", Currency: "uosmo", Numeric: []byte("\x1d\xcde\x00")}}},
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
			sort.Strings(tt.wantRev.Sender)
			sort.Strings(tt.wantRev.Recipient)
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
		wantRev *rewstruct.Tx
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
			wantRev: &rewstruct.Tx{Type: "MsgBeginRedelegate", Sender: []string{"osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, Recipient: []string{"osmo1faqfuwg5cfev205swpr53ewfjqzxn6e2e8u60f"}, Delegator: "osmo1faqfuwg5cfev205swpr53ewfjqzxn6e2e8u60f", ValidatorSrc: "osmovaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y4", ValidatorDst: "osmovaloper1t8qckan2yrygq7kl9apwhzfalwzgc2429p8f0s", Amounts: []*rewstruct.Amount{{Text: "30000000uosmo", Currency: "uosmo", Numeric: []byte("\x01\xc9À")}}, Rewards: []*rewstruct.Amount{{Text: "1059808uosmo", Currency: "uosmo", Numeric: []byte("\x10+\xe0")}}},
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
			sort.Strings(tt.wantRev.Sender)
			sort.Strings(tt.wantRev.Recipient)
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
		wantRev *rewstruct.Tx
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
			wantRev: &rewstruct.Tx{Type: "MsgWithdrawDelegatorReward", Sender: []string{"osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"}, Recipient: []string{"osmo14w50thy5jlank0c8vdzdlzehmf6hcjg89wkzmg"}, Delegator: "osmo14w50thy5jlank0c8vdzdlzehmf6hcjg89wkzmg", ValidatorSrc: "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n", Rewards: []*rewstruct.Amount{{Text: "113522uosmo", Currency: "uosmo", Numeric: []byte("\x01\xbbr")}}},
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
			sort.Strings(tt.wantRev.Sender)
			sort.Strings(tt.wantRev.Recipient)
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
		wantRev *rewstruct.Tx
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
			wantRev: &rewstruct.Tx{Type: "MsgUndelegate", Sender: []string{"osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld", "osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09"}, Recipient: []string{"osmo12d9lvpyfcq9cjcn6chty2yv27w22nye2qark47", "osmo1tygms3xhhs3yv487phx3dw4a95jn7t7lfqxwe3"}, Delegator: "osmo12d9lvpyfcq9cjcn6chty2yv27w22nye2qark47", ValidatorSrc: "osmovaloper12zwq8pcmmgwsl95rueqsf65avfg5zcj047ucw6", Amounts: []*rewstruct.Amount{{Text: "10100000uosmo", Currency: "uosmo", Numeric: []byte("\x9a\x1d ")}}, Rewards: []*rewstruct.Amount{{Text: "1ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", Currency: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", Numeric: []byte("\x01")}, {Text: "1373536uosmo", Currency: "uosmo", Numeric: []byte("\x14\xf5`")}}},
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
			sort.Strings(tt.wantRev.Sender)
			sort.Strings(tt.wantRev.Recipient)
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.MsgUndelegate() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}
