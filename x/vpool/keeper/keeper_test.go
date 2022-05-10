package keeper

import (
	"fmt"
	"testing"

	"github.com/NibiruChain/nibiru/x/common"
	pftypes "github.com/NibiruChain/nibiru/x/pricefeed/types"
	"github.com/NibiruChain/nibiru/x/testutil/mock"
	"github.com/NibiruChain/nibiru/x/vpool/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestSwapInput_Errors(t *testing.T) {
	tests := []struct {
		name        string
		pair        common.TokenPair
		direction   types.Direction
		quoteAmount sdk.Int
		baseLimit   sdk.Int
		error       error
	}{
		{
			"pair not supported",
			"BTC:UST",
			types.Direction_ADD_TO_POOL,
			sdk.NewInt(10),
			sdk.NewInt(10),
			types.ErrPairNotSupported,
		},
		{
			"base amount less than base limit in Long",
			NUSDPair,
			types.Direction_ADD_TO_POOL,
			sdk.NewInt(500_000),
			sdk.NewInt(454_500),
			fmt.Errorf("base amount (238095) is less than selected limit (454500)"),
		},
		{
			"base amount more than base limit in Short",
			NUSDPair,
			types.Direction_REMOVE_FROM_POOL,
			sdk.NewInt(1_000_000),
			sdk.NewInt(454_500),
			fmt.Errorf("base amount (555556) is greater than selected limit (454500)"),
		},
		{
			"quote input bigger than reserve ratio",
			NUSDPair,
			types.Direction_REMOVE_FROM_POOL,
			sdk.NewInt(10_000_000),
			sdk.NewInt(10),
			types.ErrOvertradingLimit,
		},
		{
			"over fluctuation limit fails",
			NUSDPair,
			types.Direction_ADD_TO_POOL,
			sdk.NewInt(1_000_000),
			sdk.NewInt(454544),
			fmt.Errorf("error updating reserve: %w", types.ErrOverFluctuationLimit),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vpoolKeeper, ctx := VpoolKeeper(t,
				mock.NewMockPriceKeeper(gomock.NewController(t)),
			)

			vpoolKeeper.CreatePool(
				ctx,
				NUSDPair,
				sdk.MustNewDecFromStr("0.9"), // 0.9 ratio
				sdk.NewInt(10_000_000),       // 10
				sdk.NewInt(5_000_000),        // 5
				sdk.MustNewDecFromStr("0.1"), // 0.1 fluctuation limit ratio
			)

			_, err := vpoolKeeper.SwapInput(
				ctx,
				tc.pair,
				tc.direction,
				tc.quoteAmount,
				tc.baseLimit,
			)
			require.EqualError(t, err, tc.error.Error())
		})
	}
}

func TestSwapInput_HappyPath(t *testing.T) {
	tests := []struct {
		name                 string
		direction            types.Direction
		quoteAmount          sdk.Int
		baseLimit            sdk.Int
		expectedQuoteReserve sdk.Int
		expectedBaseReserve  sdk.Int
		resp                 sdk.Int
	}{
		{
			"quote amount == 0",
			types.Direction_ADD_TO_POOL,
			sdk.NewInt(0),
			sdk.NewInt(10),
			sdk.NewInt(10_000_000),
			sdk.NewInt(5_000_000),
			sdk.ZeroInt(),
		},
		{
			"normal swap add",
			types.Direction_ADD_TO_POOL,
			sdk.NewInt(1_000_000),
			sdk.NewInt(454_500),
			sdk.NewInt(11_000_000),
			sdk.NewInt(4_545_455),
			sdk.NewInt(454_545),
		},
		{
			"normal swap remove",
			types.Direction_REMOVE_FROM_POOL,
			sdk.NewInt(1_000_000),
			sdk.NewInt(555_560),
			sdk.NewInt(9_000_000),
			sdk.NewInt(5_555_556),
			sdk.NewInt(555_556),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vpoolKeeper, ctx := VpoolKeeper(t,
				mock.NewMockPriceKeeper(gomock.NewController(t)),
			)

			vpoolKeeper.CreatePool(
				ctx,
				NUSDPair,
				sdk.MustNewDecFromStr("0.9"),  // 0.9 ratio
				sdk.NewInt(10_000_000),        // 10 tokens
				sdk.NewInt(5_000_000),         // 5 tokens
				sdk.MustNewDecFromStr("0.25"), // 0.25 ratio
			)

			res, err := vpoolKeeper.SwapInput(
				ctx,
				NUSDPair,
				tc.direction,
				tc.quoteAmount,
				tc.baseLimit,
			)
			require.NoError(t, err)
			require.Equal(t, res, tc.resp)

			pool, err := vpoolKeeper.getPool(ctx, NUSDPair)
			require.NoError(t, err)
			require.Equal(t, tc.expectedQuoteReserve, pool.QuoteAssetReserve)
			require.Equal(t, tc.expectedBaseReserve, pool.BaseAssetReserve)
		})
	}
}

func TestGetUnderlyingPrice(t *testing.T) {
	tests := []struct {
		name           string
		pair           common.TokenPair
		pricefeedPrice sdk.Dec
	}{
		{
			name:           "correctly fetch underlying price",
			pair:           common.TokenPair("btc:nusd"),
			pricefeedPrice: sdk.MustNewDecFromStr("40000"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockPricefeedKeeper := mock.NewMockPriceKeeper(gomock.NewController(t))
			vpoolKeeper, ctx := VpoolKeeper(t, mockPricefeedKeeper)

			mockPricefeedKeeper.
				EXPECT().
				GetCurrentPrice(
					gomock.Eq(ctx),
					gomock.Eq(tc.pair.GetBaseTokenDenom()),
					gomock.Eq(tc.pair.GetQuoteTokenDenom()),
				).
				Return(
					pftypes.CurrentPrice{
						PairID: tc.pair.String(),
						Price:  tc.pricefeedPrice,
					}, nil,
				)

			price, err := vpoolKeeper.GetUnderlyingPrice(ctx, tc.pair)
			require.NoError(t, err)
			require.EqualValues(t, tc.pricefeedPrice, price)
		})
	}
}
