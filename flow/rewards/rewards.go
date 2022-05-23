package rewards

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
	"time"

	ttypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/figment-networks/indexing-engine/proto/datastore"
	"github.com/figment-networks/indexing-engine/proto/rewstruct"
	"github.com/figment-networks/indexing-engine/structs"

	pb "github.com/figment-networks/indexing-engine/proto/datastore"

	"github.com/figment-networks/ni-cosmoslib/client/cosmosgrpc"

	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
)

var (
	ErrNoRows = errors.New("No records")
)

type RewardsNonce struct {
	Height   uint64
	Sequence uint64
}

type Crossing struct {
	Height   uint64
	Sequence uint64
}

func (c *Crossing) GetHeight() uint64 {
	return c.Height
}

func (c *Crossing) GetSequence() uint64 {
	return c.Sequence
}

type HeightTime struct {
	Height uint64
	Time   time.Time
}

type HeightError struct {
	Height uint64
	Error  error
}

type DelegatorOP int64

const (
	DelegatorOPAdd DelegatorOP = iota
	DelegatorOPRemove
	DelegatorOPBoth
)

type DelegatorValidator struct {
	Op        DelegatorOP
	Delegator string
	Validator string
	Amounts   []*rewstruct.Amount
}

type RewardProducer interface {
	GetRewards(*rewstruct.Tx) []structs.ClaimedReward
	GetDelegations(tx *rewstruct.Tx) (accounts []DelegatorValidator)
	MapTransactions(txs []*tx.Tx, txResponses []*types.TxResponse, t time.Time) (retTxs *rewstruct.Txs, err error)
}

type Client interface {
	GetBlock(ctx context.Context, height uint64) (block *ttypes.Block, blockID *ttypes.BlockID, err error)
	GetRawTxs(ctx context.Context, height uint64, perPage uint64) (txs []*tx.Tx, txResponses []*types.TxResponse, err error)

	GetHeightValidators(ctx context.Context, height, limit, page uint64) (vals []cosmosgrpc.Validator, err error)
	GetDelegators(ctx context.Context, height uint64, operatorAddress string, limit, page uint64) (vals []cosmosgrpc.DelegationResponse, err error)
	GetDelegatorDelegations(ctx context.Context, height uint64, delegatorAddress string, limit, page uint64) (vals []cosmosgrpc.DelegationResponse, err error)
	GetDelegations(ctx context.Context, height uint64, delegatorAddress string) (dels []cosmosgrpc.Delegators, err error)
}

type RewardsExtractionConfig struct {
	ValidatorFetchPage uint64
	DelegatorFetchPage uint64
	DatastorePrefix    string
	ChainID            string
	Network            string
}

type RewardsExtraction struct {
	logger *zap.Logger
	Cfg    RewardsExtractionConfig

	client   Client
	dsClient pb.DatastoreServiceClient

	orp RewardProducer
}

func NewRewardsExtraction(logger *zap.Logger, cfg RewardsExtractionConfig, client Client, dsClient pb.DatastoreServiceClient, orp RewardProducer) *RewardsExtraction {
	return &RewardsExtraction{
		client:   client,
		dsClient: dsClient,
		orp:      orp,
		logger:   logger,
		Cfg:      cfg,
	}
}

