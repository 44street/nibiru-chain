package keeper_test

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/NibiruChain/nibiru/x/oracle/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/nibiru/x/oracle"
	"github.com/NibiruChain/nibiru/x/oracle/types"
)

func TestOracleThreshold(t *testing.T) {
	exchangeRates := types.ExchangeRateTuples{
		{
			Pair:         common.PairBTCStable.String(),
			ExchangeRate: randomExchangeRate,
		},
	}
	input, h := setup(t)
	exchangeRateStr, err := exchangeRates.ToString()
	require.NoError(t, err)

	// Case 1.
	// Less than the threshold signs, exchange rate consensus fails
	salt := "1"
	hash := types.GetAggregateVoteHash(salt, exchangeRateStr, keeper.ValAddrs[0])
	prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, keeper.Addrs[0], keeper.ValAddrs[0])
	voteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, keeper.Addrs[0], keeper.ValAddrs[0])

	_, err1 := h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(0)), prevoteMsg)
	_, err2 := h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(1)), voteMsg)
	require.NoError(t, err1)
	require.NoError(t, err2)

	oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)

	_, err = input.OracleKeeper.GetExchangeRate(input.Ctx.WithBlockHeight(1), exchangeRates[0].Pair)
	require.Error(t, err)

	// Case 2.
	// More than the threshold signs, exchange rate consensus succeeds
	salt = "1"
	hash = types.GetAggregateVoteHash(salt, exchangeRateStr, keeper.ValAddrs[0])
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, keeper.Addrs[0], keeper.ValAddrs[0])
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, keeper.Addrs[0], keeper.ValAddrs[0])

	_, err1 = h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(0)), prevoteMsg)
	_, err2 = h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(1)), voteMsg)
	require.NoError(t, err1)
	require.NoError(t, err2)

	salt = "2"
	hash = types.GetAggregateVoteHash(salt, exchangeRateStr, keeper.ValAddrs[1])
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, keeper.Addrs[1], keeper.ValAddrs[1])
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, keeper.Addrs[1], keeper.ValAddrs[1])

	_, err1 = h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(0)), prevoteMsg)
	_, err2 = h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(1)), voteMsg)
	require.NoError(t, err1)
	require.NoError(t, err2)

	salt = "3"
	hash = types.GetAggregateVoteHash(salt, exchangeRateStr, keeper.ValAddrs[2])
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, keeper.Addrs[2], keeper.ValAddrs[2])
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, keeper.Addrs[2], keeper.ValAddrs[2])

	_, err1 = h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(0)), prevoteMsg)
	_, err2 = h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(1)), voteMsg)
	require.NoError(t, err1)
	require.NoError(t, err2)

	oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)

	rate, err := input.OracleKeeper.GetExchangeRate(input.Ctx.WithBlockHeight(1), exchangeRates[0].Pair)
	require.NoError(t, err)
	require.Equal(t, randomExchangeRate, rate)

	// Case 3.
	// Increase voting power of absent validator, exchange rate consensus fails
	val, _ := input.StakingKeeper.GetValidator(input.Ctx, keeper.ValAddrs[2])
	_, _ = input.StakingKeeper.Delegate(input.Ctx.WithBlockHeight(0), keeper.Addrs[2], stakingAmt.MulRaw(3), stakingtypes.Unbonded, val, false)

	salt = "1"
	hash = types.GetAggregateVoteHash(salt, exchangeRateStr, keeper.ValAddrs[0])
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, keeper.Addrs[0], keeper.ValAddrs[0])
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, keeper.Addrs[0], keeper.ValAddrs[0])

	_, err1 = h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(0)), prevoteMsg)
	_, err2 = h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(1)), voteMsg)
	require.NoError(t, err1)
	require.NoError(t, err2)

	salt = "2"
	hash = types.GetAggregateVoteHash(salt, exchangeRateStr, keeper.ValAddrs[1])
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, keeper.Addrs[1], keeper.ValAddrs[1])
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, keeper.Addrs[1], keeper.ValAddrs[1])

	_, err1 = h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(0)), prevoteMsg)
	_, err2 = h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(1)), voteMsg)
	require.NoError(t, err1)
	require.NoError(t, err2)

	oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)

	_, err = input.OracleKeeper.GetExchangeRate(input.Ctx.WithBlockHeight(1), exchangeRates[0].Pair)
	require.Error(t, err)
}

