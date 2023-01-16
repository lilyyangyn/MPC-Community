# MPC Community

A p2p community with access control, where participants can start MPCs, sell data for MPC to use without worried about leackage, participate in MPC to earn fees, etc.

## System Architecture

## Quick setup
>>>>>>> Stashed changes

Install go >= 1.19.

To setup a permissioned blockchain, you should set address for the node and add the address to the chain config file. 
- If you do not have a valid address, you can ask the node to generate a new one and store the private key in the disk. Otherwise, you should provide the path to the private key file to the node, which will automatically generate the address based on the private key
- The config file is the initial configuration of the permissioned chain that is stored in the genesis block. Things you could specify:
  - **Participant List**: Node will not get any blockchain and MPC messages if it is not in the permissioned chain's participant list.
  - **Maximal number of transactions**: the maximal number of transactions in a block except for the genesis block
  - **Maximal waiting time**: the maximal waiting time for a miner to wait for the next transaction before it produces a new block. This works only when miner has at least one transaction for the next block to be mined
  - **MPC basic gain**: This is the minimal amount of coins nodes can earn in a single MPC Calculation. It can also earn extra coins if its value is used in that MPC Calculation.
- an example of `config.yaml` and some key files are provided for quick setup

#### Run a node with CLI:

```sh
go run main.go cli
```

#### Run a node with Web GUI:

```sh
go run main.go daemon
```


### Run the tests

See commands in the Makefile. For example: `make test`.
