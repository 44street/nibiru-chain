package keeper

import (
	"errors"
	"fmt"
	"time"

	"github.com/NibiruChain/nibiru/x/common"
	pooltypes "github.com/NibiruChain/nibiru/x/vpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/nibiru/x/perp/types"
)

/* TODO tests | These _ vars are here to pass the golangci-lint for unused methods.
They also serve as a reminder of which functions still need MVP unit or
integration tests */
var (
	_ = Keeper.swapQuoteForBase
	_ = Keeper.closePosition
	_ = Keeper.increasePosition
	_ = Keeper.reducePosition
	_ = Keeper.closeAndOpenReversePosition
	_ = Keeper.openReversePosition
	_ = Keeper.transferFee
)

// TODO test: OpenPosition | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) OpenPosition(
	ctx sdk.Context,
	pair common.TokenPair,
	side types.Side,
	traderAddr sdk.AccAddress,
	quoteAssetAmount sdk.Int,
	leverage sdk.Dec,
	baseAssetAmountLimit sdk.Int,
) (err error) {
	// require vpool
	err = k.requireVpool(ctx, pair)
	if err != nil {
		return err
	}
	// require params
	params := k.GetParams(ctx)
	// TODO: missing checks

	position, err := k.GetPosition(ctx, pair, traderAddr.String())
	var isNewPosition bool = errors.Is(err, types.ErrPositionNotFound)
	if isNewPosition {
		position = types.ZeroPosition(ctx, pair, traderAddr.String())
		k.SetPosition(ctx, pair, traderAddr.String(), position)
	} else if err != nil && !isNewPosition {
		return err
	}

	var positionResp *types.PositionResp
	sameSideLong := position.Size_.IsPositive() && side == types.Side_BUY
	sameSideShort := position.Size_.IsNegative() && side == types.Side_SELL
	var openSideMatchesPosition bool = (sameSideLong || sameSideShort)
	switch {
	case isNewPosition || openSideMatchesPosition:
		// increase position case

		positionResp, err = k.increasePosition(
			ctx,
			position,
			side,
			/* openNotional */ leverage.MulInt(quoteAssetAmount),
			/* minPositionSize */ baseAssetAmountLimit.ToDec(),
			/* leverage */ leverage)
		if err != nil {
			return err
		}

	// everything else decreases the position
	default:
		positionResp, err = k.openReversePosition(
			ctx,
			position,
			side,
			quoteAssetAmount,
			leverage,
			baseAssetAmountLimit,
			false,
		)
		if err != nil {
			return err
		}
	}

	// update position in state
	k.SetPosition(ctx, pair, traderAddr.String(), positionResp.Position)

	if !isNewPosition && !positionResp.Position.Size_.IsZero() {
		marginRatio, err := k.GetMarginRatio(ctx, *positionResp.Position)
		if err != nil {
			return err
		}
		if err = requireMoreMarginRatio(
			marginRatio, params.MaintenanceMarginRatio, true); err != nil {
			// TODO(mercilex): should panic? it's a require
			return err
		}
	}

	if !positionResp.BadDebt.IsZero() {
		return fmt.Errorf(
			"bad debt must be zero to prevent attacker from leveraging it")
	}

	// transfer trader <=> vault
	switch {
	case positionResp.MarginToVault.IsPositive():
		err = k.BankKeeper.SendCoinsFromAccountToModule(
			ctx, traderAddr, types.VaultModuleAccount,
			sdk.NewCoins(sdk.NewCoin(pair.GetQuoteTokenDenom(), positionResp.MarginToVault.TruncateInt())))
		if err != nil {
			return err
		}
	case positionResp.MarginToVault.IsNegative():
		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.VaultModuleAccount, traderAddr,
			sdk.NewCoins(sdk.NewCoin(pair.GetQuoteTokenDenom(), positionResp.MarginToVault.Abs().TruncateInt())))
		if err != nil {
			return err
		}
	}

	transferredFee, err := k.transferFee(ctx, pair, traderAddr, positionResp.ExchangedQuoteAssetAmount.TruncateInt())
	if err != nil {
		return err
	}

	spotPrice, err := k.VpoolKeeper.GetSpotPrice(ctx, pair)
	if err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.PositionChangedEvent{
		Trader:                traderAddr.String(),
		Pair:                  pair.String(),
		Margin:                positionResp.Position.Margin,
		PositionNotional:      positionResp.ExchangedPositionSize,
		ExchangedPositionSize: positionResp.ExchangedPositionSize,
		Fee:                   transferredFee.ToDec(), // TODO(mercilex): this feels like should be a coin?
		PositionSizeAfter:     positionResp.Position.Size_,
		RealizedPnl:           positionResp.RealizedPnl,
		UnrealizedPnlAfter:    positionResp.UnrealizedPnlAfter,
		BadDebt:               positionResp.BadDebt,
		LiquidationPenalty:    sdk.ZeroDec(),
		SpotPrice:             spotPrice,
		FundingPayment:        positionResp.FundingPayment,
	})
}