func TestOracleDrop(t *testing.T) {
	input, h := setup(t)

	input.OracleKeeper.SetExchangeRate(input.Ctx, common.PairGovStable.String(), randomExchangeRate)

	// Account 1, pair gov stable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 0)

	// Immediately swap halt after an illiquid oracle vote
	oracle.EndBlocker(input.Ctx, input.OracleKeeper)

	_, err := input.OracleKeeper.GetExchangeRate(input.Ctx, common.PairGovStable.String())
	require.Error(t, err)
}

func TestOracleTally(t *testing.T) {
	input, _ := setup(t)

	ballot := types.ExchangeRateBallot{}
	rates, valAddrs, stakingKeeper := types.GenerateRandomTestCase()
	input.OracleKeeper.StakingKeeper = stakingKeeper
	h := keeper.NewMsgServerImpl(input.OracleKeeper)
	for i, rate := range rates {
		decExchangeRate := sdk.NewDecWithPrec(int64(rate*math.Pow10(keeper.OracleDecPrecision)), int64(keeper.OracleDecPrecision))
		exchangeRateStr, err := types.ExchangeRateTuples{
			{ExchangeRate: decExchangeRate, Pair: common.PairBTCStable.String()}}.ToString()
		require.NoError(t, err)

		salt := fmt.Sprintf("%d", i)
		hash := types.GetAggregateVoteHash(salt, exchangeRateStr, valAddrs[i])
		prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, sdk.AccAddress(valAddrs[i]), valAddrs[i])
		voteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, sdk.AccAddress(valAddrs[i]), valAddrs[i])

		_, err1 := h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(0)), prevoteMsg)
		_, err2 := h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(1)), voteMsg)
		require.NoError(t, err1)
		require.NoError(t, err2)

		power := stakingAmt.QuoRaw(int64(6)).Int64()
		if decExchangeRate.IsZero() {
			power = int64(0)
		}

		vote := types.NewBallotVoteForTally(
			decExchangeRate, common.PairBTCStable.String(), valAddrs[i], power)
		ballot = append(ballot, vote)

		// change power of every three validator
		if i%3 == 0 {
			stakingKeeper.Validators()[i].SetConsensusPower(int64(i + 1))
		}
	}

	validatorClaimMap := make(map[string]types.ValidatorPerformance)
	for _, valAddr := range valAddrs {
		validatorClaimMap[valAddr.String()] = types.ValidatorPerformance{
			Power:      stakingKeeper.Validator(input.Ctx, valAddr).GetConsensusPower(sdk.DefaultPowerReduction),
			Weight:     int64(0),
			WinCount:   int64(0),
			ValAddress: valAddr,
		}
	}
	sort.Sort(ballot)
	weightedMedian := ballot.WeightedMedianWithAssertion()
	standardDeviation := ballot.StandardDeviation(weightedMedian)
	maxSpread := weightedMedian.Mul(input.OracleKeeper.RewardBand(input.Ctx).QuoInt64(2))

	if standardDeviation.GT(maxSpread) {
		maxSpread = standardDeviation
	}

	expectedValidatorClaimMap := make(map[string]types.ValidatorPerformance)
	for _, valAddr := range valAddrs {
		expectedValidatorClaimMap[valAddr.String()] = types.ValidatorPerformance{
			Power:      stakingKeeper.Validator(input.Ctx, valAddr).GetConsensusPower(sdk.DefaultPowerReduction),
			Weight:     int64(0),
			WinCount:   int64(0),
			ValAddress: valAddr,
		}
	}

	for _, vote := range ballot {
		if (vote.ExchangeRate.GTE(weightedMedian.Sub(maxSpread)) &&
			vote.ExchangeRate.LTE(weightedMedian.Add(maxSpread))) ||
			!vote.ExchangeRate.IsPositive() {
			key := vote.Voter.String()
			claim := expectedValidatorClaimMap[key]
			claim.Weight += vote.Power
			claim.WinCount++
			expectedValidatorClaimMap[key] = claim
		}
	}

	tallyMedian := keeper.Tally(input.Ctx, ballot, input.OracleKeeper.RewardBand(input.Ctx), validatorClaimMap)

	require.Equal(t, validatorClaimMap, expectedValidatorClaimMap)
	require.Equal(t, tallyMedian.MulInt64(100).TruncateInt(), weightedMedian.MulInt64(100).TruncateInt())
}