func (re *RewardsExtraction) FetchHeights(ctx context.Context, startHeight, endHeight, sequence uint64) (h structs.Heights, crossingHeights []structs.Crossing, err error) {
	const fetchTxWorkersNumber = 24

	txStream, err := re.dsClient.StoreRecords(ctx)
	if err != nil {
		return h, crossingHeights, fmt.Errorf("Error initializing fetchHeight store stream %w", err)
	}

	heights := make(chan HeightTime, fetchTxWorkersNumber)
	resp := make(chan HeightError, fetchTxWorkersNumber+1)
	for i := 0; i < fetchTxWorkersNumber; i++ {
		go re.fetchHeightData(ctx, heights, resp, txStream)
	}

	var (
		counter int
		sent    int
		gErr    error
	)

FETCH_HEIGHTS_LOOP:
	for height := startHeight; height < endHeight+1; height++ {
		block, _, err := re.client.GetBlock(ctx, height)
		if err != nil {
			h.ErrorAt = append(h.ErrorAt, height)
			gErr = err
			break FETCH_HEIGHTS_LOOP
		}

		timeID := uint64(math.Floor(float64(block.Header.Time.Truncate(time.Hour).Unix()) / 3600))
		if sequence != timeID {
			crossingHeights = append(crossingHeights, &Crossing{
				Height:   height,
				Sequence: timeID,
			})
			sequence = timeID
		}

		ht := HeightTime{Height: height, Time: block.Header.Time}
		select {
		case heights <- ht:
		case r := <-resp:
			counter++
			if r.Error != nil {
				h.ErrorAt = append(h.ErrorAt, r.Height)
				gErr = r.Error
				break FETCH_HEIGHTS_LOOP
			}
			h.Heights = append(h.Heights, r.Height)
			if h.LatestData.LastHeight < r.Height {
				h.LatestData.LastHeight = r.Height
				h.LatestData.LastMark = r.Height
			}

			// and schedule the next one
			heights <- ht
		}
		sent++
	}
	close(heights)

	if sent > 0 && counter != sent {
		// read remainder if any
		for r := range resp {
			counter++
			if r.Error != nil {
				h.ErrorAt = append(h.ErrorAt, r.Height)
				gErr = r.Error
			} else {
				h.Heights = append(h.Heights, r.Height)
				if h.LatestData.LastHeight < r.Height {
					h.LatestData.LastHeight = r.Height
					h.LatestData.LastMark = r.Height
				}
			}

			if counter == sent {
				break
			}
		}
	}
	close(resp)

	acks, err := txStream.CloseAndRecv()
	if err != nil {
		return h, crossingHeights, fmt.Errorf("error closing transactions stream %w", err)
	}
	for _, ack := range acks.Acks {
		if ack.Error != "" {
			return h, crossingHeights, fmt.Errorf("error closing transactions stream (acks) %s", ack.Error)
		}
	}

	if gErr != nil {
		return h, crossingHeights, gErr
	}

	return h, crossingHeights, nil
}

