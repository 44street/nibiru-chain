package types

import (
	"errors"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/NibiruChain/nibiru/x/common"
)

const (
	ModuleName           = "perp"
	VaultModuleAccount   = "vault"
	PerpEFModuleAccount  = "perp_ef"
	FeePoolModuleAccount = "fee_pool"
)

// x/perp module sentinel errors
var (
	ErrMarginHighEnough = sdkerrors.Register(ModuleName, 1,
		"Margin is higher than required maintenant margin ratio")
	ErrPositionNotFound = errors.New("no position found")
	ErrPairNotFound     = errors.New("pair doesn't have live vpool")
	ErrPositionZero     = errors.New("position is zero")
)

func ZeroPosition(ctx sdk.Context, tokenPair common.TokenPair, traderAddr sdk.AccAddress) *Position {
	return &Position{
		TraderAddress:                       traderAddr,
		Pair:                                tokenPair.String(),
		Size_:                               sdk.ZeroDec(),
		Margin:                              sdk.ZeroDec(),
		OpenNotional:                        sdk.ZeroDec(),
		LastUpdateCumulativePremiumFraction: sdk.ZeroDec(),
		BlockNumber:                         ctx.BlockHeight(),
	}
}

func (l *LiquidateResp) Validate() error {
	nilFieldError := fmt.Errorf(
		`invalid liquidationOutput: %v,
				must not have nil fields`, l.String())

	// nil sdk.Int check
	for _, field := range []sdk.Int{
		l.FeeToLiquidator, l.FeeToPerpEcosystemFund} {
		if field.IsNil() {
			return nilFieldError
		}
	}

	// nil sdk.Dec check
	for _, field := range []sdk.Dec{l.BadDebt} {
		if field.IsNil() {
			return nilFieldError
		}
	}

	_, err := sdk.AccAddressFromBech32(l.Liquidator.String())
	if err != nil {
		return err
	}

	return nil
}