func TestOracleTallyTiming(t *testing.T) {
	input, h := setup(t)

	// all the keeper.Addrs vote for the block ... not last period block yet, so tally fails
	for i := range keeper.Addrs[:2] {
		makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairBTCStable.String(), ExchangeRate: randomExchangeRate}}, i)
	}

	params := input.OracleKeeper.GetParams(input.Ctx)
	params.VotePeriod = 10 // set vote period to 10 for now, for convenience
	input.OracleKeeper.SetParams(input.Ctx, params)
	require.Equal(t, 0, int(input.Ctx.BlockHeight()))

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)
	_, err := input.OracleKeeper.GetExchangeRate(input.Ctx, common.PairBTCStable.String())
	require.Error(t, err)

	input.Ctx = input.Ctx.WithBlockHeight(int64(params.VotePeriod - 1))

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)
	_, err = input.OracleKeeper.GetExchangeRate(input.Ctx, common.PairBTCStable.String())
	require.NoError(t, err)
}

func TestOracleRewardDistribution(t *testing.T) {
	// the following test scenario simulates that two validators, out of three, are voting for one common pair.
	// they have the same voting power, and the reward allocation lasts for 1 voting period.
	input, h := setup(t)

	// Account 1, btcstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairBTCStable.String(), ExchangeRate: randomExchangeRate}}, 0)

	// Account 2, btcstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairBTCStable.String(), ExchangeRate: randomExchangeRate}}, 1)

	rewardAllocation := sdk.NewCoins(sdk.NewCoin("reward", sdk.NewInt(1_000_000)))
	keeper.AllocateRewards(t, input, common.PairBTCStable.String(), rewardAllocation, 1)

	oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)

	expectedRewardOneVal := sdk.NewDecCoinsFromCoins(rewardAllocation...).QuoDec(sdk.NewDec(2))
	distributionRewards := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[0])
	require.Equalf(t, expectedRewardOneVal, distributionRewards.Rewards, "%s<=>%s", expectedRewardOneVal.String(), distributionRewards.String())
	distributionRewards = input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[1])
	require.Equal(t, expectedRewardOneVal, distributionRewards.Rewards, "%s %s", expectedRewardOneVal.String(), distributionRewards.Rewards.AmountOf(common.DenomGov).TruncateInt().String())
}

func TestOracleRewardBand(t *testing.T) {
	input, h := setup(t)
	params := input.OracleKeeper.GetParams(input.Ctx)
	params.Whitelist = types.PairList{{Name: common.PairGovStable.String()}}
	input.OracleKeeper.SetParams(input.Ctx, params)

	// clear pairs to reset vote targets
	input.OracleKeeper.ClearPairs(input.Ctx)
	input.OracleKeeper.SetPair(input.Ctx, common.PairGovStable.String())

	rewardSpread := randomExchangeRate.Mul(input.OracleKeeper.RewardBand(input.Ctx).QuoInt64(2))

	// no one will miss the vote
	// Account 1, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate.Sub(rewardSpread)}}, 0)

	// Account 2, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 1)

	// Account 3, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate.Add(rewardSpread)}}, 2)

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)

	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[0]))
	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[1]))
	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[2]))

	// Account 1 will miss the vote due to raward band condition
	// Account 1, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate.Sub(rewardSpread.Add(sdk.OneDec()))}}, 0)

	// Account 2, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 1)

	// Account 3, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate.Add(rewardSpread)}}, 2)

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)

	require.Equal(t, uint64(1), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[0]))
	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[1]))
	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[2]))
}