// TODO test: increasePosition | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) increasePosition(
	ctx sdk.Context,
	oldPosition *types.Position,
	side types.Side,
	openNotional sdk.Dec,
	minPositionSize sdk.Dec,
	leverage sdk.Dec,
) (positionResp *types.PositionResp, err error) {
	positionResp = &types.PositionResp{}

	positionResp.ExchangedPositionSize, err = k.swapQuoteForBase(
		ctx,
		common.TokenPair(oldPosition.Pair),
		side,
		openNotional,
		minPositionSize,
		false,
	)
	if err != nil {
		return nil, err
	}

	newSize := oldPosition.Size_.Add(positionResp.ExchangedPositionSize)

	increaseMarginRequirement := openNotional.Quo(leverage)

	remaining, err := k.CalcRemainMarginWithFundingPayment(
		ctx,
		*oldPosition,
		increaseMarginRequirement,
	)
	if err != nil {
		return nil, err
	}

	_, unrealizedPnL, err := k.getPositionNotionalAndUnrealizedPnL(
		ctx,
		*oldPosition,
		types.PnLCalcOption_SPOT_PRICE,
	)
	if err != nil {
		return nil, err
	}

	positionResp.ExchangedQuoteAssetAmount = openNotional
	positionResp.UnrealizedPnlAfter = unrealizedPnL
	positionResp.MarginToVault = increaseMarginRequirement
	positionResp.FundingPayment = remaining.FPayment
	positionResp.BadDebt = remaining.BadDebt
	positionResp.Position = &types.Position{
		Address:                             oldPosition.Address,
		Pair:                                oldPosition.Pair,
		Size_:                               newSize,
		Margin:                              remaining.Margin,
		OpenNotional:                        oldPosition.OpenNotional.Add(positionResp.ExchangedQuoteAssetAmount),
		LastUpdateCumulativePremiumFraction: remaining.LatestCPF,
		LiquidityHistoryIndex:               oldPosition.LiquidityHistoryIndex,
		BlockNumber:                         ctx.BlockHeight(),
	}

	return
}

// getLatestCumulativePremiumFraction returns the last cumulative premium fraction recorded for the
// specific pair.
func (k Keeper) getLatestCumulativePremiumFraction(
	ctx sdk.Context, pair common.TokenPair,
) (sdk.Dec, error) {
	pairMetadata, err := k.PairMetadata().Get(ctx, pair)
	if err != nil {
		return sdk.Dec{}, err
	}
	// this should never fail
	return pairMetadata.CumulativePremiumFractions[len(pairMetadata.CumulativePremiumFractions)-1], nil
}

// TODO test: getPositionNotionalAndUnrealizedPnL | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) getPositionNotionalAndUnrealizedPnL(
	ctx sdk.Context,
	oldPosition types.Position,
	pnlCalcOption types.PnLCalcOption,
) (notional, unrealizedPnL sdk.Dec, err error) {
	positionSizeAbs := oldPosition.Size_.Abs()
	if positionSizeAbs.IsZero() {
		return sdk.ZeroDec(), sdk.ZeroDec(), nil
	}

	isShortPosition := oldPosition.Size_.IsNegative()
	var dir pooltypes.Direction
	switch isShortPosition {
	case true:
		dir = pooltypes.Direction_REMOVE_FROM_POOL
	default:
		dir = pooltypes.Direction_ADD_TO_POOL
	}

	switch pnlCalcOption {
	case types.PnLCalcOption_TWAP:
		notionalDec, err := k.VpoolKeeper.GetBaseAssetTWAP(
			ctx,
			common.TokenPair(oldPosition.Pair),
			dir,
			positionSizeAbs,
			15*time.Minute,
		)
		if err != nil {
			return sdk.ZeroDec(), sdk.ZeroDec(), err
		}
		notional = notionalDec
	case types.PnLCalcOption_SPOT_PRICE:
		notionalDec, err := k.VpoolKeeper.GetBaseAssetPrice(ctx, common.TokenPair(oldPosition.Pair), dir, positionSizeAbs)
		if err != nil {
			return sdk.ZeroDec(), sdk.ZeroDec(), err
		}
		notional = notionalDec
	case types.PnLCalcOption_ORACLE:
		oraclePrice, err := k.VpoolKeeper.GetUnderlyingPrice(ctx, common.TokenPair(oldPosition.Pair))
		if err != nil {
			return sdk.ZeroDec(), sdk.ZeroDec(), err
		}
		notional = oraclePrice.Mul(positionSizeAbs)
	default:
		panic("unrecognized pnl calc option: " + pnlCalcOption.String())
	}

	switch isShortPosition {
	case true:
		unrealizedPnL = oldPosition.OpenNotional.Sub(notional)
	case false:
		unrealizedPnL = notional.Sub(oldPosition.OpenNotional)
	}

	return unrealizedPnL, notional, nil
}

