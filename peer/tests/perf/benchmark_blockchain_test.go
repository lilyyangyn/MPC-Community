package perf

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/tests"
	"go.dedis.ch/cs438/permissioned-chain"
	"go.dedis.ch/cs438/transport/channel"
)

var getTest = func(n int, maxTxn int, blkTimeout string, initSleepTime, sleepTime time.Duration) func(*testing.T) {
	return func(t *testing.T) {
		nodes, addrs := tests.Setup_n_peers_bc_perf(t, channel.NewTransport(), n, maxTxn, blkTimeout,
			[]float64{10000}, initSleepTime, true, true)
		fmt.Println("----------set up correctly----------")
		nodeA := nodes[0]
		addrA := addrs[0]

		total := 1000

		start := time.Now()

		for j := 0; j < total; j++ {
			err := nodeA.SetValueDBAsset("a", 1, float64(j+1))
			require.NoError(t, err)

			if j%50 == 0 {
				fmt.Println("Round", j)
			}
			time.Sleep(sleepTime)
		}

		for {
			block := nodeA.BCGetLatestBlock()
			record := permissioned.GetAssetsFromWorldState(block.States, addrA)
			commit := record.Assets["a"]

			if commit == float64(total) {
				break
			}
		}
		elapsed := time.Since(start)

		fmt.Printf("Benchmark took %s. Node: %d, Sent: %d \n", elapsed, n, total)
	}
}

func Test_Perf_BC_Troughput_2_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("2-node", getTest(2, 10, "2s", time.Millisecond*500, time.Microsecond*100))
}

func Test_Perf_BC_Troughput_3_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("3-node", getTest(3, 10, "2s", time.Millisecond*500, time.Microsecond*450))
}

func Test_Perf_BC_Troughput_4_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("4-node", getTest(4, 10, "2s", time.Millisecond*800, time.Microsecond*650))
}

func Test_Perf_BC_Troughput_5_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("5-node", getTest(5, 10, "2s", time.Second*3, time.Microsecond*800))
}

func Test_Perf_BC_Troughput_8_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("5-node", getTest(8, 10, "2s", time.Second*7, time.Millisecond*3))
}

func Test_Perf_BC_Troughput_10_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("10-node", getTest(10, 10, "2s", time.Second*10, time.Millisecond*7))
}

func Test_Perf_BC_Troughput_16_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("10-node", getTest(16, 10, "2s", time.Second*10, time.Microsecond*20500))
}

func Test_Perf_BC_Troughput_20_Nodes(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	t.Run("10-node", getTest(20, 10, "2s", time.Second*10, time.Millisecond*30))
}