/* TODO(Mercilex): not appliable right now: https://github.com/NibiruChain/nibiru/issues/805
func TestOracleMultiRewardDistribution(t *testing.T) {
	input, h := setup(t)

	// SDR and KRW have the same voting power, but KRW has been chosen as referencepair by alphabetical order.
	// Account 1, SDR, KRW
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairBTCStable.String(), ExchangeRate: randomExchangeRate}, {Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 0)

	// Account 2, SDR
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairBTCStable.String(), ExchangeRate: randomExchangeRate}}, 1)

	// Account 3, KRW
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairBTCStable.String(), ExchangeRate: randomExchangeRate}}, 2)

	rewardAmt := sdk.NewInt(100000000)
	err := input.BankKeeper.MintCoins(input.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(common.DenomGov, rewardAmt)))
	require.NoError(t, err)

	oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)

	rewardDistributedWindow := input.OracleKeeper.RewardDistributionWindow(input.Ctx)

	expectedRewardAmt := sdk.NewDecFromInt(rewardAmt.QuoRaw(3).MulRaw(2)).QuoInt64(int64(rewardDistributedWindow)).TruncateInt()
	expectedRewardAmt2 := sdk.ZeroInt() // even vote power is same KRW with SDR, KRW chosen referenceTerra because alphabetical order
	expectedRewardAmt3 := sdk.NewDecFromInt(rewardAmt.QuoRaw(3)).QuoInt64(int64(rewardDistributedWindow)).TruncateInt()

	rewards := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[0])
	require.Equal(t, expectedRewardAmt, rewards.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards = input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[1])
	require.Equal(t, expectedRewardAmt2, rewards.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards = input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[2])
	require.Equal(t, expectedRewardAmt3, rewards.Rewards.AmountOf(common.DenomGov).TruncateInt())
}
*/

func TestOracleExchangeRate(t *testing.T) {
	input, h := setup(t)

	govStableExchangeRate := sdk.NewDec(1000000000)
	ethStableExchangeRate := sdk.NewDec(1000000)

	// govstable has been chosen as referenceExchangeRate by highest voting power
	// Account 1, ethstable, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairETHStable.String(), ExchangeRate: ethStableExchangeRate}, {Pair: common.PairGovStable.String(), ExchangeRate: govStableExchangeRate}}, 0)

	// Account 2, ethstable, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairETHStable.String(), ExchangeRate: ethStableExchangeRate}, {Pair: common.PairGovStable.String(), ExchangeRate: govStableExchangeRate}}, 1)

	// Account 3, govstable, btcstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: govStableExchangeRate}, {Pair: common.PairBTCStable.String(), ExchangeRate: randomExchangeRate}}, 2)

	rewardAmt := sdk.NewInt(100000000)
	err := input.BankKeeper.MintCoins(input.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(common.DenomGov, rewardAmt)))
	require.NoError(t, err)

	oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)

	expectedRewardAmt := sdk.NewDecFromInt(rewardAmt.QuoRaw(5).MulRaw(2)).TruncateInt()
	expectedRewardAmt2 := sdk.NewDecFromInt(rewardAmt.QuoRaw(5).MulRaw(1)).TruncateInt()
	rewards := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[0])
	require.Equal(t, expectedRewardAmt, rewards.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards = input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[1])
	require.Equal(t, expectedRewardAmt, rewards.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards = input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[2])
	require.Equal(t, expectedRewardAmt2, rewards.Rewards.AmountOf(common.DenomGov).TruncateInt())
}

func TestOracleEnsureSorted(t *testing.T) {
	input, h := setup(t)

	for i := 0; i < 100; i++ {
		govStableRate1 := sdk.NewDec(int64(rand.Uint64() % 100000000))
		ethStableRate1 := sdk.NewDec(int64(rand.Uint64() % 100000000))

		govStableRate2 := sdk.NewDec(int64(rand.Uint64() % 100000000))
		ethStableRate2 := sdk.NewDec(int64(rand.Uint64() % 100000000))

		govStableRate3 := sdk.NewDec(int64(rand.Uint64() % 100000000))
		ethStableRate3 := sdk.NewDec(int64(rand.Uint64() % 100000000))

		// Account 1, ethstable, govstable
		makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairETHStable.String(), ExchangeRate: ethStableRate1}, {Pair: common.PairGovStable.String(), ExchangeRate: govStableRate1}}, 0)

		// Account 2, ethstable, govstable
		makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairETHStable.String(), ExchangeRate: ethStableRate2}, {Pair: common.PairGovStable.String(), ExchangeRate: govStableRate2}}, 1)

		// Account 3, ethstable, govstable
		makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairETHStable.String(), ExchangeRate: govStableRate3}, {Pair: common.PairGovStable.String(), ExchangeRate: ethStableRate3}}, 2)

		require.NotPanics(t, func() {
			oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)
		})
	}
}