// TODO test: openReversePosition | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) openReversePosition(
	ctx sdk.Context,
	oldPosition *types.Position,
	side types.Side,
	quoteAssetAmount sdk.Int,
	leverage sdk.Dec,
	baseAssetAmountLimit sdk.Int,
	canOverFluctuationLimit bool,
) (positionResp *types.PositionResp, err error) {
	openNotional := leverage.MulInt(quoteAssetAmount)
	oldPositionNotional, unrealizedPnL, err := k.getPositionNotionalAndUnrealizedPnL(
		ctx,
		*oldPosition,
		types.PnLCalcOption_SPOT_PRICE,
	)
	if err != nil {
		return nil, err
	}

	switch oldPositionNotional.GT(openNotional) {
	// position reduction
	case true:
		return k.reducePosition(
			ctx,
			oldPosition,
			side,
			openNotional,
			oldPositionNotional,
			baseAssetAmountLimit.ToDec(),
			unrealizedPnL,
			canOverFluctuationLimit,
		)
	// close and reverse
	default:
		return k.closeAndOpenReversePosition(
			ctx,
			oldPosition,
			side,
			quoteAssetAmount,
			leverage,
			baseAssetAmountLimit,
		)
	}
}

// TODO test: reducePosition | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) reducePosition(
	ctx sdk.Context,
	oldPosition *types.Position,
	side types.Side,
	openNotional,
	oldPositionNotional,
	baseAssetAmountLimit,
	unrealizedPnL sdk.Dec,
	canOverFluctuationLimit bool,
) (positionResp *types.PositionResp, err error) {
	positionResp = new(types.PositionResp)

	positionResp.ExchangedPositionSize, err = k.swapQuoteForBase(
		ctx,
		common.TokenPair(oldPosition.Pair),
		side,
		openNotional,
		baseAssetAmountLimit,
		canOverFluctuationLimit,
	)
	if err != nil {
		return nil, err
	}

	if !oldPosition.Size_.IsZero() {
		var realizedPnL = unrealizedPnL.Mul(positionResp.ExchangedPositionSize.Abs()).Quo(oldPosition.Size_.Abs())
		positionResp.RealizedPnl = realizedPnL
	}
	remaining, err := k.CalcRemainMarginWithFundingPayment(
		ctx,
		*oldPosition,
		positionResp.RealizedPnl,
	)
	positionResp.BadDebt = remaining.BadDebt
	positionResp.FundingPayment = remaining.FPayment
	if err != nil {
		return nil, err
	}

	positionResp.UnrealizedPnlAfter = unrealizedPnL.Sub(positionResp.RealizedPnl)
	positionResp.ExchangedQuoteAssetAmount = openNotional

	var remainOpenNotional sdk.Dec
	switch oldPosition.Size_.IsPositive() {
	case true:
		remainOpenNotional = oldPositionNotional.Sub(positionResp.ExchangedQuoteAssetAmount).Sub(positionResp.UnrealizedPnlAfter)
	case false:
		remainOpenNotional = positionResp.UnrealizedPnlAfter.Add(oldPositionNotional).Sub(positionResp.ExchangedQuoteAssetAmount)
	}

	if remainOpenNotional.LTE(sdk.ZeroDec()) {
		panic("value of open notional <= 0")
	}

	positionResp.Position = &types.Position{
		Address:                             oldPosition.Address,
		Pair:                                oldPosition.Pair,
		Size_:                               oldPosition.Size_.Add(positionResp.ExchangedPositionSize),
		Margin:                              remaining.Margin,
		OpenNotional:                        remainOpenNotional.Abs(),
		LastUpdateCumulativePremiumFraction: remaining.LatestCPF,
		LiquidityHistoryIndex:               oldPosition.LiquidityHistoryIndex,
		BlockNumber:                         ctx.BlockHeight(),
	}
	return positionResp, nil
}

