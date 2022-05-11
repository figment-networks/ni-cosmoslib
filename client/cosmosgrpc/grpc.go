package cosmosgrpc

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types/tx"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"google.golang.org/grpc"
)

type ClientConfig struct {
	TimeoutBlockCall    time.Duration
	TimeoutSearchTxCall time.Duration
}

type ErrorResolve interface {
	Check(err error) bool
}

type NOOPErrorResolve struct{}

func (ner *NOOPErrorResolve) Check(err error) bool {
	return false
}

// Client is a Tendermint RPC client for cosmos using figmentnetworks datahub
type Client struct {
	logger       *zap.Logger
	cfg          *ClientConfig
	errorResolve ErrorResolve

	// GRPC
	tmServiceClient    tmservice.ServiceClient
	txServiceClient    tx.ServiceClient
	rateLimiter        *rate.Limiter
	distributionClient distributionTypes.QueryClient
	stakingClient      stakingTypes.QueryClient
	bankClient         bankTypes.QueryClient
}

// NewClient returns a new client for a given endpoint
func NewClient(logger *zap.Logger, cli *grpc.ClientConn, cfg *ClientConfig) *Client {
	return &Client{
		logger:             logger,
		errorResolve:       &NOOPErrorResolve{},
		tmServiceClient:    tmservice.NewServiceClient(cli),
		txServiceClient:    tx.NewServiceClient(cli),
		distributionClient: distributionTypes.NewQueryClient(cli),
		stakingClient:      stakingTypes.NewQueryClient(cli),
		bankClient:         bankTypes.NewQueryClient(cli),
		cfg:                cfg,
	}
}

func (c *Client) SetRateLimitter(rateLimiter *rate.Limiter) {
	c.rateLimiter = rateLimiter
}

func (c *Client) SetErrorResolverLimitter(errorResolve ErrorResolve) {
	c.errorResolve = errorResolve
}