func (re *RewardsExtraction) CalculateRewards(ctx context.Context, height, sequence uint64) error {
	// previous full hour accounts
	accountData, err := re.dsClient.FetchRecord(ctx, &datastore.FetchRecordRequest{
		Type:     re.Cfg.DatastorePrefix + "account_records",
		Sequence: sequence - 1,
	})
	if err != nil {
		return err
	}
	if accountData.Error != "" {
		if accountData.Error != ErrNoRows.Error() {
			return fmt.Errorf("Error getting heights: %s", accountData.Error)
		}

		accountDataInitial, err := re.dsClient.FetchRecord(ctx, &datastore.FetchRecordRequest{
			Type:     re.Cfg.DatastorePrefix + "account_records",
			Sequence: sequence,
		})
		if err != nil {
			return err
		}

		if accountDataInitial.Error != ErrNoRows.Error() {
			re.logger.Warn("Fetched Initial accounts for: ", zap.Uint64("height", height), zap.Uint64("sequence", sequence))
			return nil
		}

		re.logger.Warn("Fetching Initial accounts for: ", zap.Uint64("height", height))
		acc, err := re.fetchInitialAccounts(ctx, height)
		if err != nil {
			return fmt.Errorf("Error fetching initial accounts: %w", err)
		}
		b, err := json.Marshal(AccountsHeight{Accounts: acc, Height: height, Sequence: sequence})
		if err != nil {
			return err
		}
		ack, err := re.dsClient.StoreRecord(ctx, &datastore.Payload{
			Type:     re.Cfg.DatastorePrefix + "account_records",
			Sequence: sequence,
			Content:  b,
		})
		if err != nil {
			return err
		}
		if ack.Error != "" {
			return fmt.Errorf("Error storing record: %s ", ack.Error)
		}
		// support this better
		return nil
	}

	ah := &AccountsHeight{}
	err = json.Unmarshal(accountData.Content, ah)
	if err != nil {
		return err
	}

	// Fetch Rewards from previous sequence
	rewardsRaw, err := re.dsClient.FetchRecord(ctx, &datastore.FetchRecordRequest{
		Type:     re.Cfg.DatastorePrefix + "reward_records",
		Sequence: sequence - 1,
	})
	if err != nil {
		return fmt.Errorf("Error while fetching records: %w ", err)
	}
	if rewardsRaw.Error != "" && rewardsRaw.Error != ErrNoRows.Error() {
		return fmt.Errorf("Error fetching records: %s ", rewardsRaw.Error)
	}

	previousdelegs := &rewstruct.Delegators{}
	if rewardsRaw.Error != ErrNoRows.Error() {
		if err = proto.Unmarshal(rewardsRaw.Content, previousdelegs); err != nil {
			return err
		}
	} else {
		previousdelegs.Height = ah.Height
	}

	delegatorClaims, delegationDiff, err := re.fetchTransactions(ctx, re.orp, previousdelegs.Height+1, uint32(height-previousdelegs.Height))
	if err != nil {
		return err
	}
	newAccounts := accountData.Content
	if len(delegationDiff) > 0 {
		for _, dv := range delegationDiff {
			if dv.Op == DelegatorOPRemove {
				// Delete non-delegators. Check if delagator is delegating to any validator
				// in n+1 block. If not - remove
				del, err := re.client.GetDelegations(ctx, height+1, dv.Delegator)
				if err != nil {
					return err
				}
				if len(del) == 0 {
					delete(ah.Accounts, dv.Delegator)
				}

				continue
			}
			ah.Accounts[dv.Delegator] = struct{}{}
		}
		newAccounts, err = json.Marshal(ah)
		if err != nil {
			return err
		}
	}

	ack, err := re.dsClient.StoreRecord(ctx, &datastore.Payload{
		Type:     re.Cfg.DatastorePrefix + "account_records",
		Sequence: sequence,
		Content:  newAccounts,
	})
	if err != nil {
		return err
	}
	if ack.Error != "" {
		return fmt.Errorf("Error storing account_records: %s", ack.Error)
	}

	newdelegs, err := re.fetchHeightUnclaimedRewards(ctx, height, sequence, ah.Accounts)
	if err != nil {
		return err
	}

	finalEarned := &rewstruct.Rewards{
		ChainId:  re.Cfg.ChainID,
		Network:  re.Cfg.Network,
		Sequence: sequence,
		Time:     &rewstruct.Timestamp{Seconds: int64(sequence * 60 * 60)},
		Height:   height,
		Grouping: "1h",
	}
	finalEarned.Earned = calculate(previousdelegs, newdelegs, delegatorClaims)
	for _, dc := range delegatorClaims {
		finalEarned.Claimed = append(finalEarned.Claimed, mapClaims(dc)...)
	}

	finalRewards, err := proto.Marshal(finalEarned)
	if err != nil {
		return err
	}

	ack, err = re.dsClient.StoreRecord(ctx, &datastore.Payload{
		Type:     re.Cfg.DatastorePrefix + "earned_reward_records",
		Sequence: sequence,
		Content:  finalRewards,
	})

	if err != nil {
		return err
	}
	if ack.Error != "" {
		return errors.New(ack.Error)
	}

	return nil
}

