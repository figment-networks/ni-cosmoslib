package reward

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/figment-networks/indexing-engine/proto/rewstruct"
	"github.com/figment-networks/ni-cosmoslib/client/cosmosgrpc"
	"go.uber.org/zap/zaptest"
)

func TestMapper_PostMsgBeginRedelegate(t *testing.T) {
	type args struct {
		tx   *rewstruct.RewardTx
		dels []cosmosgrpc.Delegators
	}
	tests := []struct {
		name    string
		args    args
		wantRev *rewstruct.RewardTx
		wantErr bool
	}{
		{
			name: "src_reward_only",
			args: args{
				tx: &rewstruct.RewardTx{
					Type:         "MsgBeginRedelegate",
					ValidatorSrc: "valsrc",
					ValidatorDst: "valdst",
					Rewards: []*rewstruct.RewardAmount{
						{
							Validator: "",
							Amounts: []*rewstruct.Amount{
								{
									Currency: "uatom",
									Numeric:  big.NewInt(25).Bytes(),
								},
							},
						},
					},
				},
				dels: []cosmosgrpc.Delegators{
					{
						Unclaimed: []cosmosgrpc.DelegatorsUnclaimed{{
							ValidatorAddress: "valsrc",
							Unclaimed: []cosmosgrpc.TransactionAmount{{
								Currency: "uatom",
								Numeric:  big.NewInt(25),
								Exp:      0,
							}},
						}},
					},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgBeginRedelegate",
				ValidatorSrc: "valsrc",
				ValidatorDst: "valdst",
				Rewards: []*rewstruct.RewardAmount{
					{
						Validator: "valsrc",
						Amounts: []*rewstruct.Amount{
							{
								Currency: "uatom",
								Numeric:  big.NewInt(25).Bytes(),
							},
						},
					},
				},
			},
		},
		{
			name: "dst_reward_only",
			args: args{
				tx: &rewstruct.RewardTx{
					Type:         "MsgBeginRedelegate",
					ValidatorSrc: "valsrc",
					ValidatorDst: "valdst",
					Rewards: []*rewstruct.RewardAmount{
						{
							Validator: "",
							Amounts: []*rewstruct.Amount{
								{
									Currency: "uatom",
									Numeric:  big.NewInt(25).Bytes(),
								},
							},
						},
					},
				},
				dels: []cosmosgrpc.Delegators{
					{
						Unclaimed: []cosmosgrpc.DelegatorsUnclaimed{{
							ValidatorAddress: "valdst",
							Unclaimed: []cosmosgrpc.TransactionAmount{{
								Currency: "uatom",
								Numeric:  big.NewInt(25),
								Exp:      0,
							}},
						}},
					},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgBeginRedelegate",
				ValidatorSrc: "valsrc",
				ValidatorDst: "valdst",
				Rewards: []*rewstruct.RewardAmount{
					{
						Validator: "valdst",
						Amounts: []*rewstruct.Amount{
							{
								Currency: "uatom",
								Numeric:  big.NewInt(25).Bytes(),
							},
						},
					},
				},
			},
		},
		{
			name: "src_dst_equal_amounts",
			args: args{
				tx: &rewstruct.RewardTx{
					Type:         "MsgBeginRedelegate",
					ValidatorSrc: "valsrc",
					ValidatorDst: "valdst",
					Rewards: []*rewstruct.RewardAmount{
						{
							Validator: "",
							Amounts: []*rewstruct.Amount{
								{
									Currency: "uatom",
									Numeric:  big.NewInt(25).Bytes(),
								},
							},
						},
						{
							Validator: "",
							Amounts: []*rewstruct.Amount{
								{
									Currency: "uatom",
									Numeric:  big.NewInt(25).Bytes(),
								},
							},
						},
					},
				},
				dels: []cosmosgrpc.Delegators{
					{
						Unclaimed: []cosmosgrpc.DelegatorsUnclaimed{
							{
								ValidatorAddress: "valsrc",
								Unclaimed: []cosmosgrpc.TransactionAmount{{
									Currency: "uatom",
									Numeric:  big.NewInt(25),
									Exp:      0,
								}},
							},
							{
								ValidatorAddress: "valdst",
								Unclaimed: []cosmosgrpc.TransactionAmount{{
									Currency: "uatom",
									Numeric:  big.NewInt(25),
									Exp:      0,
								}},
							},
						},
					},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgBeginRedelegate",
				ValidatorSrc: "valsrc",
				ValidatorDst: "valdst",
				Rewards: []*rewstruct.RewardAmount{
					{
						Validator: "valsrc",
						Amounts: []*rewstruct.Amount{
							{
								Currency: "uatom",
								Numeric:  big.NewInt(25).Bytes(),
							},
						},
					},
					{
						Validator: "valdst",
						Amounts: []*rewstruct.Amount{
							{
								Currency: "uatom",
								Numeric:  big.NewInt(25).Bytes(),
							},
						},
					},
				},
			},
		},
		{
			name: "src_dst_negative_subtraction",
			args: args{
				tx: &rewstruct.RewardTx{
					Type:         "MsgBeginRedelegate",
					ValidatorSrc: "valsrc",
					ValidatorDst: "valdst",
					Rewards: []*rewstruct.RewardAmount{
						{
							Validator: "",
							Amounts: []*rewstruct.Amount{
								{
									Currency: "uatom",
									Numeric:  big.NewInt(50).Bytes(),
								},
							},
						},
						{
							Validator: "",
							Amounts: []*rewstruct.Amount{
								{
									Currency: "uatom",
									Numeric:  big.NewInt(25).Bytes(),
								},
							},
						},
					},
				},
				dels: []cosmosgrpc.Delegators{
					{
						Unclaimed: []cosmosgrpc.DelegatorsUnclaimed{
							{
								ValidatorAddress: "valsrc",
								Unclaimed: []cosmosgrpc.TransactionAmount{{
									Currency: "uatom",
									Numeric:  big.NewInt(50),
									Exp:      0,
								}},
							},
							{
								ValidatorAddress: "valdst",
								Unclaimed: []cosmosgrpc.TransactionAmount{{
									Currency: "uatom",
									Numeric:  big.NewInt(25),
									Exp:      0,
								}},
							},
						},
					},
				},
			},
			wantRev: &rewstruct.RewardTx{
				Type:         "MsgBeginRedelegate",
				ValidatorSrc: "valsrc",
				ValidatorDst: "valdst",
				Rewards: []*rewstruct.RewardAmount{
					{
						Validator: "valsrc",
						Amounts: []*rewstruct.Amount{
							{
								Currency: "uatom",
								Numeric:  big.NewInt(50).Bytes(),
							},
						},
					},
					{
						Validator: "valdst",
						Amounts: []*rewstruct.Amount{
							{
								Currency: "uatom",
								Numeric:  big.NewInt(25).Bytes(),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mapper{
				Logger:          zaptest.NewLogger(t),
				DefaultCurrency: "uatom",
			}
			gotRev, err := m.PostMsgBeginRedelegate(tt.args.tx, tt.args.dels)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mapper.PostMsgBeginRedelegate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRev, tt.wantRev) {
				t.Errorf("Mapper.PostMsgBeginRedelegate() = %v, want %v", gotRev, tt.wantRev)
			}
		})
	}
}
