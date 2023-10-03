package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/nibiru/x/tokenfactory/types"

	"github.com/MakeNowJust/heredoc/v2"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		Aliases:                    []string{"tf"},
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreateDenom(),
		CmdChangeAdmin(),
		CmdMint(),
		CmdBurn(),
		// CmdModifyDenomMetadata(),
	)

	return cmd
}

// CmdCreateDenom broadcast MsgCreateDenom
func CmdCreateDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-denom [subdenom] [flags]",
		Short: "create a new denom from an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			txFactory = txFactory.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(
				clientCtx.AccountRetriever)

			msg := &types.MsgCreateDenom{
				Sender:   clientCtx.GetFromAddress().String(),
				Subdenom: args[0],
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txFactory, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdChangeAdmin: Broadcasts MsgChangeAdmin
func CmdChangeAdmin() *cobra.Command {
	cmd := &cobra.Command{
		Use: "change-admin [denom] [new-admin] [flags]",
		Short: heredoc.Doc(`
			Changes the admin address for a token factory denom.
			Must have admin authority to do so.
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txFactory = txFactory.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			msg := &types.MsgChangeAdmin{
				Sender:   clientCtx.GetFromAddress().String(),
				Denom:    args[0],
				NewAdmin: args[1],
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txFactory, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdMint: Broadcast MsgMint
func CmdMint() *cobra.Command {
	cmd := &cobra.Command{
		Use: "mint [coin] [--mint-to] [flags]",
		Short: heredoc.Doc(`
			Mint a denom to an address.
			Tx signer must be the denom admin.
			If no --mint-to address is provided, it defaults to the sender.`,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			mintTo, err := cmd.Flags().GetString("mint-to")
			if err != nil {
				return fmt.Errorf(
					"Please provide a valid address using the --mint-to flag: %s", err)
			}
			mintToAddr, err := sdk.AccAddressFromBech32(mintTo)
			if err != nil {
				return err
			}

			msg := &types.MsgMint{
				Sender: clientCtx.GetFromAddress().String(),
				Coin:   coin,
				MintTo: mintToAddr.String(),
			}

			txFactory = txFactory.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txFactory, msg)
		},
	}

	cmd.Flags().String("mint-to", "", "Address to mint to")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdBurn: Broadcast MsgBurn
func CmdBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use: "burn [coin] [--burn-from] [flags]",
		Short: heredoc.Doc(`
			Burn tokens from an address.
			Tx signer must be the denom admin.
			If no --burn-from address is provided, it defaults to the sender.`,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txFactory = txFactory.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			burnFrom, err := cmd.Flags().GetString("burn-from")
			if err != nil {
				return fmt.Errorf(
					"Please provide a valid address using the --burn-from flag: %s", err)
			}

			burnFromAddr, err := sdk.AccAddressFromBech32(burnFrom)
			if err != nil {
				return err
			}
			msg := &types.MsgBurn{
				Sender:   clientCtx.GetFromAddress().String(),
				Coin:     coin,
				BurnFrom: burnFromAddr.String(),
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txFactory, msg)
		},
	}

	cmd.Flags().String("burn-from", "", "Address to burn from")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