func (re *RewardsExtraction) fetchHeightData(ctx context.Context, heights chan HeightTime, resp chan HeightError, txStream datastore.DatastoreService_StoreRecordsClient) {
	for {
		select {
		case height, ok := <-heights:
			if !ok {
				return
			}

			/*validators, err := re.client.GetHeightValidators(ctx, height.Height, 0, 100)
			if err != nil {
				errs <- err
				return
			}

			vals, err := json.Marshal(validators) // TODO(l): revisit this encoding, proto?
			if err != nil {
				errs <- err
				return
			}
			log.Println("3 ", time.Since(now))
			err = vStream.Send(&datastore.Payload{
				Type:     re.datastorePrefix + "validator_records",
				Sequence: height.Height,
				Content:  vals,
			})
			if err != nil {
				errs <- err
				return
			}
			*/

			r := HeightError{Height: height.Height}

			rawTxs, rewTxResps, err := re.client.GetRawTxs(ctx, height.Height, 100)
			if err != nil {
				r.Error = fmt.Errorf("error getting raw tx (%d): %w ", height.Height, err)
				resp <- r
				return
			}
			txs, err := re.orp.MapTransactions(rawTxs, rewTxResps, height.Time)
			if err != nil {
				r.Error = fmt.Errorf("error mapping transaction  (%d): %w ", height.Height, err)
				resp <- r
				return
			}
			txr, err := proto.Marshal(txs)
			if err != nil {
				r.Error = fmt.Errorf("error marshaling transaction  (%d): %w ", height.Height, err)
				resp <- r
				return
			}

			// every height - even empty one has to be written
			err = txStream.Send(&datastore.Payload{
				Type:     re.Cfg.DatastorePrefix + "tx_records",
				Sequence: height.Height,
				Content:  txr,
			})
			if err != nil {
				r.Error = fmt.Errorf("error storing transaction  (%d): %w ", height.Height, err)
			}
			resp <- r
		}
	}
}

func (re *RewardsExtraction) fetchHeightUnclaimedRewards(ctx context.Context, height, sequence uint64, accounts map[string]interface{}) (newdelegs *rewstruct.Delegators, err error) {
	processing := make(chan string, 100)
	outp := make(chan DelegateResponse, 100)
	defer close(outp)

	nctx, cancel := context.WithCancel(ctx)
	for i := 0; i < 20; i++ {
		go re.UnclaimedFetcher(nctx, height, processing, outp)
	}
	// populate all the accounts regardless of the error,
	// upon failure UnclaimedFetcher would just pass through
	// it's simpler than locking, counters or whatever.
	// so below we always expect len(accounts) replies
	go populateAccounts(processing, accounts)

	newdelegs = &rewstruct.Delegators{
		Height:     height,
		Delegators: make(map[string]*rewstruct.ValidatorsUnclaimed),
	}

	var (
		counter int
		dErr    error
	)
	for delegation := range outp {
		counter++
		if delegation.Err != nil {
			dErr = delegation.Err
			cancel()
			if counter == len(accounts) {
				break
			}
			continue
		}

		addRewards(newdelegs, delegation)

		if counter%100 == 0 {
			re.logger.Debug("Rewards for: ", zap.Int("rewards", counter))
		}

		if counter == len(accounts) {
			break
		}
	}
	cancel()

	if dErr != nil {
		re.logger.Error("Delegation error", zap.Error(dErr))
		return nil, fmt.Errorf("error getting delegation %w", dErr)
	}

	rewardsEncoded, err := proto.Marshal(newdelegs)
	if err != nil {
		return nil, err
	}

	ack, err := re.dsClient.StoreRecord(ctx, &datastore.Payload{
		Type:     re.Cfg.DatastorePrefix + "reward_records",
		Sequence: sequence,
		Content:  rewardsEncoded,
	})
	if err != nil {
		return nil, err
	}

	if ack.Error != "" {
		return nil, fmt.Errorf("error storing data: %s", ack.Error)
	}

	return newdelegs, err
}

