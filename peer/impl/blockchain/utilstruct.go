package blockchain

import (
	"context"
	"crypto/ecdsa"
	"sort"
	"sync"

	permissioned "go.dedis.ch/cs438/permissioned-chain"
	"go.dedis.ch/cs438/storage"
)

// -----------------------------------------------------------------------------
// Next Block Info
type NextBlkInfo struct {
	Height uint
	Miner  bool
}

// -----------------------------------------------------------------------------
// BlkPool

type BlkPool struct {
	*sync.Mutex
	*sync.Cond
	queue []*permissioned.Block
}

func NewBlkPool() *BlkPool {
	lock := sync.Mutex{}
	return &BlkPool{
		Mutex: &lock,
		Cond:  sync.NewCond(&lock),
		queue: make([]*permissioned.Block, 0),
	}
}

func (p *BlkPool) Add(block *permissioned.Block) {
	p.Lock()
	defer p.Unlock()

	p.queue = append(p.queue, block)
	p.Broadcast()
}

func (p *BlkPool) Get(ctx context.Context) *permissioned.Block {
	p.Lock()
	defer p.Unlock()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if len(p.queue) > 0 {
			break
		}
		p.Wait()
	}

	block := p.queue[0]
	p.queue = p.queue[1:]
	return block
}

// -----------------------------------------------------------------------------
// TxnPool

const POOL_CHAN_BUFFER_SIZE = 10

type TxnPool struct {
	*sync.Mutex
	channel       chan *permissioned.SignedTransaction
	queue         []*permissioned.SignedTransaction
	newTxnChannel chan struct{}
}

func NewTxnPool() *TxnPool {
	lock := sync.Mutex{}
	return &TxnPool{
		Mutex: &lock,
		channel: make(chan *permissioned.SignedTransaction,
			POOL_CHAN_BUFFER_SIZE),
		queue:         make([]*permissioned.SignedTransaction, 0),
		newTxnChannel: make(chan struct{}, 1),
	}
}

func (p *TxnPool) Daemon(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.newTxnChannel:
			p.Lock()
			for _, txn := range p.queue {
				p.Unlock()
				p.channel <- txn
				p.Lock()
			}
			p.queue = make([]*permissioned.SignedTransaction, 0)
			p.Unlock()
		}
	}
}

func (p *TxnPool) sortedInsert(txn *permissioned.SignedTransaction) {
	i := sort.Search(len(p.queue), func(i int) bool {
		return p.queue[i].Txn.Nonce > txn.Txn.Nonce
	})
	if i == len(p.queue) {
		// Insert at end is the easy case.
		p.queue = append(p.queue, txn)
	} else {
		// Make space for the inserted element by shifting values
		p.queue = append(p.queue[:i+1], p.queue[i:]...)
		// Insert the new element.
		p.queue[i] = txn
	}
}

func (p *TxnPool) Push(txn *permissioned.SignedTransaction) {
	p.Lock()
	defer p.Unlock()

	p.sortedInsert(txn)

	if len(p.newTxnChannel) == 0 {
		p.newTxnChannel <- struct{}{}
	}
}

func (p *TxnPool) PushBackSeveral(txns []permissioned.SignedTransaction) {
	p.Lock()
	defer p.Unlock()

	for _, txn := range txns {
		p.sortedInsert(&txn)
	}

	if len(p.newTxnChannel) == 0 {
		p.newTxnChannel <- struct{}{}
	}
}

// func (p *TxnPool) Pull() <-chan *permissioned.SignedTransaction {
// 	p.Lock()
// 	defer p.Unlock()

// 	if len(p.channel) > 0 {
// 		return p.channel
// 	}

// 	i := 0
// 	queueLen := len(p.queue)
// 	for ; i < POOL_CHAN_BUFFER_SIZE && i < queueLen; i++ {
// 		p.channel <- p.queue[i]
// 	}
// 	p.queue = p.queue[i:]

// 	return p.channel
// }

// -----------------------------------------------------------------------------
// SyncCenter

type SyncCenter struct {
	*sync.Mutex
	store map[string]chan error
}

func NewSyncCenter() *SyncCenter {
	return &SyncCenter{
		Mutex: &sync.Mutex{},
		store: map[string]chan error{},
	}
}

