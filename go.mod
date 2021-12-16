module github.com/figment-networks/ni-cosmoslib

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.44.3
	github.com/cosmos/ibc-go v1.0.1
	github.com/figment-networks/indexing-engine v0.5.0
	github.com/gogo/protobuf v1.3.3
	github.com/gravity-devs/liquidity v1.4.2
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