func (re *RewardsExtraction) fetchTransactions(ctx context.Context, rp RewardProducer, startheight uint64, limit uint32) (claims map[string][]structs.ClaimedReward, accounts []DelegatorValidator, err error) {
	re.logger.Debug("processing fetchTransactions", zap.Uint64("start_height", startheight))
	// Fetch Rewards from previous sequence
	recordRewards, err := re.dsClient.FetchRecords(ctx, &datastore.DataRequest{
		Type:     re.Cfg.DatastorePrefix + "tx_records",
		Sequence: startheight,
		Limit:    limit,
	})
	if err != nil {
		return nil, nil, err
	}
	var (
		drp     *datastore.DataResponsePayload
		lastSeq uint64
	)

	for {
		drp, err = recordRewards.Recv()
		if err != nil {
			if err == io.EOF {
				// clear out EOF before returning.  so the caller doesn't consider this an error.
				err = nil
				break
			}
			return nil, nil, fmt.Errorf("Error receiving repordReward data: %w", err)
		}
		lastSeq = drp.Sequence

		if drp.Error != "" {
			return nil, nil, fmt.Errorf("Error in receive repordReward data payload: %w", err)
		}
		if drp.Sequence > 0 && drp.Content == nil {
			// this means record it present but it's empty so we don't process it
			continue
		}
		txs := &rewstruct.Txs{}
		if err = proto.Unmarshal(drp.Content, txs); err != nil {
			return nil, nil, err
		}

		for _, tx := range txs.Txs {
			for _, reward := range rp.GetRewards(tx) {
				if claims == nil {
					claims = make(map[string][]structs.ClaimedReward)
				}
				a := claims[reward.Account]
				a = append(a, reward)
				claims[reward.Account] = a
			}

			accounts = append(accounts, rp.GetDelegations(tx)...)
		}
	}
	// last request has to be current height, otherwise we cannot use it
	if lastSeq != startheight+uint64(limit)-1 {
		return nil, nil, errors.New("data is not fully persisted")
	}

	return claims, accounts, err
}

func addRewards(newdelegs *rewstruct.Delegators, delegation DelegateResponse) {
	for _, del := range delegation.Dels {
		valUncl, ok := newdelegs.Delegators[del.DelegatorAddress]
		if !ok {
			valUncl = &rewstruct.ValidatorsUnclaimed{
				Amounts: make(map[string]*rewstruct.UnclaimedDenoms),
			}
		}
		for _, uncl := range del.Unclaimed {
			unclr, ok := valUncl.Amounts[uncl.ValidatorAddress]
			if !ok {
				unclr = &rewstruct.UnclaimedDenoms{
					Amount: make(map[string]*rewstruct.Amount),
				}
			}
			for _, den := range uncl.Unclaimed {
				d, ok := unclr.Amount[den.Currency]
				if !ok {
					d = &rewstruct.Amount{
						Text:     den.Text,
						Currency: den.Currency,
						Exp:      den.Exp,
						Numeric:  den.Numeric.Bytes(),
					}
				} else { // Assumes normalization
					a := big.NewInt(0).SetBytes(d.Numeric)
					d.Numeric = a.Add(a, den.Numeric).Bytes()
				}
				unclr.Amount[den.Currency] = d
			}
			valUncl.Amounts[uncl.ValidatorAddress] = unclr
		}
		newdelegs.Delegators[del.DelegatorAddress] = valUncl
	}
}

func calculate(previousdelegs *rewstruct.Delegators, newdelegs *rewstruct.Delegators, claims map[string][]structs.ClaimedReward) (earnedRewards []*rewstruct.SimpleReward) {
	earnedRewards = []*rewstruct.SimpleReward{}
	claimedDelegators := make(map[string]struct{})
	for d := range newdelegs.Delegators {
		earned := calculateInternal(d, previousdelegs.Delegators[d], newdelegs.Delegators[d], claims[d])
		earnedRewards = append(earnedRewards, earned...)
		claimedDelegators[d] = struct{}{}
	}
	for d := range claims {
		if _, ok := claimedDelegators[d]; ok {
			continue
		}
		xnewdel := &rewstruct.ValidatorsUnclaimed{}
		earned := calculateInternal(d, previousdelegs.Delegators[d], xnewdel, claims[d])
		earnedRewards = append(earnedRewards, earned...)
	}

	return earnedRewards
}