func (c *SyncCenter) Register(id string, channel chan error) {
	c.Lock()
	defer c.Unlock()

	c.store[id] = channel
}

func (c *SyncCenter) Notify(id string, err error) {
	c.Lock()
	defer c.Unlock()

	channel, ok := c.store[id]
	if !ok {
		return
	}

	channel <- err
	delete(c.store, id)
}

// -----------------------------------------------------------------------------
// WatchRegistry

type watchCallbck func(config *permissioned.ChainConfig,
	txn *permissioned.Transaction) error

type WatchRegistry struct {
	*sync.RWMutex
	store map[permissioned.TxnType]watchCallbck
}

func NewWatchRegistry() *WatchRegistry {
	r := WatchRegistry{
		RWMutex: &sync.RWMutex{},
		store:   map[permissioned.TxnType]watchCallbck{},
	}
	return &r
}

func (r *WatchRegistry) Register(txnType permissioned.TxnType,
	watcher watchCallbck) {
	r.Lock()
	defer r.Unlock()

	r.store[txnType] = watcher
}

func (r *WatchRegistry) Tell(config *permissioned.ChainConfig, txn *permissioned.Transaction) error {
	r.RLock()
	defer r.RUnlock()

	watcher, ok := r.store[txn.Type]
	if !ok {
		return nil
	}

	return watcher(config, txn)
}

// -----------------------------------------------------------------------------
// Wallet

type Wallet struct {
	*sync.RWMutex
	account *permissioned.Account
	privKey *ecdsa.PrivateKey
	addr    *permissioned.Address
}

func NewWallet(privkey *ecdsa.PrivateKey) *Wallet {
	address := permissioned.NewAddress(&privkey.PublicKey)
	account := permissioned.NewAccount(*address)
	r := Wallet{
		RWMutex: &sync.RWMutex{},
		privKey: privkey,
		account: account,
		addr:    address,
	}
	return &r
}

func (w *Wallet) GetAddress() permissioned.Address {
	// Assume addr never change
	return *w.addr
}

func (w *Wallet) Sync(worldState storage.KVStore) {
	// w.Lock()
	// defer w.Unlock()

	// account := permissioned.GetAccountFromWorldState(worldState, w.addr.Hex)

	// if w.account.GetAddress().Hex != account.GetAddress().Hex {
	// 	return
	// }

	// if w.account.GetNonce() != account.GetNonce() {
	// 	w.account = account
	// }
}

func (w *Wallet) PreMPCTxn(expression string, budget float64, prime string) (*permissioned.SignedTransaction, error) {
	w.Lock()
	defer w.Unlock()

	propose := permissioned.MPCPropose{
		Initiator:  w.account.GetAddress().Hex,
		Budget:     budget,
		Expression: expression,
		Prime:      prime,
	}
	txn := permissioned.NewTransactionPreMPC(w.account, propose)
	signedTxn, err := txn.Sign(w.privKey)
	if err != nil {
		return nil, err
	}
	w.account.IncreaseNonce()

	return signedTxn, err
}

func (w *Wallet) PostMPCTxn(id string, result float64) (*permissioned.SignedTransaction, error) {
	w.Lock()
	defer w.Unlock()

	ResultHash := storage.Hash(result)

	record := permissioned.MPCRecord{
		UniqID: id,
		Result: ResultHash,
	}
	txn := permissioned.NewTransactionPostMPC(w.account, record)
	signedTxn, err := txn.Sign(w.privKey)
	if err != nil {
		return nil, err
	}
	w.account.IncreaseNonce()

	return signedTxn, err
}

func (w *Wallet) RegAssets(assets map[string]float64) (*permissioned.SignedTransaction, error) {
	w.Lock()
	defer w.Unlock()

	txn := permissioned.NewTransactionRegAssets(w.account, assets)
	signedTxn, err := txn.Sign(w.privKey)
	if err != nil {
		return nil, err
	}
	w.account.IncreaseNonce()

	return signedTxn, err
}

func (w *Wallet) RegEnckeyTxn(pubkey string) (*permissioned.SignedTransaction, error) {
	w.Lock()
	defer w.Unlock()

	txn := permissioned.NewTransactionRegEnckey(w.account, pubkey)
	signedTxn, err := txn.Sign(w.privKey)
	if err != nil {
		return nil, err
	}
	w.account.IncreaseNonce()

	return signedTxn, err
}
