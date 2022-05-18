package reward

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/figment-networks/indexing-engine/proto/rewstruct"
	"go.uber.org/zap/zaptest"
)

// "MsgWithdrawDelegatorReward"  10510022, 10510043
// "MsgWithdrawValidatorCommission"  10510010, 10510025
// "MsgDelegate" 10510041
// "MsgSetWithdrawAddress" 10510043, 10510066
// "MsgUndelegate" 10510043 10510063
// "MsgBeginRedelegate" 10511859 10511900

func TestMapper_MsgDelegate_Cosmos(t *testing.T) {
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
			name: "MsgDelegate_from_height_10510041",
			args: args{
				msg: []byte("\n-cosmos1mc0mxsdgsyjepsqetw3w5a459zj64k7akuhdu4\x124cosmosvaloper157v7tczs40axfgejp2m43kwuzqe0wsy0rv8puv\x1a\x0f\n\x05uatom\x12\x06140000"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "cosmos1mc0mxsdgsyjepsqetw3w5a459zj64k7akuhdu4"}, {Key: "amount", Value: "25uatom"}, {Key: "receiver", Value: "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"}, {Key: "amount", Value: "140000uatom"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "25uatom"}, {Key: "spender", Value: "cosmos1mc0mxsdgsyjepsqetw3w5a459zj64k7akuhdu4"}, {Key: "amount", Value: "140000uatom"}}},
					{Type: "delegate", Attributes: []types.Attribute{{Key: "validator", Value: "cosmosvaloper157v7tczs40axfgejp2m43kwuzqe0wsy0rv8puv"}, {Key: "amount", Value: "140000uatom"}, {Key: "new_shares", Value: "140000.000000000000000000"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.staking.v1beta1.MsgDelegate"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "module", Value: "staking"}, {Key: "sender", Value: "cosmos1mc0mxsdgsyjepsqetw3w5a459zj64k7akuhdu4"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "cosmos1mc0mxsdgsyjepsqetw3w5a459zj64k7akuhdu4"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "25uatom"}}},
				},
				},
			},
			wantRev: &rewstruct.Tx{Type: "MsgDelegate", Sender: []string{"cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl", "cosmos1mc0mxsdgsyjepsqetw3w5a459zj64k7akuhdu4"}, ValidatorDst: "cosmosvaloper157v7tczs40axfgejp2m43kwuzqe0wsy0rv8puv", Delegator: "cosmos1mc0mxsdgsyjepsqetw3w5a459zj64k7akuhdu4", Amounts: []*rewstruct.Amount{{Text: "140000uatom", Currency: "uatom", Numeric: []byte("\x02\"\xe0")}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uatom",
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

func TestMapper_MsgBeginRedelegate_Cosmos(t *testing.T) {
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
			name: "MsgBeginRedelegate_from_height_10511859",
			args: args{
				msg: []byte("\n-cosmos1gv5vf68d4rfww9v2lg568vut36d5eth39rgvmh\x124cosmosvaloper132juzk0gdmwuxvx4phug7m3ymyatxlh9734g4w\x1a4cosmosvaloper16k579jk6yt2cwmqx9dz5xvq9fug2tekvlu9qdv\"\x10\n\x05uatom\x12\a1099780"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "cosmos1gv5vf68d4rfww9v2lg568vut36d5eth39rgvmh"}, {Key: "amount", Value: "32800uatom"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "32800uatom"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.staking.v1beta1.MsgBeginRedelegate"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "module", Value: "staking"}, {Key: "sender", Value: "cosmos1gv5vf68d4rfww9v2lg568vut36d5eth39rgvmh"}}},
					{Type: "redelegate", Attributes: []types.Attribute{{Key: "source_validator", Value: "cosmosvaloper132juzk0gdmwuxvx4phug7m3ymyatxlh9734g4w"}, {Key: "destination_validator", Value: "cosmosvaloper16k579jk6yt2cwmqx9dz5xvq9fug2tekvlu9qdv"}, {Key: "amount", Value: "1099780uatom"}, {Key: "completion_time", Value: "2022-06-06T14:08:52Z"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "cosmos1gv5vf68d4rfww9v2lg568vut36d5eth39rgvmh"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "32800uatom"}}},
				},
				},
			},
			wantRev: &rewstruct.Tx{Type: "MsgBeginRedelegate", Sender: []string{"cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, Recipient: []string{"cosmos1gv5vf68d4rfww9v2lg568vut36d5eth39rgvmh"}, Delegator: "cosmos1gv5vf68d4rfww9v2lg568vut36d5eth39rgvmh", ValidatorSrc: "cosmosvaloper132juzk0gdmwuxvx4phug7m3ymyatxlh9734g4w", ValidatorDst: "cosmosvaloper16k579jk6yt2cwmqx9dz5xvq9fug2tekvlu9qdv", Amounts: []*rewstruct.Amount{{Text: "1099780uatom", Currency: "uatom", Numeric: []byte("\x10\xc8\x04")}}, Rewards: []*rewstruct.Amount{{Text: "32800uatom", Currency: "uatom", Numeric: []byte("\x80 ")}}},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uatom",
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

func TestMapper_MsgWithdrawDelegatorReward_Cosmos(t *testing.T) {
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
			name: "MsgWithdrawDelegatorReward_from_height_10510043",
			args: args{
				msg: []byte("\n-cosmos1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2hsucxnd\x124cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"}, {Key: "sender", Value: "cosmos1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2hsucxnd"}, {Key: "module", Value: "distribution"}}},
					{Type: "withdraw_rewards", Attributes: []types.Attribute{{Key: "amount", Value: ""}, {Key: "validator", Value: "cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7"}}},
				},
				},
			},
			wantRev: &rewstruct.Tx{Type: "MsgWithdrawDelegatorReward", Delegator: "cosmos1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2hsucxnd", ValidatorSrc: "cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7"},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uatom",
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

func TestMapper_MsgSetWithdrawAddress_Cosmos(t *testing.T) {
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
			name: "MsgSetWithdrawAddress_from_height_10510025",
			args: args{
				msg: []byte("\n-cosmos1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2hsucxnd\x12-cosmos1z8wrnv35mmezpseym0jy7lngvsan2alwn8gma9"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress"}, {Key: "sender", Value: "cosmos1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2hsucxnd"}, {Key: "module", Value: "distribution"}}},
					{Type: "set_withdraw_address", Attributes: []types.Attribute{{Key: "withdraw_address", Value: "cosmos1z8wrnv35mmezpseym0jy7lngvsan2alwn8gma9"}}},
				},
				},
			},
			wantRev: &rewstruct.Tx{Type: "MsgSetWithdrawAddress", Delegator: "cosmos1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2hsucxnd"},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uatom",
			}
			gotRev, err := m.MsgSetWithdrawAddress(tt.args.msg, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mapper.MsgSetWithdrawAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.MsgSetWithdrawAddress() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}

