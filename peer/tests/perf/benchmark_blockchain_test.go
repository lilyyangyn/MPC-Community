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

func Test_Perf_BC_Troughput(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	getTest := func(n int, maxTxn int, blkTimeout string, repeat int) func(*testing.T) {
		return func(t *testing.T) {
			var sentNum float64 = 0
			var commitNum float64 = 0

			for i := 0; i < repeat; i++ {

				nodes, addrs := tests.Setup_n_peers_bc_helper(t, channel.NewTransport(), n, maxTxn, blkTimeout,
					[]float64{10000}, true, true)
				fmt.Println("----------set up correctly----------")
				nodeA := nodes[0]
				addrA := addrs[0]

				timeout := time.After(time.Second * 1)
				var count float64 = 0
			out:
				for {
					select {
					case <-timeout:
						break out
					default:
						count++
						// fmt.Println(count)
						err := nodeA.SetValueDBAsset("a", 1, float64(count))
						if err != nil {
							break out
						}
						time.Sleep(time.Microsecond * 400)

					}
				}

				block := nodeA.BCGetLatestBlock()
				record := permissioned.GetAssetsFromWorldState(block.States, addrA)
				commit := record.Assets["a"]

				sentNum += count
				commitNum += commit

				fmt.Printf("Round %d: Count= %f , Commit= %f \n", i, count, commit)

				for _, node := range nodes {
					node.Stop()
				}
				time.Sleep(time.Second * 2)
			}

			fmt.Printf("Result: Count=%f, Commit=%f\n", sentNum/float64(repeat), commitNum/float64(repeat))
		}
	}

	t.Run("small group", getTest(3, 10, "2s", 5))
	// t.Run("medium group", getTest(5, 10, "2s", 10))
	// t.Run("large group", getTest(100, 10, "2s", 10))
}

func Test_Perf_BC_Troughput_2(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	getTest := func(n int, maxTxn int, blkTimeout string, repeat int) func(*testing.T) {
		return func(t *testing.T) {
			nodes, addrs := tests.Setup_n_peers_bc_helper(t, channel.NewTransport(), n, maxTxn, blkTimeout,
				[]float64{10000}, true, true)
			fmt.Println("----------set up correctly----------")
			nodeA := nodes[0]
			addrA := addrs[0]

			start := time.Now()

			for j := 0; j < 1000; j++ {
				err := nodeA.SetValueDBAsset("a", 1, float64(j))
				require.NoError(t, err)
				time.Sleep(time.Microsecond * 400)
			}

			elapsed := time.Since(start)

			block := nodeA.BCGetLatestBlock()
			record := permissioned.GetAssetsFromWorldState(block.States, addrA)
			commit := record.Assets["a"]

			fmt.Printf("Benchmark took %s. Sent: 1000, Commit: %f \n", elapsed, commit)
		}
	}

	t.Run("small group", getTest(3, 10, "2s", 5))
	// t.Run("medium group", getTest(5, 10, "2s", 10))
	// t.Run("large group", getTest(100, 10, "2s", 10))
}
