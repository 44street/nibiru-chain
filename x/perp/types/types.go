package types

import (
	"errors"

	"github.com/NibiruChain/nibiru/x/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName           = "perp"
	VaultModuleAccount   = "vault"
	PerpEFModuleAccount  = "perp_ef"
	FeePoolModuleAccount = "fee_pool"
)

var (
	ErrNotFound = errors.New("not found")
)

func ZeroPosition(ctx sdk.Context, vpair common.TokenPair, trader string) *Position {
	return &Position{
		Address:                             trader,
		Pair:                                vpair.String(),
		Size_:                               sdk.ZeroDec(),
		Margin:                              sdk.ZeroDec(),
		OpenNotional:                        sdk.ZeroDec(),
		LastUpdateCumulativePremiumFraction: sdk.ZeroDec(),
		LiquidityHistoryIndex:               0,
		BlockNumber:                         ctx.BlockHeight(),
	}
}
