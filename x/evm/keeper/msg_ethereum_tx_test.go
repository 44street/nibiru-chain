package keeper_test

import (
	"math/big"
	"strconv"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	gethcore "github.com/ethereum/go-ethereum/core/types"
	gethparams "github.com/ethereum/go-ethereum/params"

	"github.com/NibiruChain/nibiru/x/common/testutil"
	"github.com/NibiruChain/nibiru/x/common/testutil/testapp"
	"github.com/NibiruChain/nibiru/x/evm/embeds"

	"github.com/NibiruChain/nibiru/x/evm/evmtest"
)

func (s *Suite) TestMsgEthereumTx_CreateContract() {
	testCases := []struct {
		name     string
		scenario func()
	}{
		{
			name: "happy: deploy contract, sufficient gas limit",
			scenario: func() {
				deps := evmtest.NewTestDeps()
				ethAcc := deps.Sender

				// Leftover gas fee is refunded within ApplyEvmTx from the FeeCollector
				// so, the module must have some coins
				err := testapp.FundModuleAccount(
					deps.Chain.BankKeeper,
					deps.Ctx,
					authtypes.FeeCollectorName,
					sdk.NewCoins(sdk.NewCoin("unibi", math.NewInt(1000_000))),
				)
				s.Require().NoError(err)
				s.T().Log("create eth tx msg, increase gas limit")
				gasLimit := big.NewInt(1000_000)
				args := evmtest.ArgsCreateContract{
					EthAcc:        ethAcc,
					EthChainIDInt: deps.K.EthChainID(deps.Ctx),
					GasPrice:      big.NewInt(1),
					Nonce:         deps.StateDB().GetNonce(ethAcc.EthAddr),
					GasLimit:      gasLimit,
				}
				ethTxMsg, err := evmtest.CreateContractTxMsg(args)
				s.Require().NoError(err)
				s.Require().NoError(ethTxMsg.ValidateBasic())
				s.Equal(ethTxMsg.GetGas(), gasLimit.Uint64())

				resp, err := deps.Chain.EvmKeeper.EthereumTx(deps.GoCtx(), ethTxMsg)
				s.Require().NoError(
					err,
					"resp: %s\nblock header: %s",
					resp,
					deps.Ctx.BlockHeader().ProposerAddress,
				)
				s.Require().Empty(resp.VmError)

				// Event "EventContractDeployed" must present
				var sdkEvents = deps.Ctx.EventManager().Events()
				contractDeployedEventType := "eth.evm.v1.EventContractDeployed"
				err = testutil.AssertEventPresent(sdkEvents, contractDeployedEventType)
				s.Require().NoError(err)

				var contractDeployedEvent sdk.Event
				for _, abciEvent := range sdkEvents {
					if abciEvent.Type == contractDeployedEventType {
						contractDeployedEvent = abciEvent
					}
				}
				for _, err = range []error{
					testutil.EventHasAttributeValue(contractDeployedEvent, "sender", ethAcc.EthAddr.String()),
					testutil.EventHasAttributeValue(contractDeployedEvent, "contract_addr", resp.Logs[0].Address),
				} {
					s.Require().NoError(err)
				}
			},
		},
		{
			name: "sad: deploy contract, exceed gas limit",
			scenario: func() {
				deps := evmtest.NewTestDeps()
				ethAcc := deps.Sender

				s.T().Log("create eth tx msg, default create contract gas")
				gasLimit := gethparams.TxGasContractCreation
				args := evmtest.ArgsCreateContract{
					EthAcc:        ethAcc,
					EthChainIDInt: deps.K.EthChainID(deps.Ctx),
					GasPrice:      big.NewInt(1),
					Nonce:         deps.StateDB().GetNonce(ethAcc.EthAddr),
				}
				ethTxMsg, err := evmtest.CreateContractTxMsg(args)
				s.NoError(err)
				s.Require().NoError(ethTxMsg.ValidateBasic())
				s.Equal(ethTxMsg.GetGas(), gasLimit)

				resp, err := deps.Chain.EvmKeeper.EthereumTx(deps.GoCtx(), ethTxMsg)
				s.Require().ErrorContains(
					err,
					core.ErrIntrinsicGas.Error(),
					"resp: %s\nblock header: %s",
					resp,
					deps.Ctx.BlockHeader().ProposerAddress,
				)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, tc.scenario)
	}
}

func (s *Suite) TestMsgEthereumTx_ExecuteContract() {
	deps := evmtest.NewTestDeps()
	ethAcc := deps.Sender

	// Leftover gas fee is refunded within ApplyEvmTx from the FeeCollector
	// so, the module must have some coins
	err := testapp.FundModuleAccount(
		deps.Chain.BankKeeper,
		deps.Ctx,
		authtypes.FeeCollectorName,
		sdk.NewCoins(sdk.NewCoin("unibi", math.NewInt(1000_000))),
	)
	s.Require().NoError(err)
	deployResp, err := evmtest.DeployContract(
		&deps, embeds.SmartContract_TestERC20, s.T(),
	)
	s.Require().NoError(err)
	contractAddr := deployResp.ContractAddr
	testContract, err := embeds.SmartContract_TestERC20.Load()
	s.Require().NoError(err)
	to := gethcommon.HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	input, err := testContract.ABI.Pack("transfer", to, big.NewInt(123))
	s.NoError(err)

	gasLimit := big.NewInt(1000_000)
	args := evmtest.ArgsExecuteContract{
		EthAcc:          ethAcc,
		EthChainIDInt:   deps.K.EthChainID(deps.Ctx),
		GasPrice:        big.NewInt(1),
		Nonce:           deps.StateDB().GetNonce(ethAcc.EthAddr),
		GasLimit:        gasLimit,
		ContractAddress: &contractAddr,
		Data:            input,
	}
	ethTxMsg, err := evmtest.ExecuteContractTxMsg(args)
	s.NoError(err)
	s.Require().NoError(ethTxMsg.ValidateBasic())
	s.Equal(ethTxMsg.GetGas(), gasLimit.Uint64())
	resp, err := deps.Chain.EvmKeeper.EthereumTx(deps.GoCtx(), ethTxMsg)
	s.Require().NoError(
		err,
		"resp: %s\nblock header: %s",
		resp,
		deps.Ctx.BlockHeader().ProposerAddress,
	)
	s.Require().Empty(resp.VmError)

	// Event "EventContractExecuted" must present
	var sdkEvents = deps.Ctx.EventManager().Events()
	contractExecutedEventType := "eth.evm.v1.EventContractDeployed"
	err = testutil.AssertEventPresent(sdkEvents, contractExecutedEventType)
	s.Require().NoError(err)

	var contractExecutedEvent sdk.Event
	for _, abciEvent := range sdkEvents {
		if abciEvent.Type == contractExecutedEventType {
			contractExecutedEvent = abciEvent
		}
	}
	for _, err = range []error{
		testutil.EventHasAttributeValue(contractExecutedEvent, "sender", ethAcc.EthAddr.String()),
		testutil.EventHasAttributeValue(contractExecutedEvent, "contract_addr", resp.Logs[0].Address),
	} {
		s.Require().NoError(err)
	}
}

func (s *Suite) TestMsgEthereumTx_SimpleTransfer() {
	testCases := []struct {
		name   string
		txType evmtest.GethTxType
	}{
		{
			name:   "happy: AccessListTx",
			txType: gethcore.AccessListTxType,
		},
		{
			name:   "happy: LegacyTx",
			txType: gethcore.LegacyTxType,
		},
	}

	for _, tc := range testCases {
		deps := evmtest.NewTestDeps()
		ethAcc := deps.Sender

		amount := int64(123)
		err := testapp.FundAccount(
			deps.Chain.BankKeeper,
			deps.Ctx,
			deps.Sender.NibiruAddr,
			sdk.NewCoins(sdk.NewInt64Coin("unibi", amount)),
		)
		s.Require().NoError(err)

		s.T().Log("create eth tx msg")
		var innerTxData []byte = nil
		var accessList gethcore.AccessList = nil
		to := gethcommon.HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")

		ethTxMsg, err := evmtest.NewEthTxMsgFromTxData(
			&deps,
			tc.txType,
			innerTxData,
			deps.StateDB().GetNonce(ethAcc.EthAddr),
			&to,
			big.NewInt(amount),
			gethparams.TxGas,
			accessList,
		)
		s.NoError(err)

		resp, err := deps.Chain.EvmKeeper.EthereumTx(deps.GoCtx(), ethTxMsg)
		s.Require().NoError(err)
		s.Require().Empty(resp.VmError)

		gasUsed := strconv.FormatUint(resp.GasUsed, 10)
		wantGasUsed := strconv.FormatUint(gethparams.TxGas, 10)
		s.Equal(gasUsed, wantGasUsed)

		// Event "EventContractDeployed" must present
		var sdkEvents = deps.Ctx.EventManager().Events()
		evmTransferEventType := "eth.evm.v1.EventTransfer"
		err = testutil.AssertEventPresent(sdkEvents, evmTransferEventType)
		s.Require().NoError(err)

		var evmTransferEvent sdk.Event
		for _, abciEvent := range sdkEvents {
			if abciEvent.Type == evmTransferEventType {
				evmTransferEvent = abciEvent
			}
		}
		for _, err = range []error{
			testutil.EventHasAttributeValue(evmTransferEvent, "sender", ethAcc.EthAddr.String()),
			testutil.EventHasAttributeValue(evmTransferEvent, "recipient", to.String()),
			testutil.EventHasAttributeValue(evmTransferEvent, "amount", strconv.FormatInt(amount, 10)),
		} {
			s.Require().NoError(err)
		}
	}
}
