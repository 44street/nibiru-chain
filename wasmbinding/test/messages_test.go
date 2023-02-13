package wasmbinding

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/nibiru/wasmbinding"
	"github.com/NibiruChain/nibiru/wasmbinding/bindings"
	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/nibiru/x/common/asset"
	perptypes "github.com/NibiruChain/nibiru/x/perp/types"
	vpooltypes "github.com/NibiruChain/nibiru/x/vpool/types"

	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/nibiru/app"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/assert"
)

func fundAccount(t *testing.T, ctx sdk.Context, app *app.NibiruApp, addr sdk.AccAddress, coins sdk.Coins) {
	err := simapp.FundAccount(
		app.BankKeeper,
		ctx,
		addr,
		coins,
	)
	require.NoError(t, err)
}

func TestOpenClosePosition(t *testing.T) {
	actor := RandomAccountAddress()
	app, ctx := SetupCustomApp(t, actor)
	ctx = ctx.WithBlockTime(time.Now())
	tokenPair := asset.MustNewPair("BTC:NUSD")

	specs := map[string]struct {
		openPosition *bindings.OpenPosition
		expErr       bool
	}{
		"valid open-position": {
			openPosition: &bindings.OpenPosition{
				Pair:                 "BTC:NUSD",
				Side:                 int(perptypes.Side_BUY),
				QuoteAssetAmount:     sdk.NewInt(10),
				Leverage:             sdk.OneDec(),
				BaseAssetAmountLimit: sdk.ZeroInt(),
			},
		},
		"invalid open-position": {
			openPosition: &bindings.OpenPosition{
				Pair: "",
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			t.Log("Create vpool")
			vpoolKeeper := &app.VpoolKeeper
			perpKeeper := &app.PerpKeeper
			assert.NoError(t, vpoolKeeper.CreatePool(
				ctx,
				tokenPair,
				sdk.NewDec(10*common.Precision),
				sdk.NewDec(5*common.Precision),
				vpooltypes.VpoolConfig{
					TradeLimitRatio:        sdk.MustNewDecFromStr("0.9"),
					FluctuationLimitRatio:  sdk.OneDec(),
					MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.1"),
					MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
					MaxLeverage:            sdk.MustNewDecFromStr("15"),
				},
			))
			require.True(t, vpoolKeeper.ExistsPool(ctx, tokenPair))
			app.OracleKeeper.SetPrice(ctx, tokenPair, sdk.NewDec(2))

			pairMetadata := perptypes.PairMetadata{
				Pair:                            tokenPair,
				LatestCumulativePremiumFraction: sdk.ZeroDec(),
			}
			perpKeeper.PairsMetadata.Insert(ctx, pairMetadata.Pair, pairMetadata)

			t.Log("Fund trader account with sufficient quote")
			fundAccount(t, ctx, app, actor, sdk.NewCoins(sdk.NewInt64Coin("NUSD", 50_100)))

			t.Log("Increment block height and time for TWAP calculation")
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).
				WithBlockTime(time.Now().Add(time.Minute))

			t.Log("Open position")
			gotErr := wasmbinding.PerformOpenPosition(perpKeeper, ctx, actor, spec.openPosition)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestClosePosition(t *testing.T) {
	actor := RandomAccountAddress()
	app, ctx := SetupCustomApp(t, actor)
	ctx = ctx.WithBlockTime(time.Now())
	tokenPair := asset.MustNewPair("BTC:NUSD")

	specs := map[string]struct {
		closePosition *bindings.ClosePosition
		expErr        bool
	}{
		"valid close-position": {
			closePosition: &bindings.ClosePosition{
				Pair: "BTC:NUSD",
			},
		},
		"invalid close-position": {
			closePosition: &bindings.ClosePosition{
				Pair: "",
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			t.Log("Create vpool")
			vpoolKeeper := &app.VpoolKeeper
			perpKeeper := &app.PerpKeeper
			assert.NoError(t, vpoolKeeper.CreatePool(
				ctx,
				tokenPair,
				sdk.NewDec(10*common.Precision),
				sdk.NewDec(5*common.Precision),
				vpooltypes.VpoolConfig{
					TradeLimitRatio:        sdk.MustNewDecFromStr("0.9"),
					FluctuationLimitRatio:  sdk.OneDec(),
					MaxOracleSpreadRatio:   sdk.MustNewDecFromStr("0.1"),
					MaintenanceMarginRatio: sdk.MustNewDecFromStr("0.0625"),
					MaxLeverage:            sdk.MustNewDecFromStr("15"),
				},
			))
			require.True(t, vpoolKeeper.ExistsPool(ctx, tokenPair))
			app.OracleKeeper.SetPrice(ctx, tokenPair, sdk.NewDec(2))

			pairMetadata := perptypes.PairMetadata{
				Pair:                            tokenPair,
				LatestCumulativePremiumFraction: sdk.ZeroDec(),
			}
			perpKeeper.PairsMetadata.Insert(ctx, pairMetadata.Pair, pairMetadata)

			t.Log("Fund trader account with sufficient quote")
			fundAccount(t, ctx, app, actor, sdk.NewCoins(sdk.NewInt64Coin("NUSD", 50_100)))

			t.Log("Increment block height and time for TWAP calculation")
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).
				WithBlockTime(time.Now().Add(time.Minute))

			t.Log("Open position")
			assert.NoError(t, wasmbinding.PerformOpenPosition(perpKeeper, ctx, actor, &bindings.OpenPosition{
				Pair:                 "BTC:NUSD",
				Side:                 int(perptypes.Side_BUY),
				QuoteAssetAmount:     sdk.NewInt(10),
				Leverage:             sdk.OneDec(),
				BaseAssetAmountLimit: sdk.ZeroInt(),
			}))

			t.Log("Close position")
			gotErr := wasmbinding.PerformClosePosition(perpKeeper, ctx, actor, spec.closePosition)

			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