// TODO test: closeAndOpenReversePosition | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) closeAndOpenReversePosition(
	ctx sdk.Context,
	oldPosition *types.Position,
	side types.Side,
	quoteAssetAmount sdk.Int,
	leverage sdk.Dec,
	baseAssetAmountLimit sdk.Int,
) (positionResp *types.PositionResp, err error) {
	positionResp = new(types.PositionResp)

	closePositionResp, err := k.closePosition(ctx, oldPosition, sdk.ZeroInt())
	if err != nil {
		return nil, err
	}

	if closePositionResp.BadDebt.LTE(sdk.ZeroDec()) {
		return nil, fmt.Errorf("underwater position")
	}

	openNotional := leverage.MulInt(quoteAssetAmount).Sub(closePositionResp.ExchangedQuoteAssetAmount)

	switch openNotional.Quo(leverage).IsZero() {
	case true:
		positionResp = closePositionResp
	case false:
		var updatedBaseAssetAmountLimit sdk.Dec
		if baseAssetAmountLimit.ToDec().GT(closePositionResp.ExchangedPositionSize) {
			updatedBaseAssetAmountLimit = baseAssetAmountLimit.ToDec().
				Sub(closePositionResp.ExchangedPositionSize.Abs())
		}

		var increasePositionResp *types.PositionResp
		increasePositionResp, err = k.increasePosition(
			ctx,
			oldPosition,
			side,
			openNotional,
			updatedBaseAssetAmountLimit,
			leverage,
		)
		if err != nil {
			return nil, err
		}
		positionResp = &types.PositionResp{
			Position:                  increasePositionResp.Position,
			ExchangedQuoteAssetAmount: closePositionResp.ExchangedQuoteAssetAmount.Add(increasePositionResp.ExchangedQuoteAssetAmount),
			BadDebt:                   closePositionResp.BadDebt.Add(increasePositionResp.BadDebt),
			ExchangedPositionSize:     closePositionResp.ExchangedPositionSize.Add(increasePositionResp.ExchangedPositionSize),
			FundingPayment:            closePositionResp.FundingPayment.Add(increasePositionResp.FundingPayment),
			RealizedPnl:               closePositionResp.RealizedPnl.Add(increasePositionResp.RealizedPnl),
			MarginToVault:             closePositionResp.MarginToVault.Add(increasePositionResp.MarginToVault),
			UnrealizedPnlAfter:        sdk.ZeroDec(),
		}
	}

	return positionResp, nil
}

// TODO test: closePosition | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) closePosition(
	ctx sdk.Context,
	oldPosition *types.Position,
	quoteAssetAmountLimit sdk.Int,
) (positionResp *types.PositionResp, err error) {
	positionResp = new(types.PositionResp)

	if oldPosition.Size_.IsZero() {
		return nil, fmt.Errorf("zero position size")
	}
	_, unrealizedPnL, err := k.getPositionNotionalAndUnrealizedPnL(
		ctx,
		*oldPosition,
		types.PnLCalcOption_SPOT_PRICE,
	)
	if err != nil {
		return nil, err
	}

	remaining, err := k.CalcRemainMarginWithFundingPayment(
		ctx, *oldPosition, unrealizedPnL)
	if err != nil {
		return nil, err
	}

	positionResp.ExchangedPositionSize = oldPosition.Size_.Neg()
	positionResp.RealizedPnl = unrealizedPnL
	positionResp.BadDebt = remaining.BadDebt
	positionResp.FundingPayment = remaining.FPayment
	positionResp.MarginToVault = remaining.Margin.Neg()

	var vammDir pooltypes.Direction
	switch oldPosition.Size_.GTE(sdk.ZeroDec()) {
	case true:
		vammDir = pooltypes.Direction_ADD_TO_POOL
	case false:
		vammDir = pooltypes.Direction_REMOVE_FROM_POOL
	}
	exchangedQuoteAssetAmount, err :=
		k.VpoolKeeper.SwapBaseForQuote(
			ctx,
			common.TokenPair(oldPosition.Pair),
			vammDir,
			oldPosition.Size_.Abs(),
			quoteAssetAmountLimit.ToDec(),
		)
	if err != nil {
		return nil, err
	}

	positionResp.ExchangedQuoteAssetAmount = exchangedQuoteAssetAmount

	err = k.ClearPosition(ctx, common.TokenPair(oldPosition.Pair), oldPosition.Address)
	if err != nil {
		return nil, err
	}

	return positionResp, nil
}