func TestOracleExchangeRateVal5(t *testing.T) {
	input, h := setupVal5(t)

	govStableRate := sdk.NewDec(505000)
	govStableRateErr := sdk.NewDec(500000)
	ethStableRate := sdk.NewDec(505)
	ethStableRateErr := sdk.NewDec(500)

	// govstable has been chosen as reference pair by highest voting power
	// Account 1, govstable, ethstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: govStableRate}, {Pair: common.PairETHStable.String(), ExchangeRate: ethStableRate}}, 0)

	// Account 2, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: govStableRate}}, 1)

	// Account 3, govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: govStableRate}}, 2)

	// Account 4, govstable, ethstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: govStableRateErr}, {Pair: common.PairETHStable.String(), ExchangeRate: ethStableRateErr}}, 3)

	// Account 5, govstable, ethstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: govStableRateErr}, {Pair: common.PairETHStable.String(), ExchangeRate: ethStableRateErr}}, 4)

	rewardAmt := sdk.NewInt(100000000)
	err := input.BankKeeper.MintCoins(input.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(common.DenomGov, rewardAmt)))
	require.NoError(t, err)

	oracle.EndBlocker(input.Ctx.WithBlockHeight(1), input.OracleKeeper)

	gotGovStableRate, err := input.OracleKeeper.GetExchangeRate(input.Ctx, common.PairGovStable.String())
	require.NoError(t, err)
	gotEthStableRate, err := input.OracleKeeper.GetExchangeRate(input.Ctx, common.PairETHStable.String())
	require.NoError(t, err)

	// legacy version case
	require.NotEqual(t, ethStableRateErr, gotEthStableRate)

	// new version case
	require.Equal(t, govStableRate, gotGovStableRate)
	require.Equal(t, ethStableRate, gotEthStableRate)

	expectedRewardAmt := sdk.NewDecFromInt(rewardAmt.QuoRaw(8).MulRaw(2)).TruncateInt()
	expectedRewardAmt2 := sdk.NewDecFromInt(rewardAmt.QuoRaw(8).MulRaw(1)).TruncateInt()
	rewards := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[0])
	require.Equal(t, expectedRewardAmt, rewards.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards1 := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[1])
	require.Equal(t, expectedRewardAmt2, rewards1.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards2 := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[2])
	require.Equal(t, expectedRewardAmt2, rewards2.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards3 := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[3])
	require.Equal(t, expectedRewardAmt, rewards3.Rewards.AmountOf(common.DenomGov).TruncateInt())
	rewards4 := input.DistrKeeper.GetValidatorOutstandingRewards(input.Ctx.WithBlockHeight(2), keeper.ValAddrs[4])
	require.Equal(t, expectedRewardAmt, rewards4.Rewards.AmountOf(common.DenomGov).TruncateInt())
}

func TestVoteTargets(t *testing.T) {
	input, h := setup(t)
	params := input.OracleKeeper.GetParams(input.Ctx)
	params.Whitelist = types.PairList{{Name: common.PairGovStable.String()}, {Name: common.PairBTCStable.String()}}
	input.OracleKeeper.SetParams(input.Ctx, params)

	// clear tobin tax to reset vote targets
	input.OracleKeeper.ClearPairs(input.Ctx)
	input.OracleKeeper.SetPair(input.Ctx, common.PairGovStable.String())

	// govstable
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 0)
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 1)
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 2)

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)

	// no missing current
	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[0]))
	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[1]))
	require.Equal(t, uint64(0), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[2]))

	// vote targets are {govstable, btcstable}
	require.Equal(t, []string{common.PairBTCStable.String(), common.PairGovStable.String()}, input.OracleKeeper.GetVoteTargets(input.Ctx))

	// delete btcstable
	params.Whitelist = types.PairList{{Name: common.PairGovStable.String()}}
	input.OracleKeeper.SetParams(input.Ctx, params)

	// govstable, missing
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 0)
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 1)
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 2)

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)

	require.Equal(t, uint64(1), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[0]))
	require.Equal(t, uint64(1), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[1]))
	require.Equal(t, uint64(1), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[2]))

	// btcstable must be deleted
	require.Equal(t, []string{common.PairGovStable.String()}, input.OracleKeeper.GetVoteTargets(input.Ctx))

	exists := input.OracleKeeper.PairExists(input.Ctx, common.PairBTCStable.String())
	require.False(t, exists)

	// change govstable tobin tax
	params.Whitelist = types.PairList{{Name: common.PairGovStable.String()}}
	input.OracleKeeper.SetParams(input.Ctx, params)

	// govstable, no missing
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 0)
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 1)
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: randomExchangeRate}}, 2)

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)

	require.Equal(t, uint64(1), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[0]))
	require.Equal(t, uint64(1), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[1]))
	require.Equal(t, uint64(1), input.OracleKeeper.GetMissCounter(input.Ctx, keeper.ValAddrs[2]))
}

