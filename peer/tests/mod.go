package tests

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl"
	"go.dedis.ch/cs438/permissioned-chain"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/transport/channel"
	"go.dedis.ch/cs438/types"
)

var peerFac peer.Factory = impl.NewPeer

func Setup_n_peers_bc_perf(t *testing.T, transp transport.Transport, n int, maxTxn int,
	timeout string, gains []float64, waitTime time.Duration,
	disableMPC bool, disablePubkeyTxn bool) ([]*z.TestNode, []string) {
	nodes := make([]*z.TestNode, n)

	opts := []z.Option{
		z.WithMPCMaxWaitBlock(1),
	}

	if disableMPC {
		opts = append(opts, z.WithDisableMPC())
	}
	if disablePubkeyTxn {
		opts = append(opts, z.WithDisableAnnonceEnckey())
	}
	for i := 0; i < n; i++ {
		antiAntroppyOpt := z.WithAntiEntropy(time.Second * time.Duration(5+rand.Intn(5)))
		nodeOpts := append(opts, antiAntroppyOpt)
		node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			nodeOpts...)
		nodes[i] = &node
	}

	// generate key pairs
	addrs := make([]string, n)
	for i := 0; i < n; i++ {
		privkey1, err := crypto.GenerateKey()
		require.NoError(t, err)
		nodes[i].BCSetKeyPair(*privkey1)
		addr, err := nodes[i].BCGetAddress()
		require.NoError(t, err)
		addrs[i] = addr.Hex
		fmt.Printf("-----%s : %s--------\n", nodes[i].GetAddr(), addr)
	}

	// get encryption pubkeys
	pubkeys := make([]types.Pubkey, n)
	for i := 0; i < n; i++ {
		pubkeys[i] = nodes[i].GetPubkeyStore()[nodes[i].GetAddr()]
	}

	// add peer
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			nodes[i].AddPeer(nodes[j].GetAddr())
			// nodes[i].SetPubkeyEntry(nodes[j].GetAddr(), &pubkeys[j])
		}
	}

	// > init blockchain. Should success
	// all should have the block
	participants := make(map[string]string)
	if disablePubkeyTxn {
		for i, addr := range addrs {
			pubBytes, err := x509.MarshalPKIXPublicKey((*rsa.PublicKey)(&pubkeys[i]))
			require.NoError(t, err)
			participants[addr] = hex.EncodeToString(pubBytes)
		}
	} else {
		for _, addr := range addrs {
			participants[addr] = ""
		}
	}

	config := permissioned.NewChainConfig(
		participants,
		maxTxn, timeout, 1, 1,
	)
	initialGain := make(map[string]float64)
	for i, gain := range gains {
		initialGain[addrs[i]] = gain
	}

	err := nodes[0].InitBlockchain(*config, initialGain)
	require.NoError(t, err)

	time.Sleep(waitTime)

	for _, node := range nodes {
		block0 := node.BCGetLatestBlock()
		require.NotNil(t, block0)
		require.Equal(t, uint(0), block0.Height)
	}

	return nodes, addrs
}

func Setup_n_peers_bc(t *testing.T, n int, maxTxn int,
	timeout string, gains []float64, disableMPC bool, disablePubkeyTxn bool) ([]*z.TestNode, []string) {
	transp := channel.NewTransport()
	nodes := make([]*z.TestNode, n)

	opt := []z.Option{
		z.WithMPCMaxWaitBlock(1),
	}

	if disableMPC {
		opt = append(opt, z.WithDisableMPC())
	}
	if disablePubkeyTxn {
		opt = append(opt, z.WithDisableAnnonceEnckey())
	}
	for i := 0; i < n; i++ {
		node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			opt...)
		nodes[i] = &node
	}

	// generate key pairs
	addrs := make([]string, n)
	for i := 0; i < n; i++ {
		privkey1, err := crypto.GenerateKey()
		require.NoError(t, err)
		nodes[i].BCSetKeyPair(*privkey1)
		addr, err := nodes[i].BCGetAddress()
		require.NoError(t, err)
		addrs[i] = addr.Hex
		fmt.Printf("-----%s : %s--------\n", nodes[i].GetAddr(), addr)
	}

	// get encryption pubkeys
	pubkeys := make([]types.Pubkey, n)
	for i := 0; i < n; i++ {
		pubkeys[i] = nodes[i].GetPubkeyStore()[nodes[i].GetAddr()]
	}

	// add peer
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			nodes[i].AddPeer(nodes[j].GetAddr())
			// nodes[i].SetPubkeyEntry(nodes[j].GetAddr(), &pubkeys[j])
		}
	}

	// > init blockchain. Should success
	// all should have the block
	participants := make(map[string]string)
	if disablePubkeyTxn {
		for i, addr := range addrs {
			pubBytes, err := x509.MarshalPKIXPublicKey((*rsa.PublicKey)(&pubkeys[i]))
			require.NoError(t, err)
			participants[addr] = hex.EncodeToString(pubBytes)
		}
	} else {
		for _, addr := range addrs {
			participants[addr] = ""
		}
	}

	config := permissioned.NewChainConfig(
		participants,
		maxTxn, timeout, 1, 1,
	)
	initialGain := make(map[string]float64)
	for i, gain := range gains {
		initialGain[addrs[i]] = gain
	}

	err := nodes[0].InitBlockchain(*config, initialGain)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 500)

	for _, node := range nodes {
		block0 := node.BCGetLatestBlock()
		require.NotNil(t, block0)
		require.Equal(t, uint(0), block0.Height)
	}

	return nodes, addrs
}

func Setup_n_peers(n int, t *testing.T, opt ...z.Option) []z.TestNode {
	nodes := make([]z.TestNode, n)

	transp := channel.NewTransport()

	for i := 0; i < n; i++ {
		node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithMPCPaxos(),
			z.WithTotalPeers(uint(n)), z.WithPaxosID(uint(i+1)))
		nodes[i] = node
	}

	pubkeys := make([]types.Pubkey, n)
	for i := 0; i < n; i++ {
		pubkeys[i] = nodes[i].GetPubkeyStore()[nodes[i].GetAddr()]
	}

	// add peer & setPubkey
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			nodes[i].AddPeer(nodes[j].GetAddr())
			nodes[i].SetPubkeyEntry(nodes[j].GetAddr(), &pubkeys[j])
		}
	}
	return nodes
}
