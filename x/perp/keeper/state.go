package keeper

import (
	"context"
	"fmt"

	"github.com/NibiruChain/nibiru/x/common"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/NibiruChain/nibiru/x/perp/types"
)

func (k Keeper) Params(
	goCtx context.Context, req *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Positions() PositionsState {
	return (PositionsState)(k)
}

func (k Keeper) PairMetadata() PairMetadata {
	return (PairMetadata)(k)
}

func (k Keeper) Whitelist() Whitelist {
	return (Whitelist)(k)
}

func (k Keeper) PrepaidBadDebtState() PrepaidBadDebtState {
	return (PrepaidBadDebtState)(k)
}

var paramsNamespace = []byte{0x0}
var paramsKey = []byte{0x0}

type ParamsState Keeper

func (p ParamsState) getKV(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(p.storeKey), paramsNamespace)
}

func (p ParamsState) Get(ctx sdk.Context) (*types.Params, error) {
	kv := p.getKV(ctx)

	value := kv.Get(paramsKey)
	if value == nil {
		return nil, fmt.Errorf("not found")
	}

	params := new(types.Params)
	p.cdc.MustUnmarshal(value, params)
	return params, nil
}

func (p ParamsState) Set(ctx sdk.Context, params *types.Params) {
	kv := p.getKV(ctx)
	kv.Set(paramsKey, p.cdc.MustMarshal(params))
}

var positionsNamespace = []byte{0x1}

type PositionsState Keeper

func (p PositionsState) getKV(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(p.storeKey), positionsNamespace)
}

func (p PositionsState) keyFromType(position *types.Position) []byte {
	traderAddress, err := sdk.AccAddressFromBech32(position.TraderAddress)
	if err != nil {
		panic(err)
	}
	return p.keyFromRaw(position.GetAssetPair(), traderAddress)
}

func (p PositionsState) keyFromRaw(pair common.AssetPair, address sdk.AccAddress) []byte {
	// TODO(mercilex): not sure if namespace overlap safe | update(mercilex) it is not overlap safe
	return []byte(pair.String() + address.String())
}

func (p PositionsState) Create(ctx sdk.Context, position *types.Position) error {
	key := p.keyFromType(position)
	kv := p.getKV(ctx)
	if kv.Has(key) {
		return fmt.Errorf("already exists")
	}

	kv.Set(key, p.cdc.MustMarshal(position))
	return nil
}

func (p PositionsState) Get(ctx sdk.Context, pair common.AssetPair, traderAddr sdk.AccAddress) (*types.Position, error) {
	kv := p.getKV(ctx)

	key := p.keyFromRaw(pair, traderAddr)
	valueBytes := kv.Get(key)
	if valueBytes == nil {
		return nil, types.ErrPositionNotFound
	}

	position := new(types.Position)
	p.cdc.MustUnmarshal(valueBytes, position)

	return position, nil
}

func (p PositionsState) Update(ctx sdk.Context, position *types.Position) error {
	kv := p.getKV(ctx)
	key := p.keyFromType(position)

	if !kv.Has(key) {
		return types.ErrPositionNotFound
	}

	kv.Set(key, p.cdc.MustMarshal(position))
	return nil
}

func (p PositionsState) Set(
	ctx sdk.Context, pair common.AssetPair, traderAddr sdk.AccAddress, position *types.Position,
) {
	positionID := p.keyFromRaw(pair, traderAddr)
	kvStore := p.getKV(ctx)
	kvStore.Set(positionID, p.cdc.MustMarshal(position))
}

func (p PositionsState) Iterate(ctx sdk.Context, do func(position *types.Position) (stop bool)) {
	store := p.getKV(ctx)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		position := new(types.Position)
		p.cdc.MustUnmarshal(iter.Value(), position)
		if !do(position) {
			break
		}
	}
}

func (p PositionsState) Delete(ctx sdk.Context, pair common.AssetPair, addr sdk.AccAddress) error {
	store := p.getKV(ctx)
	primaryKey := p.keyFromRaw(pair, addr)

	if !store.Has(primaryKey) {
		return types.ErrPositionNotFound.Wrapf("in pair %s", pair)
	}
	store.Delete(primaryKey)

	return nil
}

var pairMetadataNamespace = []byte{0x2}

type PairMetadata Keeper

func (p PairMetadata) getKV(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(p.storeKey), pairMetadataNamespace)
}