// TODO test: transferFee | https://github.com/NibiruChain/nibiru/issues/299
func (k Keeper) transferFee(
	ctx sdk.Context, pair common.TokenPair, trader sdk.AccAddress,
	positionNotional sdk.Int,
) (sdk.Int, error) {
	toll, spread, err := k.CalcFee(ctx, positionNotional)
	if err != nil {
		return sdk.Int{}, err
	}

	hasToll := toll.IsPositive()
	hasSpread := spread.IsPositive()

	if !hasToll && hasSpread {
		// TODO(mercilex): what's the meaning of returning sdk.Int if both evaluate to false, should this happen?
		return sdk.Int{}, nil
	}

	if hasSpread {
		err = k.BankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.PerpEFModuleAccount,
			sdk.NewCoins(sdk.NewCoin(pair.GetQuoteTokenDenom(), spread)))
		if err != nil {
			return sdk.Int{}, err
		}
	}
	if hasToll {
		err = k.BankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.FeePoolModuleAccount,
			sdk.NewCoins(sdk.NewCoin(pair.GetQuoteTokenDenom(), toll)))
		if err != nil {
			return sdk.Int{}, err
		}
	}

	return toll.Add(spread), nil
}

// TODO test: getPreferencePositionNotionalAndUnrealizedPnL
/* getPreferencePositionNotionalAndUnrealizedPnL

Returns:
  pnl: unrealized profits and losses (PnL)
  notional: positional notional.
*/
func (k Keeper) getPreferencePositionNotionalAndUnrealizedPnL(
	ctx sdk.Context,
	oldPosition types.Position,
	pnLPreferenceOption types.PnLPreferenceOption,
) (pnl sdk.Dec, notional sdk.Dec, er error) {
	// TODO(mercilex): maybe inefficient get position notional and unrealized pnl
	spotPositionNotional, spotPricePnl, err := k.getPositionNotionalAndUnrealizedPnL(
		ctx,
		oldPosition,
		types.PnLCalcOption_SPOT_PRICE,
	)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}

	twapPositionNotional, twapPricePnL, err := k.getPositionNotionalAndUnrealizedPnL(
		ctx,
		oldPosition,
		types.PnLCalcOption_TWAP,
	)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}

	// todo(mercilex): logic can be simplified here but keeping it for now as perp reference
	switch pnLPreferenceOption {
	// if MAX PNL
	case types.PnLPreferenceOption_MAX:
		// spotPNL > twapPnL
		switch spotPricePnl.GT(twapPricePnL) {
		// true: spotPNL > twapPNL -> return spot pnl, spot position notional
		case true:
			return spotPricePnl, spotPositionNotional, nil
		// false: spotPNL <= twapPNL -> return twapPNL twapPositionNotional
		default:
			return twapPricePnL, twapPositionNotional, nil
		}
	// if min PNL
	case types.PnLPreferenceOption_MIN:
		switch spotPricePnl.GT(twapPricePnL) {
		// true: spotPNL > twapPNL -> return twapPNL, twapPositionNotional
		case true:
			return twapPricePnL, twapPositionNotional, nil
		// false: spotPNL <= twapPNL -> return spotPNL, spotPositionNotional
		default:
			return spotPricePnl, spotPositionNotional, nil
		}
	default:
		panic("invalid pnl preference option " + pnLPreferenceOption.String())
	}
}

// TODO: Check Can Over Fluctuation Limit
func (k Keeper) swapQuoteForBase(
	ctx sdk.Context,
	pair common.TokenPair,
	side types.Side,
	quoteAmount sdk.Dec,
	baseLimit sdk.Dec,
	canOverFluctuationLimit bool,
) (baseAmount sdk.Dec, err error) {
	var quoteAssetDirection pooltypes.Direction
	if side == types.Side_BUY {
		quoteAssetDirection = pooltypes.Direction_ADD_TO_POOL
	} else {
		// side == types.Side_SELL
		quoteAssetDirection = pooltypes.Direction_REMOVE_FROM_POOL
	}

	baseAmount, err = k.VpoolKeeper.SwapQuoteForBase(
		ctx, pair, quoteAssetDirection, quoteAmount, baseLimit)
	if err != nil {
		return sdk.Dec{}, err
	}

	if side == types.Side_BUY {
		return baseAmount, nil
	} else {
		// side == types.Side_SELL
		return baseAmount.Neg(), nil
	}
}
