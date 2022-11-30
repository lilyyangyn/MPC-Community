package paxos

import (
	"crypto"
	"encoding/hex"
	"strconv"

	"go.dedis.ch/cs438/storage"
	"go.dedis.ch/cs438/types"
)

const (
	Idle = iota
	InConsensus
	ReadyToSwitch
)

type MultiPaxosSate int

type MultiPaxos struct {
	TLC uint
	*Paxos

	state           MultiPaxosSate
	blockchainStore storage.Store

	blockCounter int
	futureBlocks map[uint][]*types.BlockchainBlock
}

func NewMultiPaxos(blockchainStore storage.Store) *MultiPaxos {
	multipaxos := MultiPaxos{
		TLC:   0,
		Paxos: NewPaxos(),

		state:           Idle,
		blockchainStore: blockchainStore,

		futureBlocks: map[uint][]*types.BlockchainBlock{},
	}

	return &multipaxos
}

func (multipaxos *MultiPaxos) InitNewPaxos(value *types.PaxosValue, id uint) (uint, bool) {
	if multipaxos.state != Idle {
		return multipaxos.TLC, false
	}

	multipaxos.Paxos.setFirstValue(value)
	success := multipaxos.Paxos.joinPhaseOne(id)
	if success {
		multipaxos.state = InConsensus
	}

	return multipaxos.TLC, success
}

func (multipaxos *MultiPaxos) RetryPaxos(increase uint) (uint, uint, bool) {
	success := multipaxos.Paxos.joinPhaseOne(multipaxos.Paxos.proposeID + increase)
	return multipaxos.TLC, multipaxos.Paxos.proposeID, success
}

func (multipaxos *MultiPaxos) RecordID(step uint, id uint) bool {
	if step != multipaxos.TLC {
		return false
	}
	if multipaxos.Paxos.recordID(id) {
		return true
	}
	return false
}

func (multipaxos *MultiPaxos) RecordPromise(step uint, id uint, acceptedID uint,
	acceptedValue *types.PaxosValue, threshold int) (*types.PaxosValue, bool) {
	if step != multipaxos.TLC {
		return nil, false
	}
	if multipaxos.state == InConsensus && multipaxos.Paxos.recordPromise(id, acceptedID, acceptedValue, threshold) {
		return multipaxos.Paxos.proposeVal, true
	}

	return nil, false
}

func (multipaxos *MultiPaxos) RecordAccept(step uint, id uint, value *types.PaxosValue,
	threshold int) (*types.BlockchainBlock, bool) {
	if step != multipaxos.TLC {
		return nil, false
	}

	if multipaxos.state != ReadyToSwitch && multipaxos.Paxos.RecordAccept(id, value, threshold) {
		return multipaxos.createBlock(value), true
	}

	return nil, false
}

func (multipaxos *MultiPaxos) Accept(step uint, id uint, value *types.PaxosValue) bool {
	if step != multipaxos.TLC {
		return false
	}

	if multipaxos.state != Complete {
		return multipaxos.Paxos.accept(id, value)
	}
	return false
}

func (multipaxos *MultiPaxos) RecordBlock(step uint, block *types.BlockchainBlock,
	threshold int) (bool, bool) {
	if step < multipaxos.TLC {
		return false, false
	}

	if multipaxos.state != ReadyToSwitch && step > multipaxos.TLC {
		if _, ok := multipaxos.futureBlocks[step]; !ok {
			multipaxos.futureBlocks[step] = []*types.BlockchainBlock{block}
		} else {
			multipaxos.futureBlocks[step] = append(multipaxos.futureBlocks[step], block)
		}
		return false, false
	}

	return multipaxos.recordBlock(block, threshold), multipaxos.isCatchUp()
}

func (multipaxos *MultiPaxos) AppendBlock(block *types.BlockchainBlock) error {
	blockKey := hex.EncodeToString(block.Hash)

	// add to store
	buf, err := block.Marshal()
	if err != nil {
		return err
	}
	multipaxos.blockchainStore.Set(blockKey, buf)

	// update last block
	multipaxos.blockchainStore.Set(storage.LastBlockKey, block.Hash)

	return nil
}

func (multipaxos *MultiPaxos) AdvanceClock(block *types.BlockchainBlock) ([]*types.BlockchainBlock, bool) {
	if multipaxos.state != ReadyToSwitch {
		return nil, false
	}
	if block.Index != multipaxos.TLC {
		return nil, false
	}

	multipaxos.TLC++
	multipaxos.Paxos = NewPaxos()
	multipaxos.blockCounter = 0
	multipaxos.state = Idle

	nextStepBlocks := multipaxos.futureBlocks[multipaxos.TLC]
	delete(multipaxos.futureBlocks, multipaxos.TLC)

	return nextStepBlocks, true
}

/** Private Helpfer Functions **/

func (multipaxos *MultiPaxos) isCatchUp() bool {
	counter := 0
	for _, blocks := range multipaxos.futureBlocks {
		counter += len(blocks)
	}

	return counter > 0
}

func (multipaxos *MultiPaxos) recordBlock(block *types.BlockchainBlock, threshold int) bool {
	if block.Index != multipaxos.TLC {
		return false
	}

	multipaxos.blockCounter++
	if multipaxos.blockCounter >= threshold {
		multipaxos.state = ReadyToSwitch
		return true
	}

	return false
}

func (multipaxos *MultiPaxos) createBlock(val *types.PaxosValue) *types.BlockchainBlock {
	prevHash := multipaxos.blockchainStore.Get(storage.LastBlockKey)
	if len(prevHash) == 0 {
		prevHash = make([]byte, 32)
	}
	// compute block hash
	currClock := multipaxos.TLC
	// create block
	block := &types.BlockchainBlock{
		Index:    currClock,
		Value:    *val,
		PrevHash: prevHash,
	}

	h := crypto.SHA256.New()
	h.Write([]byte(strconv.Itoa(int(block.Index))))
	h.Write([]byte(block.Value.UniqID))
	h.Write([]byte(block.Value.Filename))
	h.Write([]byte(block.Value.Metahash))
	h.Write(block.PrevHash)
	blockHash := h.Sum(nil)
	block.Hash = blockHash

	return block
}