func (p PairMetadata) Get(ctx sdk.Context, pair common.AssetPair) (*types.PairMetadata, error) {
	kv := p.getKV(ctx)

	v := kv.Get([]byte(pair.String()))
	if v == nil {
		return nil, types.ErrPairMetadataNotFound
	}

	pairMetadata := new(types.PairMetadata)
	p.cdc.MustUnmarshal(v, pairMetadata)

	return pairMetadata, nil
}

func (p PairMetadata) Set(ctx sdk.Context, metadata *types.PairMetadata) {
	kv := p.getKV(ctx)
	kv.Set([]byte(metadata.Pair), p.cdc.MustMarshal(metadata))
}

func (p PairMetadata) GetAll(ctx sdk.Context) []*types.PairMetadata {
	store := ctx.KVStore(p.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, pairMetadataNamespace)

	var pairMetadatas []*types.PairMetadata
	for ; iterator.Valid(); iterator.Next() {
		var pairMetadata = new(types.PairMetadata)
		p.cdc.MustUnmarshal(iterator.Value(), pairMetadata)
		pairMetadatas = append(pairMetadatas, pairMetadata)
	}

	return pairMetadatas
}

var whitelistNamespace = []byte{0x3}

type Whitelist Keeper

func (w Whitelist) getKV(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(w.storeKey), whitelistNamespace)
}

func (w Whitelist) IsWhitelisted(ctx sdk.Context, address sdk.AccAddress) bool {
	kv := w.getKV(ctx)

	return kv.Has(address)
}

func (w Whitelist) Whitelist(ctx sdk.Context, address sdk.AccAddress) {
	kv := w.getKV(ctx)
	kv.Set(address, []byte{})
}

func (w Whitelist) Iterate(ctx sdk.Context, do func(addr sdk.AccAddress) (stop bool)) {
	kv := w.getKV(ctx)
	iter := kv.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if !do(iter.Key()) {
			break
		}
	}
}

var prepaidBadDebtNamespace = []byte{0x4}

type PrepaidBadDebtState Keeper

func (pbd PrepaidBadDebtState) getKVStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(pbd.storeKey), prepaidBadDebtNamespace)
}

/*
Fetches the amount of bad debt prepaid by denom. Returns zero if the denom is not found.
*/
func (pbd PrepaidBadDebtState) Get(ctx sdk.Context, denom string) (amount sdk.Int) {
	kv := pbd.getKVStore(ctx)

	v := kv.Get([]byte(denom))
	if v == nil {
		return sdk.ZeroInt()
	}

	err := amount.Unmarshal(v)
	if err != nil {
		panic(err)
	}

	return amount
}

func (pbd PrepaidBadDebtState) Iterate(ctx sdk.Context, do func(denom string, amount sdk.Int) (stop bool)) {
	kv := pbd.getKVStore(ctx)
	iter := kv.Iterator(nil, nil)

	for ; iter.Valid(); iter.Next() {
		amount := sdk.Int{}
		err := amount.Unmarshal(iter.Value())
		if err != nil {
			panic(err)
		}
		if !do(string(iter.Key()), amount) {
			break
		}
	}
}

/*
Sets the amount of bad debt prepaid by denom.
*/
func (pbd PrepaidBadDebtState) Set(ctx sdk.Context, denom string, amount sdk.Int) {
	kv := pbd.getKVStore(ctx)
	b, err := amount.Marshal()
	if err != nil {
		panic(err)
	}
	kv.Set([]byte(denom), b)
}

/*
Increments the amount of bad debt prepaid by denom.
Calling this method on a denom that doesn't exist is effectively the same as setting the amount (0 + increment).
*/
func (pbd PrepaidBadDebtState) Increment(ctx sdk.Context, denom string, increment sdk.Int) (
	amount sdk.Int,
) {
	kv := pbd.getKVStore(ctx)
	amount = pbd.Get(ctx, denom).Add(increment)

	b, err := amount.Marshal()
	if err != nil {
		panic(err)
	}
	kv.Set([]byte(denom), b)

	return amount
}

/*
Decrements the amount of bad debt prepaid by denom.

The lowest it can be decremented to is zero. Trying to decrement a prepaid bad
debt balance to below zero will clip it at zero.

*/
func (pbd PrepaidBadDebtState) Decrement(ctx sdk.Context, denom string, decrement sdk.Int) (
	amount sdk.Int,
) {
	kv := pbd.getKVStore(ctx)
	amount = sdk.MaxInt(pbd.Get(ctx, denom).Sub(decrement), sdk.ZeroInt())

	b, err := amount.Marshal()
	if err != nil {
		panic(err)
	}
	kv.Set([]byte(denom), b)

	return amount
}