// calculate earned rewards for a particular delegator
func calculateInternal(delegatorAddress string, previousdelegs *rewstruct.ValidatorsUnclaimed, newdelegs *rewstruct.ValidatorsUnclaimed, claims []structs.ClaimedReward) (earnedRewards []*rewstruct.SimpleReward) {
	if previousdelegs == nil {
		previousdelegs = &rewstruct.ValidatorsUnclaimed{
			Amounts: make(map[string]*rewstruct.UnclaimedDenoms),
		}
	}

	diff := make(map[string]map[string]structs.TransactionAmount)

	// diff is initially just the newdelegs amounts
	for va, newUncl := range newdelegs.Amounts {
		diff[va] = make(map[string]structs.TransactionAmount)
		for currency, newAmount := range newUncl.Amount {
			na := big.NewInt(0).SetBytes(newAmount.Numeric)
			if _, ok := diff[va]; !ok {
				diff[va] = make(map[string]structs.TransactionAmount)
			}
			diff[va][currency] = structs.TransactionAmount{
				Currency: newAmount.Currency,
				Exp:      newAmount.Exp,
				Numeric:  na,
			}
		}
	}

	for va, newUncl := range newdelegs.Amounts {
		prevUncl, ok := previousdelegs.Amounts[va]
		if !ok {
			// if the previousdelegs amounts don't exist then the
			// amounts are all just the newdelegs amounts as is, ie
			// newdelegs amount - 0.
			continue
		}

		for currency, newAmount := range newUncl.Amount {
			prevAmount, ok := prevUncl.Amount[currency]
			if !ok {
				// if the previous currency doesn't exist, then
				// the amount is just the new amount.
				continue
			}
			// use the cosmos decimal type to subtract here since
			// delegator rewards are returned as a fractional
			// amount in the base unit. Ie 5.1334 uatom
			na := toDec(big.NewInt(0).SetBytes(newAmount.Numeric), newAmount.Exp)
			pa := toDec(big.NewInt(0).SetBytes(prevAmount.Numeric), prevAmount.Exp)
			if _, ok := diff[va]; !ok {
				diff[va] = make(map[string]structs.TransactionAmount)
			}
			ra := na.Sub(pa).BigInt()
			diff[va][currency] = structs.TransactionAmount{
				Currency: newAmount.Currency,
				// after using a dec for calculating the exp becomes -18
				Exp:     -1 * types.Precision,
				Numeric: ra,
			}
		}

		// deal with currencies that exist in the previous amount, but
		// not the new amount, in that case it's 0 - previous amount,
		// so negate.
		for currency, prevAmount := range prevUncl.Amount {
			_, ok := newUncl.Amount[currency]
			if ok {
				continue
			}
			pa := big.NewInt(0).SetBytes(prevAmount.Numeric)
			if _, ok := diff[va]; !ok {
				diff[va] = make(map[string]structs.TransactionAmount)
			}
			ra := pa.Neg(pa)
			diff[va][currency] = structs.TransactionAmount{
				Currency: prevAmount.Currency,
				Exp:      prevAmount.Exp,
				Numeric:  ra,
			}
		}
	}

	// deal with previousdelegs amounts that don't exist in newdelegs. set
	// all amounts to 0 - previous amount.
	for va, prevUncl := range previousdelegs.Amounts {
		_, ok := newdelegs.Amounts[va]
		if ok {
			continue
		}
		if _, ok := diff[va]; !ok {
			diff[va] = make(map[string]structs.TransactionAmount)
		}
		for currency, prevAmount := range prevUncl.Amount {
			pa := big.NewInt(0).SetBytes(prevAmount.Numeric)
			ra := pa.Neg(pa)
			diff[va][currency] = structs.TransactionAmount{
				Currency: prevAmount.Currency,
				Exp:      prevAmount.Exp,
				Numeric:  ra,
			}
		}
	}

	// add claims
	for _, c := range claims {
		va := c.Validator
		_, ok := diff[va]
		if !ok {
			diff[va] = make(map[string]structs.TransactionAmount)
		}
		for _, r := range c.ClaimedReward {
			currency := r.Currency
			amount, ok := diff[va][currency]
			if !ok {
				amount = structs.TransactionAmount{
					Numeric:  big.NewInt(0),
					Currency: r.Currency,
					Exp:      r.Exp,
				}
				diff[va][currency] = amount
			}
			a := toDec(amount.Numeric, amount.Exp)
			ra := toDec(r.Numeric, r.Exp)
			fra := a.Add(ra).BigInt()
			diff[va][currency] = structs.TransactionAmount{
				Currency: amount.Currency,
				// after using a dec for calculating the exp becomes -18
				Exp:     -1 * types.Precision,
				Numeric: fra,
			}
		}
	}

	// generate rewards
	for va, amounts := range diff {
		earnedAmounts := []*rewstruct.Amount{}
		for _, amount := range amounts {
			earnedAmounts = append(earnedAmounts, &rewstruct.Amount{
				Text: fmt.Sprintf("%s%s", strings.TrimRight(
					strings.TrimRight(toDec(amount.Numeric, amount.Exp).String(), "0"), "."),
					amount.Currency),
				Currency: amount.Currency,
				Numeric:  amount.Numeric.Bytes(),
				Exp:      amount.Exp,
			})
		}
		earnedRewards = append(earnedRewards, &rewstruct.SimpleReward{
			Account:   delegatorAddress,
			Validator: va,
			Amounts:   earnedAmounts,
		})
	}

	return earnedRewards
}

