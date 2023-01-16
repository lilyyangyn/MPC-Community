# MPC Community

A p2p community with access control, where participants can start MPCs, sell data for MPC to use without worrying about leakage, participate in MPC to earn fees, etc.

## System Architecture

The system is constituted by three main components: Message module, Permissioned Blockchain module, and MPC module. Modules communicate with each other internally and all connections to the outside wolrd goes through Message module.
![system-arch](/docs/system-arch.png)

The full process of a paid MPC goes like the following. It contains three stages: PreMPC, MPC, and PostMPC. Both PreMPC and PostMPC involves Blockchain interactions, while MPC is the where the true MPC goes on.
![mpc-proccess](/docs/mpc-process.png)

For more information, please refer to the report and [slides](/docs/slides.pdf).

## Quick setup

Install go >= 1.19.

To setup a permissioned blockchain, you should set address for the node and add the address to the chain config file. 
- If you do not have a valid address, you can ask the node to generate a new one and store the private key in the disk. Otherwise, you should provide the path to the private key file to the node, which will automatically generate the address based on the private key
- The config file is the initial configuration of the permissioned chain that is stored in the genesis block. Things you could specify:
  - **Participant List**: Node will not get any blockchain and MPC messages if it is not in the permissioned chain's participant list.
  - **Maximal number of transactions**: the maximal number of transactions in a block except for the genesis block
  - **Maximal waiting time**: the maximal waiting time for a miner to wait for the next transaction before it produces a new block. This works only when miner has at least one transaction for the next block to be mined
  - **MPC basic gain**: This is the minimal amount of coins nodes can earn in a single MPC Calculation. It can also earn extra coins if its value is used in that MPC Calculation.
- an example of `config.yaml` and three key files are provided for quickly setting up a THREE-node network

#### Run a node with the interactive CLI tool:

```sh
go run main.go cli -p <port>
```

#### Run a node with Web GUI:

```sh
go run main.go daemon -p <port>
```


### Run the tests

See commands in the Makefile. For example: `make test`.
