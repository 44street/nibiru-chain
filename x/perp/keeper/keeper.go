package keeper

import (
	"fmt"

	"github.com/NibiruChain/nibiru/coll"

	"github.com/NibiruChain/nibiru/x/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/NibiruChain/nibiru/x/perp/types"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      sdk.StoreKey
	ParamSubspace paramtypes.Subspace

	BankKeeper      types.BankKeeper
	AccountKeeper   types.AccountKeeper
	PricefeedKeeper types.PricefeedKeeper
	VpoolKeeper     types.VpoolKeeper
	EpochKeeper     types.EpochKeeper

	Positions      coll.Map[coll.Pair[common.AssetPair, sdk.AccAddress], types.Position]
	PairsMetadata  coll.Map[common.AssetPair, types.PairMetadata]
	PrepaidBadDebt coll.Map[string, types.PrepaidBadDebt]
}

// NewKeeper Creates a new x/perp Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	paramSubspace paramtypes.Subspace,

	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	priceKeeper types.PricefeedKeeper,
	vpoolKeeper types.VpoolKeeper,
	epochKeeper types.EpochKeeper,
) Keeper {
	// Ensure that the module account is set.
	if moduleAcc := accountKeeper.GetModuleAddress(types.ModuleName); moduleAcc == nil {
		panic("The x/perp module account has not been set")
	}

	// Set param.types.'KeyTable' if it has not already been set
	if !paramSubspace.HasKeyTable() {
		paramSubspace = paramSubspace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		ParamSubspace:   paramSubspace,
		BankKeeper:      bankKeeper,
		AccountKeeper:   accountKeeper,
		PricefeedKeeper: priceKeeper,
		VpoolKeeper:     vpoolKeeper,
		EpochKeeper:     epochKeeper,
		Positions: coll.NewMap(
			storeKey, 0,
			coll.PairKeyEncoder[common.AssetPair, sdk.AccAddress](common.AssetPairKeyEncoder, coll.Keys.AccAddress),
			coll.ProtoValueEncoder[types.Position](cdc),
		),
		PairsMetadata:  coll.NewMap[common.AssetPair, types.PairMetadata](storeKey, 1, common.AssetPairKeyEncoder, coll.ProtoValueEncoder[types.PairMetadata](cdc)),
		PrepaidBadDebt: coll.NewMap[string, types.PrepaidBadDebt](storeKey, 2, coll.Keys.String, coll.ProtoValueEncoder[types.PrepaidBadDebt](cdc)),
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.ParamSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.ParamSubspace.SetParamSet(ctx, &params)
}