func toDec(numeric *big.Int, exp int32) types.Dec {
	if exp < 0 {
		return types.NewDecFromBigIntWithPrec(numeric, int64(-1*exp))
	}
	powerTen := big.NewInt(10)
	powerTen = powerTen.Exp(powerTen, big.NewInt(int64(exp)), nil)
	return types.NewDecFromBigIntWithPrec(powerTen.Mul(numeric, powerTen), 0)
}

func populateAccounts(processing chan<- string, accounts map[string]interface{}) {
	defer close(processing)
	for k := range accounts {
		processing <- k
	}
}

type DelegateResponse struct {
	Dels []cosmosgrpc.Delegators
	Err  error
}

var errorThreshold = 5

func (re *RewardsExtraction) UnclaimedFetcher(ctx context.Context, height uint64, in <-chan string, out chan<- DelegateResponse) {
	for address := range in {
		select { // on error passthrough all the unread messages
		case <-ctx.Done():
			out <- DelegateResponse{nil, errors.New("closed after error")}
			continue
		default:
		}

		var consecutiveErrors int
		del, err := re.client.GetDelegations(ctx, height, address)
		if err != nil {
		REPEATLOOP:
			for {
				del, err = re.client.GetDelegations(ctx, height, address)
				if err == nil {
					break REPEATLOOP
				}
				consecutiveErrors++
				if consecutiveErrors >= errorThreshold {
					break REPEATLOOP
				}
				<-time.After(1 * time.Second)
			}
		}
		out <- DelegateResponse{del, err}
	}
}

type AccountsHeight struct {
	Sequence uint64
	Height   uint64
	Accounts map[string]interface{}
}

func (re *RewardsExtraction) fetchInitialAccounts(ctx context.Context, height uint64) (accounts map[string]interface{}, err error) {
	validators, err := re.client.GetHeightValidators(ctx, height, 0, re.Cfg.ValidatorFetchPage)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator lists %w", err)
	}

	accountsMap := make(map[string]interface{})
	a := struct{}{}
	for _, v := range validators {
		deleg, err := re.client.GetDelegators(ctx, height, v.OperatorAddress, 0, re.Cfg.DelegatorFetchPage)
		if err != nil {
			// GetHeightValidators can produce Delegators not at this height
			// this is ok b/c we are just trying to fetch an initial list at this height.
			if strings.Contains(err.Error(), "validator does not exist") {
				continue
			}
			return accounts, fmt.Errorf("Error getting delegators %w", err)
		}
		for _, d := range deleg {
			accountsMap[d.Delegation.DelegatorAddress] = a // dedupe
		}
	}
	return accountsMap, nil
}

func mapClaims(claims []structs.ClaimedReward) (rews []*rewstruct.SimpleReward) {
	for _, c := range claims {
		r := &rewstruct.SimpleReward{
			Account:   c.Account,
			Validator: c.Validator,
			Height:    c.Mark,
			Time:      &rewstruct.Timestamp{Seconds: c.Time.Unix(), Nanos: int32(c.Time.Nanosecond())},
		}
		for _, a := range c.ClaimedReward {
			r.Amounts = append(r.Amounts, &rewstruct.Amount{
				Text:     a.Text,
				Currency: a.Currency,
				Exp:      a.Exp,
				Numeric:  a.Numeric.Bytes(),
			})
		}

		rews = append(rews, r)
	}
	return rews
}
