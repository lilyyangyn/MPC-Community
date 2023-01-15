package perf

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"go.dedis.ch/cs438/peer/tests"
	"go.dedis.ch/cs438/transport/channel"
)

// --------------------------------- benchmark --------------------------------------
// WITHOUT balance correctness check!

// ##################################################
// test the impact of total node number on throughput
// ##################################################

func setup(t *testing.T, n int, maxTxn int,
	timeout string, gains []float64, initSleepTime time.Duration) ([]*z.TestNode, []string) {
	return tests.Setup_n_peers_bc_perf(t, channel.NewTransport(), n, maxTxn, timeout, []float64{10000}, initSleepTime, false, true)
	// return tests.Setup_n_peers_bc(t, n, maxTxn, timeout, gains, false, true)
}

// addition tests
func Test_Throughput_Simple_Add_3_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 3, 1, "2s", []float64{float64(iniBalanceA)}, time.Millisecond*500)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 0.1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 0.1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 1000)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			//start := time.Now()

			_, err = nodeA.Calculate("a+b", 3.2)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond * 8)
			//time.Sleep(time.Second * 3)

			//timer(start, "MPC")
			//fmt.Println()
		}
		timer(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 30)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

func Test_Throughput_Simple_Add_4_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 4, 1, "2s", []float64{float64(iniBalanceA)}, time.Second)
	nodeA, nodeB, nodeC, nodeD := nodes[0], nodes[1], nodes[2], nodes[3]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()
	defer nodeD.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 0.1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 0.1)
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

			//start := time.Now()

			_, err = nodeA.Calculate("a+b", 4.2)
			require.NoError(t, err)

			//timer(start, "MPC")

			// try to minimize
			time.Sleep(time.Millisecond * 30)
			//time.Sleep(time.Second * 3)
		}
		timer(overall, "overall execution")

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 30)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

func Test_Throughput_Simple_Add_5_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	// zerolog.SetGlobalLevel(zerolog.InfoLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 5, 1, "10s", []float64{float64(iniBalanceA)}, time.Second*3)
	nodeA, nodeB, nodeC, nodeD, nodeE := nodes[0], nodes[1], nodes[2], nodes[3], nodes[4]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()
	defer nodeD.Stop()
	defer nodeE.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 1000)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			//start := time.Now()

			_, err = nodeA.Calculate("a+b", 7)
			require.NoError(t, err)

			//timer(start, "MPC")

			// try to minimize
			//time.Sleep(time.Millisecond * 500)
			time.Sleep(time.Millisecond * 200)
		}
		timer(overall, "overall execution")

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 60)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

// multiplication tests
func Test_Throughput_Simple_Mul_3_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 3, 1, "2s", []float64{float64(iniBalanceA)}, time.Millisecond*200)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 1000)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			start := time.Now()

			_, err = nodeA.Calculate("a*b", 5)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond * 15)
			//time.Sleep(time.Second * 3)

			timer(start, "MPC")
			fmt.Println()
		}
		timer(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 30)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

func Test_Throughput_Simple_Mul_4_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 4, 1, "2s", []float64{float64(iniBalanceA)}, time.Second)
	nodeA, nodeB, nodeC, nodeD := nodes[0], nodes[1], nodes[2], nodes[3]
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

	time.Sleep(time.Millisecond * 1000)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			start := time.Now()

			_, err = nodeA.Calculate("a*b", 6)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond * 45)
			//time.Sleep(time.Second * 3)

			timer(start, "MPC")
			//fmt.Println()
		}
		timer(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 30)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

func Test_Throughput_Simple_Mul_5_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 5, 1, "2s", []float64{float64(iniBalanceA)}, time.Second*3)
	nodeA, nodeB, nodeC, nodeD, nodeE := nodes[0], nodes[1], nodes[2], nodes[3], nodes[4]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()
	defer nodeD.Stop()
	defer nodeE.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 1000)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			start := time.Now()

			_, err = nodeA.Calculate("a*b", 7)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond * 250)
			//time.Sleep(time.Second * 3)

			timer(start, "MPC")
			//fmt.Println()
		}
		timer(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 60)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

// complex computation
func Test_Throughput_Complex_3_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 3, 1, "2s", []float64{float64(iniBalanceA)}, time.Millisecond*200)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 1)
	require.NoError(t, err)
	err = nodeC.SetValueDBAsset("c", 1, 1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 1000)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			start := time.Now()

			_, err = nodeA.Calculate("(a+b+c)*(a+b)*c*a*b*(b-c)", 6)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond * 50)
			//time.Sleep(time.Second * 1)

			timer(start, "MPC")
			fmt.Println()
		}
		timer(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 60)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

func Test_Throughput_Complex_4_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 500
	nodes, _ := setup(t, 4, 1, "5s", []float64{float64(iniBalanceA)}, time.Second)
	nodeA, nodeB, nodeC, nodeD := nodes[0], nodes[1], nodes[2], nodes[3]
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()
	defer nodeD.Stop()

	//start := time.Now()

	err := nodeA.SetValueDBAsset("a", 1, 1)
	require.NoError(t, err)
	err = nodeB.SetValueDBAsset("b", 1, 1)
	require.NoError(t, err)
	err = nodeC.SetValueDBAsset("c", 1, 1)
	require.NoError(t, err)

	//timeTrack(start, "'set assets value'")

	time.Sleep(time.Millisecond * 1000)

	testNumber := 50
	mpcDone := make(chan struct{})
	go func() {
		// stress test: compute MPC continuously for 50 times
		overall := time.Now()
		for i := 1; i <= testNumber; i++ {
			fmt.Printf("the %v iteration\n", i)

			start := time.Now()

			_, err = nodeA.Calculate("(a+b+c)*(a+b)*c*a*b*(b-c)", 7)
			require.NoError(t, err)

			// try to minimize
			//time.Sleep(time.Millisecond * 600)
			time.Sleep(time.Second * 5)

			timer(start, "MPC")
			fmt.Println()
		}
		timer(overall, "overall execution")
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

// #######################################
// test the impact of macTxn on throughput
// #######################################
func Test_Throughput_maxTxn_3_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 3, 7, "2s", []float64{float64(iniBalanceA)}, time.Second*3)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]
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

			//start := time.Now()

			_, err = nodeA.Calculate("a+b", 5)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond)
			//time.Sleep(time.Second * 3)

			//timer(start, "MPC")
			//fmt.Println()
		}
		timer(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 200)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

// #######################################
// test the impact of txn timeout on throughput
// #######################################
func Test_Throughput_timeout_3_nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)

	const iniBalanceA = 300
	nodes, _ := setup(t, 3, 5, "5s", []float64{float64(iniBalanceA)}, time.Millisecond*200)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]
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

			//start := time.Now()

			_, err = nodeA.Calculate("a+b", 5)
			require.NoError(t, err)

			// try to minimize
			time.Sleep(time.Millisecond * 10)
			//time.Sleep(time.Second * 3)

			//timer(start, "MPC")
			//fmt.Println()
		}
		timer(overall, "overall execution")
		//fmt.Printf("overall execution time: %s", overall)

		close(mpcDone)
	}()

	timeout := time.After(time.Second * 500)

	select {
	case <-mpcDone:
	case <-timeout:
		t.Error(t, "timeout error: cannot finish test in given time")
	}
}

// --------------------------- utility ---------------------------

func timer(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s \n", name, elapsed)
}
