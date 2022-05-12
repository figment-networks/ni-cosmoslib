module github.com/figment-networks/ni-cosmoslib/api

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.44.3
	github.com/figment-networks/indexing-engine v0.5.0
	github.com/gogo/protobuf v1.3.3
	github.com/gravity-devs/liquidity v1.4.2
	github.com/tendermint/tendermint v0.34.14
	go.uber.org/zap v1.17.0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