func TestAbstainWithSmallStakingPower(t *testing.T) {
	input, h := setupWithSmallVotingPower(t)

	// clear tobin tax to reset vote targets
	input.OracleKeeper.ClearPairs(input.Ctx)
	input.OracleKeeper.SetPair(input.Ctx, common.PairGovStable.String())
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.PairGovStable.String(), ExchangeRate: sdk.ZeroDec()}}, 0)

	oracle.EndBlocker(input.Ctx, input.OracleKeeper)
	_, err := input.OracleKeeper.GetExchangeRate(input.Ctx, common.PairGovStable.String())
	require.Error(t, err)
}

func makeAggregatePrevoteAndVote(t *testing.T, input keeper.TestInput, h types.MsgServer, height int64, rates types.ExchangeRateTuples, idx int) {
	salt := "1"
	ratesStr, err := rates.ToString()
	require.NoError(t, err)
	hash := types.GetAggregateVoteHash(salt, ratesStr, keeper.ValAddrs[idx])

	prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, keeper.Addrs[idx], keeper.ValAddrs[idx])
	_, err = h.AggregateExchangeRatePrevote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(height)), prevoteMsg)
	require.NoError(t, err)

	voteMsg := types.NewMsgAggregateExchangeRateVote(salt, ratesStr, keeper.Addrs[idx], keeper.ValAddrs[idx])
	_, err = h.AggregateExchangeRateVote(sdk.WrapSDKContext(input.Ctx.WithBlockHeight(height+1)), voteMsg)
	require.NoError(t, err)
}

func setupWithSmallVotingPower(t *testing.T) (keeper.TestInput, types.MsgServer) {
	input := keeper.CreateTestInput(t)
	params := input.OracleKeeper.GetParams(input.Ctx)
	params.VotePeriod = 1
	params.SlashWindow = 100
	input.OracleKeeper.SetParams(input.Ctx, params)
	h := keeper.NewMsgServerImpl(input.OracleKeeper)

	sh := staking.NewHandler(input.StakingKeeper)
	_, err := sh(input.Ctx, keeper.NewTestMsgCreateValidator(keeper.ValAddrs[0], keeper.ValPubKeys[0], sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)))
	require.NoError(t, err)

	staking.EndBlocker(input.Ctx, input.StakingKeeper)

	return input, h
}

func setupVal5(t *testing.T) (keeper.TestInput, types.MsgServer) {
	input := keeper.CreateTestInput(t)
	params := input.OracleKeeper.GetParams(input.Ctx)
	params.VotePeriod = 1
	params.SlashWindow = 100
	input.OracleKeeper.SetParams(input.Ctx, params)
	h := keeper.NewMsgServerImpl(input.OracleKeeper)

	sh := staking.NewHandler(input.StakingKeeper)

	// Validator created
	_, err := sh(input.Ctx, keeper.NewTestMsgCreateValidator(keeper.ValAddrs[0], keeper.ValPubKeys[0], stakingAmt))
	require.NoError(t, err)
	_, err = sh(input.Ctx, keeper.NewTestMsgCreateValidator(keeper.ValAddrs[1], keeper.ValPubKeys[1], stakingAmt))
	require.NoError(t, err)
	_, err = sh(input.Ctx, keeper.NewTestMsgCreateValidator(keeper.ValAddrs[2], keeper.ValPubKeys[2], stakingAmt))
	require.NoError(t, err)
	_, err = sh(input.Ctx, keeper.NewTestMsgCreateValidator(keeper.ValAddrs[3], keeper.ValPubKeys[3], stakingAmt))
	require.NoError(t, err)
	_, err = sh(input.Ctx, keeper.NewTestMsgCreateValidator(keeper.ValAddrs[4], keeper.ValPubKeys[4], stakingAmt))
	require.NoError(t, err)
	staking.EndBlocker(input.Ctx, input.StakingKeeper)

	return input, h
}
