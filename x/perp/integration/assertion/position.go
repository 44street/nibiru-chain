package assertion

import (
	"fmt"
	"github.com/NibiruChain/collections"
	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/x/common/asset"
	"github.com/NibiruChain/nibiru/x/perp/types"
	"github.com/NibiruChain/nibiru/x/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
)

type positionShouldBeEqual struct {
	Account          sdk.AccAddress
	Pair             asset.Pair
	ExpectedPosition types.Position
}

func (p positionShouldBeEqual) Do(app *app.NibiruApp, ctx sdk.Context) (sdk.Context, error) {
	position, err := app.PerpKeeper.Positions.Get(ctx, collections.Join(p.Pair, p.Account))
	if err != nil {
		return ctx, err
	}

	if err = types.PositionsAreEqual(&p.ExpectedPosition, &position); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func PositionShouldBeEqual(account sdk.AccAddress, pair asset.Pair, expectedPosition types.Position) testutil.Action {
	return positionShouldBeEqual{
		Account: account,
		Pair:    pair,

		ExpectedPosition: expectedPosition,
	}
}

type positionChangedEventShouldBeEqual struct {
	ExpectedEvent *types.PositionChangedEvent
}

func (p positionChangedEventShouldBeEqual) Do(_ *app.NibiruApp, ctx sdk.Context) (sdk.Context, error) {
	for _, abciEvent := range ctx.EventManager().Events() {
		if abciEvent.Type != proto.MessageName(p.ExpectedEvent) {
			continue
		}
		typedEvent, err := sdk.ParseTypedEvent(abci.Event{
			Type:       abciEvent.Type,
			Attributes: abciEvent.Attributes,
		})
		if err != nil {
			return ctx, err
		}

		theEvent, ok := typedEvent.(*types.PositionChangedEvent)
		if !ok {
			return ctx, fmt.Errorf("expected event is not of type PositionChangedEvent")
		}

		if theEvent.Pair != p.ExpectedEvent.Pair {
			return ctx, fmt.Errorf("expected pair %s, got %s", p.ExpectedEvent.Pair, theEvent.Pair)
		}

		if theEvent.TraderAddress != p.ExpectedEvent.TraderAddress {
			return ctx, fmt.Errorf("expected trader address %s, got %s", p.ExpectedEvent.TraderAddress, theEvent.TraderAddress)
		}

		if !theEvent.Margin.Equal(p.ExpectedEvent.Margin) {
			return ctx, fmt.Errorf("expected margin %s, got %s", p.ExpectedEvent.Margin, theEvent.Margin)
		}

		if !theEvent.PositionNotional.Equal(p.ExpectedEvent.PositionNotional) {
			return ctx, fmt.Errorf("expected position notional %s, got %s", p.ExpectedEvent.PositionNotional, theEvent.PositionNotional)
		}

		if !theEvent.ExchangedSize.Equal(p.ExpectedEvent.ExchangedSize) {
			return ctx, fmt.Errorf("expected exchanged size %s, got %s", p.ExpectedEvent.ExchangedSize, theEvent.ExchangedSize)
		}

		if !theEvent.ExchangedNotional.Equal(p.ExpectedEvent.ExchangedNotional) {
			return ctx, fmt.Errorf("expected exchanged notional %s, got %s", p.ExpectedEvent.ExchangedNotional, theEvent.ExchangedNotional)
		}

		if !theEvent.TransactionFee.Equal(p.ExpectedEvent.TransactionFee) {
			return ctx, fmt.Errorf("expected transaction fee %s, got %s", p.ExpectedEvent.TransactionFee, theEvent.TransactionFee)
		}

		if !theEvent.PositionSize.Equal(p.ExpectedEvent.PositionSize) {
			return ctx, fmt.Errorf("expected position size %s, got %s", p.ExpectedEvent.PositionSize, theEvent.PositionSize)
		}

		if !theEvent.RealizedPnl.Equal(p.ExpectedEvent.RealizedPnl) {
			return ctx, fmt.Errorf("expected realized pnl %s, got %s", p.ExpectedEvent.RealizedPnl, theEvent.RealizedPnl)
		}

		if !theEvent.UnrealizedPnlAfter.Equal(p.ExpectedEvent.UnrealizedPnlAfter) {
			return ctx, fmt.Errorf("expected unrealized pnl after %s, got %s", p.ExpectedEvent.UnrealizedPnlAfter, theEvent.UnrealizedPnlAfter)
		}

		if !theEvent.BadDebt.Equal(p.ExpectedEvent.BadDebt) {
			return ctx, fmt.Errorf("expected bad debt %s, got %s", p.ExpectedEvent.BadDebt, theEvent.BadDebt)
		}

		if !theEvent.MarkPrice.Equal(p.ExpectedEvent.MarkPrice) {
			return ctx, fmt.Errorf("expected mark price %s, got %s", p.ExpectedEvent.MarkPrice, theEvent.MarkPrice)
		}

		if !theEvent.FundingPayment.Equal(p.ExpectedEvent.FundingPayment) {
			return ctx, fmt.Errorf("expected funding payment %s, got %s", p.ExpectedEvent.FundingPayment, theEvent.FundingPayment)
		}

		if theEvent.BlockHeight != p.ExpectedEvent.BlockHeight {
			return ctx, fmt.Errorf("expected block height %s, got %s", p.ExpectedEvent.BlockHeight, theEvent.BlockHeight)
		}

		if theEvent.BlockTimeMs != p.ExpectedEvent.BlockTimeMs {
			return ctx, fmt.Errorf("expected block time ms %s, got %s", p.ExpectedEvent.BlockTimeMs, theEvent.BlockTimeMs)
		}
	}

	return ctx, nil
}

// PositionChangedEventShouldBeEqual checks that the position changed event is equal to the expected event.
func PositionChangedEventShouldBeEqual(expectedEvent *types.PositionChangedEvent) testutil.Action {
	return positionChangedEventShouldBeEqual{
		ExpectedEvent: expectedEvent,
	}
}
