package cosmosgrpc

import (
	"math/big"
	"time"
)

type Validator struct {
	OperatorAddress string
	Jailed          bool
	Status          string

	Tokens *big.Int

	DelegatorShares *big.Int

	Description     ValidatorDescription
	UnbondingHeight int64
	UnbondingTime   time.Time

	Commission        Commission
	MinSelfDelegation *big.Int

	Rewards []TransactionAmount
}

type ValidatorDescription struct {
	Moniker         string
	Identity        string
	Website         string
	SecurityContact string
	Details         string
}

type Commission struct {
	Rate          *big.Int
	MaxRate       *big.Int
	MaxChangeRate *big.Int
	UpdateTime    time.Time
}

type DelegationResponse struct {
	Delegation Delegation
	Balance    TransactionAmount
}

type Delegation struct {
	DelegatorAddress string
	ValidatorAddress string
	Shares           *big.Int // dec
}

type Balance struct {
	DelegatorAddress string
	ValidatorAddress string
}

type Delegators struct {
	DelegatorAddress string
	Unclaimed        []DelegatorsUnclaimed
}

type DelegatorsUnclaimed struct {
	ValidatorAddress string
	Unclaimed        []TransactionAmount
}

// TransactionAmount structure holding amount information with decimal implementation (numeric * 10 ^ exp)
type TransactionAmount struct {
	// Textual representation of Amount
	Text string `json:"text,omitempty"`
	// The currency in what amount is returned (if applies)
	Currency string `json:"currency,omitempty"`

	// Numeric part of the amount
	Numeric *big.Int `json:"numeric,omitempty"`
	// Exponential part of amount obviously 0 by default
	Exp int32 `json:"exp,omitempty"`
}