func TestMapper_MsgUndelegate_Cosmos(t *testing.T) {
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
			name: "MsgUndelegate_from_height_10510043",
			args: args{
				msg: []byte("\n-cosmos1q7h3kuuvnzd3fq4snhls543uvv8stt9em0nmkf\x124cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3\x1a\x12\n\x05uatom\x12\t200000000"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "cosmos1q7h3kuuvnzd3fq4snhls543uvv8stt9em0nmkf"}, {Key: "amount", Value: "11628006uatom"}, {Key: "receiver", Value: "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r"}, {Key: "amount", Value: "200000000uatom"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "11628006uatom"}, {Key: "spender", Value: "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"}, {Key: "amount", Value: "200000000uatom"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.staking.v1beta1.MsgUndelegate"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "sender", Value: "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"}, {Key: "module", Value: "staking"}, {Key: "sender", Value: "cosmos1q7h3kuuvnzd3fq4snhls543uvv8stt9em0nmkf"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "cosmos1q7h3kuuvnzd3fq4snhls543uvv8stt9em0nmkf"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "11628006uatom"}, {Key: "recipient", Value: "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r"}, {Key: "sender", Value: "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"}, {Key: "amount", Value: "200000000uatom"}}},
					{Type: "unbond", Attributes: []types.Attribute{{Key: "validator", Value: "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"}, {Key: "amount", Value: "200000000uatom"}, {Key: "completion_time", Value: "2022-06-06T10:41:55Z"}}},
				},
				},
			},
			wantRev: &rewstruct.Tx{Type: "MsgUndelegate", Sender: []string{"cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl", "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"}, Recipient: []string{"cosmos1q7h3kuuvnzd3fq4snhls543uvv8stt9em0nmkf", "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r"}, Delegator: "cosmos1q7h3kuuvnzd3fq4snhls543uvv8stt9em0nmkf", ValidatorSrc: "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3", Amounts: []*rewstruct.Amount{{Text: "200000000uatom", Currency: "uatom", Numeric: []byte("\x0b\xeb\xc2\x00")}}, Rewards: []*rewstruct.Amount{{Text: "11628006uatom", Currency: "uatom", Numeric: []byte("\xb1m\xe6")}}},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uatom",
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

func TestMapper_MsgWithdrawValidatorCommission_Cosmos(t *testing.T) {
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
			name: "MsgWithdrawValidatorCommission_from_height_10510025",
			args: args{
				msg: []byte("\n4cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7"),
				lg: types.ABCIMessageLog{MsgIndex: 0, Log: "", Events: []types.StringEvent{
					{Type: "coin_received", Attributes: []types.Attribute{{Key: "receiver", Value: "cosmos1z8wrnv35mmezpseym0jy7lngvsan2alwn8gma9"}, {Key: "amount", Value: "36370uatom"}}},
					{Type: "coin_spent", Attributes: []types.Attribute{{Key: "spender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "36370uatom"}}},
					{Type: "message", Attributes: []types.Attribute{{Key: "action", Value: "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "module", Value: "distribution"}, {Key: "sender", Value: "cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7"}}},
					{Type: "transfer", Attributes: []types.Attribute{{Key: "recipient", Value: "cosmos1z8wrnv35mmezpseym0jy7lngvsan2alwn8gma9"}, {Key: "sender", Value: "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, {Key: "amount", Value: "36370uatom"}}},
					{Type: "withdraw_commission", Attributes: []types.Attribute{{Key: "amount", Value: "36370uatom"}}},
				},
				},
			},
			wantRev: &rewstruct.Tx{Type: "MsgWithdrawValidatorCommission", Sender: []string{"cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl"}, ValidatorDst: "cosmosvaloper1hvsdf03tl6w5pnfvfv5g8uphjd4wfw2h4gvnl7", Amounts: []*rewstruct.Amount{{Text: "36370uatom", Currency: "uatom", Numeric: []byte("\x8e\x12")}}},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uatom",
			}
			gotRev, err := m.MsgWithdrawValidatorCommission(tt.args.msg, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mapper.MsgWithdrawValidatorCommission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.MsgWithdrawValidatorCommission() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}
