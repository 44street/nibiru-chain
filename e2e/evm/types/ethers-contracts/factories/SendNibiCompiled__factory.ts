/* Autogenerated file. Do not edit manually. */
/* tslint:disable */
/* eslint-disable */
import {
  Contract,
  ContractFactory,
  ContractTransactionResponse,
  Interface,
} from "ethers";
import type { Signer, ContractDeployTransaction, ContractRunner } from "ethers";
import type { NonPayableOverrides } from "../common";
import type {
  SendNibiCompiled,
  SendNibiCompiledInterface,
} from "../SendNibiCompiled";

const _abi = [
  {
    inputs: [
      {
        internalType: "address payable",
        name: "_to",
        type: "address",
      },
    ],
    name: "sendViaCall",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address payable",
        name: "_to",
        type: "address",
      },
    ],
    name: "sendViaSend",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address payable",
        name: "_to",
        type: "address",
      },
    ],
    name: "sendViaTransfer",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
] as const;

const _bytecode =
  "0x608060405234801561001057600080fd5b50610390806100206000396000f3fe6080604052600436106100345760003560e01c8063636e082b1461003957806374be480614610055578063830c29ae14610071575b600080fd5b610053600480360381019061004e919061026a565b61008d565b005b61006f600480360381019061006a919061026a565b6100d7565b005b61008b6004803603810190610086919061026a565b610154565b005b8073ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f193505050501580156100d3573d6000803e3d6000fd5b5050565b60008173ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050905080610150576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610147906102f4565b60405180910390fd5b5050565b6000808273ffffffffffffffffffffffffffffffffffffffff163460405161017b90610345565b60006040518083038185875af1925050503d80600081146101b8576040519150601f19603f3d011682016040523d82523d6000602084013e6101bd565b606091505b509150915081610202576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101f9906102f4565b60405180910390fd5b505050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006102378261020c565b9050919050565b6102478161022c565b811461025257600080fd5b50565b6000813590506102648161023e565b92915050565b6000602082840312156102805761027f610207565b5b600061028e84828501610255565b91505092915050565b600082825260208201905092915050565b7f4661696c656420746f2073656e64204e69626900000000000000000000000000600082015250565b60006102de601383610297565b91506102e9826102a8565b602082019050919050565b6000602082019050818103600083015261030d816102d1565b9050919050565b600081905092915050565b50565b600061032f600083610314565b915061033a8261031f565b600082019050919050565b600061035082610322565b915081905091905056fea26469706673582212201fcd9f47953315963ca2a2687073914cbb3f29161100cec83979926b96714b2b64736f6c63430008180033";

type SendNibiCompiledConstructorParams =
  | [signer?: Signer]
  | ConstructorParameters<typeof ContractFactory>;

const isSuperArgs = (
  xs: SendNibiCompiledConstructorParams
): xs is ConstructorParameters<typeof ContractFactory> => xs.length > 1;

export class SendNibiCompiled__factory extends ContractFactory {
  constructor(...args: SendNibiCompiledConstructorParams) {
    if (isSuperArgs(args)) {
      super(...args);
    } else {
      super(_abi, _bytecode, args[0]);
    }
  }

  override getDeployTransaction(
    overrides?: NonPayableOverrides & { from?: string }
  ): Promise<ContractDeployTransaction> {
    return super.getDeployTransaction(overrides || {});
  }
  override deploy(overrides?: NonPayableOverrides & { from?: string }) {
    return super.deploy(overrides || {}) as Promise<
      SendNibiCompiled & {
        deploymentTransaction(): ContractTransactionResponse;
      }
    >;
  }
  override connect(runner: ContractRunner | null): SendNibiCompiled__factory {
    return super.connect(runner) as SendNibiCompiled__factory;
  }

  static readonly bytecode = _bytecode;
  static readonly abi = _abi;
  static createInterface(): SendNibiCompiledInterface {
    return new Interface(_abi) as SendNibiCompiledInterface;
  }
  static connect(
    address: string,
    runner?: ContractRunner | null
  ): SendNibiCompiled {
    return new Contract(address, _abi, runner) as unknown as SendNibiCompiled;
  }
}