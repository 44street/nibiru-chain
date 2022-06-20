package common

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	DenomGov        = "unibi"
	DenomColl       = "uusdc"
	DenomStable     = "unusd"
	DenomStakeToken = "stake"
	DenomTestToken  = "test"
	DenomuBTC       = "ubtc"
	DenomAxlBTC     = "axlwbtc"
	DenomAxlETH     = "axlweth"

	ModuleName = "common"

	TreasuryPoolModuleAccount = "treasury_pool"

	PairSeparator = ":"

	WhitelistedColl = []string{DenomColl}

	PairGovStable  = AssetPair{Token0: DenomGov, Token1: DenomStable}
	PairCollStable = AssetPair{Token0: DenomColl, Token1: DenomStable}
	PairTestStable = AssetPair{Token0: DenomTestToken, Token1: DenomStable}
	PairBTCStable  = AssetPair{Token0: DenomAxlBTC, Token1: DenomStable}
	PairETHStable  = AssetPair{Token0: DenomAxlETH, Token1: DenomStable}

	ErrInvalidTokenPair = sdkerrors.Register(ModuleName, 1, "invalid token pair")
)

//-----------------------------------------------------------------------------
// AssetPair

// NewAssetPair returns a new asset pair instance if the pair is valid.
// The form, "token0:token1", is expected for 'pair'.
// Use this function to return an error instead of panicking.
func NewAssetPair(pair string) (AssetPair, error) {
	split := strings.Split(pair, PairSeparator)
	splitLen := len(split)
	if splitLen != 2 {
		if splitLen == 1 {
			return AssetPair{}, sdkerrors.Wrapf(ErrInvalidTokenPair,
				"pair separator missing for pair name, %v", pair)
		} else {
			return AssetPair{}, sdkerrors.Wrapf(ErrInvalidTokenPair,
				"pair name %v must have exactly two assets, not %v", pair, splitLen)
		}
	}

	if split[0] == "" || split[1] == "" {
		return AssetPair{}, sdkerrors.Wrapf(ErrInvalidTokenPair,
			"empty token identifiers are not allowed. token0: %v, token1: %v.",
			split[0], split[1])
	}

	return AssetPair{Token0: split[0], Token1: split[1]}, nil
}

// MustNewAssetPair returns a new asset pair. It will panic if 'pair' is invalid.
// The form, "token0:token1", is expected for 'pair'.
func MustNewAssetPair(pair string) AssetPair {
	assetPair, err := NewAssetPair(pair)
	if err != nil {
		panic(err)
	}
	return assetPair
}

// SortedName is the string representation of the pair with sorted assets.
func (pair AssetPair) SortedName() string {
	return SortedPairNameFromDenoms([]string{pair.Token0, pair.Token1})
}

// Name returns the string representation of the asset pair.
func (pair AssetPair) Name() string {
	return pair.String()
}

/* String returns the string representation of the asset pair.
Note that this differs from the output of the proto-generated 'String' method.
*/
func (pair AssetPair) String() string {
	return fmt.Sprintf("%s%s%s", pair.Token0, PairSeparator, pair.Token1)
}

func (pair AssetPair) IsSortedOrder() bool {
	return pair.SortedName() == pair.String()
}

func (pair AssetPair) Inverse() AssetPair {
	return AssetPair{pair.Token1, pair.Token0}
}

func (pair AssetPair) GetBaseTokenDenom() string {
	return pair.Token0
}

func (pair AssetPair) GetQuoteTokenDenom() string {
	return pair.Token1
}

func DenomsFromPoolName(pool string) (denoms []string) {
	return strings.Split(pool, ":")
}

// SortedPairNameFromDenoms returns a sorted string representing a pool of assets
func SortedPairNameFromDenoms(denoms []string) string {
	sort.Strings(denoms) // alphabetically sort in-place
	return PairNameFromDenoms(denoms)
}

// PairNameFromDenoms returns a string representing a pool of assets in the
// exact order the denoms were given as args
func PairNameFromDenoms(denoms []string) string {
	poolName := denoms[0]
	for idx, denom := range denoms {
		if idx != 0 {
			poolName += fmt.Sprintf("%s%s", PairSeparator, denom)
		}
	}
	return poolName
}

// Validate performs a basic validation of the market params
func (pair AssetPair) Validate() error {
	if err := sdk.ValidateDenom(pair.Token1); err != nil {
		return fmt.Errorf("invalid token1 asset: %w", err)
	}
	if err := sdk.ValidateDenom(pair.Token0); err != nil {
		return fmt.Errorf("invalid token0 asset: %w", err)
	}
	return nil
}

//-----------------------------------------------------------------------------
// AssetPairs

// AssetPairs is a set of AssetPair, one per pair.
type AssetPairs []AssetPair

// NewAssetPairs constructs a new asset pair set. A panic will occur if one of
// the provided pair names is invalid.
func NewAssetPairs(pairStrings []string) (pairs AssetPairs) {
	for _, pairString := range pairStrings {
		pairs = append(pairs, MustNewAssetPair(pairString))
	}
	return pairs
}

func (pairs AssetPairs) Contains(pair AssetPair) bool {
	for _, element := range pairs {
		if (element.Token0 == pair.Token0) && (element.Token1 == pair.Token1) {
			return true
		}
	}
	return false
}

func (pairs AssetPairs) Strings() []string {
	pairsStrings := []string{}
	for _, pair := range pairs {
		pairsStrings = append(pairsStrings, pair.String())
	}
	return pairsStrings
}

func (pairs AssetPairs) Validate() error {
	seenPairs := make(map[string]bool)
	for _, pair := range pairs {
		pairID := SortedPairNameFromDenoms([]string{pair.Token0, pair.Token1})
		if seenPairs[pairID] {
			return fmt.Errorf("duplicate pair %s", pairID)
		}
		if err := pair.Validate(); err != nil {
			return err
		}
		seenPairs[pairID] = true
	}
	return nil
}

// Contains checks if a token pair is contained within 'Pairs'
func (pairs AssetPairs) ContainsAtIndex(pair AssetPair) (bool, int) {
	for idx, element := range pairs {
		if (element.Token0 == pair.Token0) && (element.Token1 == pair.Token1) {
			return true, idx
		}
	}
	return false, -1
}

type assetPairsJSON AssetPairs

// MarshalJSON implements a custom JSON marshaller for the AssetPairs type to allow
// nil AssetPairs to be encoded as empty
func (pairs AssetPairs) MarshalJSON() ([]byte, error) {
	if pairs == nil {
		return json.Marshal(assetPairsJSON(AssetPairs{}))
	}
	return json.Marshal(assetPairsJSON(pairs))
}
