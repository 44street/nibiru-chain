package common_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/nibiru/x/common"
)

func TestPairNameFromDenoms(t *testing.T) {
	testCases := []struct {
		name     string
		denoms   []string
		poolName string
	}{
		{
			name:     "ATOM:OSMO in correct order",
			denoms:   []string{"atom", "osmo"},
			poolName: "atom:osmo",
		},
		{
			name:     "ATOM:OSMO in wrong order",
			denoms:   []string{"osmo", "atom"},
			poolName: "atom:osmo",
		},
		{
			name:     "X:Y:Z in correct order",
			denoms:   []string{"x", "y", "z"},
			poolName: "x:y:z",
		},
		{
			name:     "X:Y:Z in wrong order",
			denoms:   []string{"z", "x", "y"},
			poolName: "x:y:z",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			outPoolName := common.PoolNameFromDenoms(tc.denoms)
			require.Equal(t, tc.poolName, outPoolName)
		})
	}
}

func TestAssetPair(t *testing.T) {
	testCases := []struct {
		name   string
		pair   common.AssetPair
		proper bool
	}{
		{
			name:   "proper and improper order pairs are inverses-1",
			pair:   common.AssetPair{Token0: "atom", Token1: "osmo"},
			proper: true,
		},
		{
			name:   "proper and improper order pairs are inverses-2",
			pair:   common.AssetPair{Token0: "osmo", Token1: "atom"},
			proper: false,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			if tc.proper {
				require.True(t, tc.pair.IsProperOrder())
				require.Equal(t, tc.pair.SortedName(), tc.pair.String())
			} else {
				require.True(t, tc.pair.Inverse().IsProperOrder())
				require.Equal(t, tc.pair.SortedName(), tc.pair.Inverse().String())
			}

			require.True(t, true)
		})
	}
}

func TestAsset_Constructor(t *testing.T) {
	tests := []struct {
		name      string
		tokenPair string
		err       error
	}{
		{
			"only one token",
			common.GovDenom,
			common.ErrInvalidTokenPair,
		},
		{
			"more than 2 tokens",
			fmt.Sprintf("%s%s%s%s%s", common.GovDenom, common.PairSeparator, common.StableDenom,
				common.PairSeparator, common.CollDenom),
			common.ErrInvalidTokenPair,
		},
		{
			"different separator",
			fmt.Sprintf("%s%s%s", common.GovDenom, "%", common.StableDenom),
			common.ErrInvalidTokenPair,
		},
		{
			"correct pair",
			fmt.Sprintf("%s%s%s", common.GovDenom, common.PairSeparator, common.StableDenom),
			nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := common.NewAssetPairFromStr(tc.tokenPair)
			if tc.err != nil {
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAsset_GetQuoteBaseToken(t *testing.T) {
	pair, err := common.NewAssetPairFromStr("uatom:unibi")
	require.NoError(t, err)

	require.Equal(t, "uatom", pair.BaseDenom())
	require.Equal(t, "unibi", pair.QuoteDenom())
}
