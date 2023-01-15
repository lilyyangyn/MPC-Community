package perf

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/tests"
	"go.dedis.ch/cs438/permissioned-chain"
)

// --------------------------------- benchmark --------------------------------------

func Test_Benchmark_Throughput_Simple_Add_3_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 200
	nodes, addrs := tests.Setup_n_peers_bc(t, 3, 1, "2s", []float64{float64(iniBalanceA)}, false, true)
	nodeA := nodes[0]
	nodeB := nodes[1]
	nodeC := nodes[2]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 100)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			start := time.Now()

			_, err = nodeA.Calculate("a+b", 5)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond * 10)
			//time.Sleep(time.Second * 3)

			timeTrack(start, "MPC")

			// verify balance
			//start = time.Now()

			block2a := nodeA.BCGetLatestBlock()
			require.NotNil(t, block2a)
			worldstate := block2a.GetWorldStateCopy()
			accountA := permissioned.GetAccountFromWorldState(worldstate, addrs[0])
			require.Equal(t, float64(iniBalanceA-i*3), accountA.GetBalance())
			accountB := permissioned.GetAccountFromWorldState(worldstate, addrs[1])
			require.Equal(t, float64(i*2), accountB.GetBalance())
			accountC := permissioned.GetAccountFromWorldState(worldstate, addrs[2])
			require.Equal(t, float64(i), accountC.GetBalance())
			//timeTrack(start, "'verification'")

			fmt.Println()
		}
		timeTrack(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 120)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

func Test_Benchmark_Throughput_Simple_Add_4_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 100
	nodes, addrs := tests.Setup_n_peers_bc(t, 4, 1, "5s", []float64{float64(iniBalanceA)}, false, true)
	nodeA := nodes[0]
	nodeB := nodes[1]
	nodeC := nodes[2]
	nodeD := nodes[3]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()
	defer nodeD.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 100)

	testNumber := 10
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			start := time.Now()

			_, err = nodeA.Calculate("a+b", 6)
			require.NoError(t, err)

			// try to minimize
			//time.Sleep(time.Millisecond * 1000)
			time.Sleep(time.Second * 2)

			timeTrack(start, "MPC")

			// verify balance
			//start = time.Now()

			block2a := nodeA.BCGetLatestBlock()
			require.NotNil(t, block2a)

			worldstate := block2a.GetWorldStateCopy()
			accountA := permissioned.GetAccountFromWorldState(worldstate, addrs[0])
			require.Equal(t, float64(iniBalanceA-i*4), accountA.GetBalance())
			accountB := permissioned.GetAccountFromWorldState(worldstate, addrs[1])
			require.Equal(t, float64(i*2), accountB.GetBalance())
			accountC := permissioned.GetAccountFromWorldState(worldstate, addrs[2])
			require.Equal(t, float64(i), accountC.GetBalance())
			accountD := permissioned.GetAccountFromWorldState(worldstate, addrs[3])
			require.Equal(t, float64(i), accountD.GetBalance())
			//timeTrack(start, "'verification'")

			fmt.Println()
		}
		timeTrack(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 40)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

// --------------------------- utility ---------------------------

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s \n", name, elapsed)
}

func computeTime(start time.Time, total time.Duration) time.Duration {
	period := time.Since(start)
	return total + period
}
