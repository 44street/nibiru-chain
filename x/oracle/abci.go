package oracle

import (
	"time"

	"github.com/NibiruChain/nibiru/x/oracle/keeper"
	"github.com/NibiruChain/nibiru/x/oracle/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	params := k.GetParams(ctx)
	if types.IsPeriodLastBlock(ctx, params.VotePeriod) {
		UpdateExchangeRates(ctx, k, params)
	}

	// Do slash who did miss voting over threshold and
	// reset miss counters of all validators at the last block of slash window
	if types.IsPeriodLastBlock(ctx, params.SlashWindow) {
		k.SlashAndResetMissCounters(ctx)
	}
}

// TODO(mercilex): move this logic for consistency on the keeper as SlashAndResetMissCountersLogic is kept there
func UpdateExchangeRates(ctx sdk.Context, k keeper.Keeper, params types.Params) {
	k.Logger(ctx).Info("processing validator price votes")
	// Build claim map over all validators in active set
	validatorClaimMap := make(map[string]types.Claim)

	maxValidators := k.StakingKeeper.MaxValidators(ctx)
	iterator := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
	defer iterator.Close()

	powerReduction := k.StakingKeeper.PowerReduction(ctx)

	i := 0
	for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
		validator := k.StakingKeeper.Validator(ctx, iterator.Value())

		// Exclude not bonded validator
		if validator.IsBonded() {
			valAddr := validator.GetOperator()
			validatorClaimMap[valAddr.String()] = types.NewClaim(validator.GetConsensusPower(powerReduction), 0, 0, valAddr)
			i++
		}
	}

	pairsMap := make(map[string]struct{})
	k.IteratePairs(ctx, func(pair string) bool {
		pairsMap[pair] = struct{}{}
		return false
	})

	// Clear all exchange rates
	k.IterateExchangeRates(ctx, func(pair string, _ sdk.Dec) (stop bool) {
		k.DeleteExchangeRate(ctx, pair)
		return false
	})

	// Organize votes to ballot by pair
	// NOTE: **Filter out inactive or jailed validators**
	// NOTE: **Make abstain votes to have zero vote power**
	pairBallotMap := k.OrganizeBallotByPair(ctx, validatorClaimMap)

	if referencePair := PickReferencePair(ctx, k, pairsMap, pairBallotMap); referencePair != "" {
		// make voteMap of reference pair to calculate cross exchange rates
		referenceBallot := pairBallotMap[referencePair]
		referenceValidatorExchangeRateMap := referenceBallot.ToMap()
		referenceExchangeRate := referenceBallot.WeightedMedianWithAssertion()

		// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
		for pair, ballot := range pairBallotMap {
			// Convert ballot to cross exchange rates
			if pair != referencePair {
				ballot = ballot.ToCrossRateWithSort(referenceValidatorExchangeRateMap)
			}

			// Get weighted median of cross exchange rates
			exchangeRate := Tally(ctx, ballot, params.RewardBand, validatorClaimMap)

			// Transform into the original exchange rate
			if pair != referencePair {
				exchangeRate = referenceExchangeRate.Quo(exchangeRate)
			}

			// Set the exchange rate, emit ABCI event
			k.SetExchangeRateWithEvent(ctx, pair, exchangeRate)
		}
	}

	//---------------------------
	// Do miss counting & slashing
	voteTargetsLen := len(pairsMap)
	for _, claim := range validatorClaimMap {
		// Skip abstain & valid voters
		if int(claim.WinCount) == voteTargetsLen {
			continue
		}

		// Increase miss counter
		k.SetMissCounter(ctx, claim.Recipient, k.GetMissCounter(ctx, claim.Recipient)+1)
		k.Logger(ctx).Info("vote miss", "validator", claim.Recipient.String())
	}

	// Distribute rewards to ballot winners
	k.RewardBallotWinners(
		ctx,
		(int64)(params.VotePeriod),
		(int64)(params.RewardDistributionWindow),
		pairsMap,
		validatorClaimMap,
	)

	// Clear the ballot
	k.ClearBallots(ctx, params.VotePeriod)

	// Update vote targets and tobin tax
	k.ApplyWhitelist(ctx, params.Whitelist, pairsMap)
}
