package permissioned

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"go.dedis.ch/cs438/storage"
)

// -----------------------------------------------------------------------------
// Transaction

type TxnType string

const (
	TxnTypeConfig  TxnType = "txn-config"
	TxnTypePreMPC  TxnType = "txn-prempc"
	TxnTypePostMPC TxnType = "txn-postmpc"
)

// Transaction represents the transaction that happens inside this chain
type Transaction struct {
	ID    string
	Nonce uint
	From  string
	To    string
	Type  TxnType
	Value float64
	Data  interface{}
}

// NewTransaction creates a new transaction and computes its ID
func NewTransaction(from *Account, to *Address, txntype TxnType,
	value float64, data interface{}) *Transaction {
	txn := Transaction{
		Nonce: from.nonce,
		From:  from.addr.Hex,
		To:    to.Hex,
		Type:  txntype,
		Value: value,
		Data:  data,
	}
	txn.ID = txn.Hash()

	return &txn
}

// HashBytes computes the hash of the transaction
func (txn *Transaction) HashBytes() []byte {
	h := sha256.New()

	h.Write([]byte(fmt.Sprintf("%d", txn.Nonce)))
	h.Write([]byte(txn.From))
	h.Write([]byte(txn.To))
	h.Write([]byte(txn.Type))
	h.Write([]byte(fmt.Sprintf("%f", txn.Value)))

	bytes, err := json.Marshal(txn.Data)
	if err != nil {
		panic(err)
	}
	h.Write(bytes)

	return h.Sum(nil)
}

// Hash computes the hex-encoded hash of the transaction
func (txn *Transaction) Hash() string {
	return hex.EncodeToString(txn.HashBytes())
}

// Exec executes the transaction based on the input worldState
func (txn *Transaction) Exec(worldState storage.KVStore) error {
	// check nonce
	account := getAccountFromWorldState(worldState, txn.From)
	if account.nonce != txn.Nonce {
		return fmt.Errorf("transaction %s has invalid nonce from %s. Expected: %d, Got: %d",
			txn.ID, txn.From, account.nonce, txn.Nonce)
	}

	// execute handler
	handler, ok := txnHandlerStore[txn.Type]
	if !ok {
		return fmt.Errorf("invalid transaction type: %s", txn.Type)
	}
	err := handler(worldState, txn)
	if err != nil {
		return err
	}

	// advance nonce to avoid replay attack
	return updateNonce(worldState, txn.From)
}

// -----------------------------------------------------------------------------
// Signed Transaction

type SignedTransaction struct {
	Txn       Transaction
	Signature []byte
}

// Sign creates a signature for the trasaction using the given private key
func (txn *Transaction) Sign(privateKey *ecdsa.PrivateKey) (*SignedTransaction, error) {
	signature, err := crypto.Sign(txn.HashBytes(), privateKey)
	if err != nil {
		return nil, err
	}

	return &SignedTransaction{Txn: *txn, Signature: signature}, nil
}

// Hash computes the hash of the signed transaction
func (signedTxn *SignedTransaction) Hash() []byte {
	h := sha256.New()

	h.Write([]byte(fmt.Sprintf("%d", signedTxn.Txn.Nonce)))
	h.Write([]byte(signedTxn.Txn.From))
	h.Write([]byte(signedTxn.Txn.To))
	h.Write([]byte(signedTxn.Txn.Type))
	h.Write([]byte(fmt.Sprintf("%f", signedTxn.Txn.Value)))

	bytes, err := json.Marshal(signedTxn.Txn.Data)
	if err != nil {
		panic(err)
	}
	h.Write(bytes)

	h.Write(signedTxn.Signature)

	return h.Sum(nil)
}

// Verify verify the signature and then execute to see whether the result is consistent with worldState
func (signedTxn *SignedTransaction) Verify(worldState storage.KVStore) error {
	// verify signature
	txn := signedTxn.Txn
	digestHash := txn.HashBytes()
	publicKey, err := crypto.Ecrecover(digestHash, signedTxn.Signature)
	if err != nil {
		return err
	}
	addr := NewAddress(publicKey)
	if addr.Hex != txn.From {
		return fmt.Errorf("transaction %s is not signed by sender %s", signedTxn.Txn.ID, signedTxn.Txn.From)
	}
	// verify sig input needs to be in [R || S] format
	sigValid := crypto.VerifySignature(publicKey, digestHash, signedTxn.Signature[:len(signedTxn.Signature)-1])
	if !sigValid {
		return fmt.Errorf("transaction %s has invalid signature from %s", signedTxn.Txn.ID, signedTxn.Txn.From)
	}

	// execute txn
	err = txn.Exec(worldState)

	return err
}

// -----------------------------------------------------------------------------
// Transaction Polymophism - PreMPC

func NewTransactionPreMPC(initiator *Account, data MPCRecord) *Transaction {
	return NewTransaction(
		initiator,
		&ZeroAddress,
		TxnTypePreMPC,
		data.Budget,
		data,
	)
}

func execPreMPC(worldState storage.KVStore, txn *Transaction) error {
	record := txn.Data.(MPCRecord)
	if record.Budget != txn.Value || record.Initiator != txn.From {
		return fmt.Errorf("Transaction data inconsistent")
	}

	// lock balance to avoid double spending
	config := getConfigFromWorldState(worldState)
	err := lockBalance(worldState, txn.From, record.Budget*float64(len(config.Participants)))
	if err != nil {
		return err
	}
	// add MPC record to worldState
	err = worldState.Put(mpcKeyFromUniqID(record.UniqID), MPCEndorsement{
		Peers:     config.Participants,
		Endorsers: map[string]struct{}{},
		Budget:    record.Budget,
		Locked:    true,
	})
	if err != nil {
		panic(err)
	}

	return nil
}

// -----------------------------------------------------------------------------
// Transaction Polymophism - PostMPC

func NewTransactionPostMPC(from *Account, data MPCRecord) *Transaction {
	return NewTransaction(
		from,
		&ZeroAddress,
		TxnTypePostMPC,
		0,
		data,
	)
}

func execPostMPC(worldState storage.KVStore, txn *Transaction) error {
	record := txn.Data.(MPCRecord)
	initiator := getAccountFromWorldState(worldState, record.Initiator)

	// update endorsement information, collect awawrd if threshold is reached
	err := updateMPCEndorsement(worldState, mpcKeyFromUniqID(record.UniqID), initiator, txn.From)
	if err != nil {
		return err
	}

	return nil
}

// -----------------------------------------------------------------------------
// Utilities

var AwardUnlockThreshold = 0.5

type MPCRecord struct {
	UniqID     string
	Initiator  string
	Budget     float64
	Expression string
	Result     float64
}

type MPCEndorsement struct {
	Peers     map[string]struct{}
	Endorsers map[string]struct{}
	Budget    float64
	Locked    bool
}

var txnHandlerStore = map[TxnType]func(storage.KVStore, *Transaction) error{
	TxnTypePreMPC:  execPreMPC,
	TxnTypePostMPC: execPostMPC,
}

func mpcKeyFromUniqID(uniqID string) string {
	return fmt.Sprintf("ongoging-mpc-%s", uniqID)
}

func getAccountFromWorldState(worldState storage.KVStore, key string) *Account {
	object, ok := worldState.Get(key)
	if !ok {
		return NewAccount(Address{Hex: key})
	}
	account := object.(Account)
	return &account
}

func getConfigFromWorldState(worldState storage.KVStore) *ChainConfig {
	object, ok := worldState.Get(STATE_CONFIG_KEY)
	if !ok {
		panic(fmt.Errorf("config not exists"))
	}
	config := object.(ChainConfig)
	return &config
}

func getMPCEndorsementFromWorldState(worldState storage.KVStore, key string) *MPCEndorsement {
	object, ok := worldState.Get(key)
	if !ok {
		panic(fmt.Errorf("MPC endorsement not exists"))
	}
	endorsement := object.(MPCEndorsement)
	return &endorsement
}

func updateNonce(worldState storage.KVStore, accountID string) error {
	account := getAccountFromWorldState(worldState, accountID)
	account.nonce++

	err := worldState.Put(accountID, *account)
	if err != nil {
		panic(err)
	}
	return nil
}

func lockBalance(worldState storage.KVStore, accountID string, amount float64) error {
	account := getAccountFromWorldState(worldState, accountID)
	if account.balance < amount {
		return fmt.Errorf("Initiator(%s) balance not enough", accountID)
	}

	// lock balance
	account.balance -= amount
	account.lockedBalance += amount
	err := worldState.Put(accountID, *account)
	if err != nil {
		panic(err)
	}
	return nil
}

func getAward(worldState storage.KVStore, from *Account, to string, amount float64) {
	account := from
	if from.addr.Hex != to {
		account = getAccountFromWorldState(worldState, to)
	}

	if from.lockedBalance < amount {
		panic(fmt.Errorf("%s's locked balance not enough", from.addr.Hex))
	}
	from.lockedBalance -= amount
	account.balance += amount
	err := worldState.Put(from.addr.Hex, *from)
	if err != nil {
		panic(err)
	}
	err = worldState.Put(account.addr.Hex, *account)
	if err != nil {
		panic(err)
	}
}

func updateMPCEndorsement(worldState storage.KVStore, key string, initiator *Account, accountID string) error {
	endorsement := getMPCEndorsementFromWorldState(worldState, key)
	if _, ok := endorsement.Peers[accountID]; !ok {
		return fmt.Errorf("no-participant sending endorsement. Potentially an attack")
	}

	endorsement.Endorsers[accountID] = struct{}{}
	if !endorsement.Locked {
		if len(endorsement.Endorsers) == len(endorsement.Peers) {
			err := worldState.Del(key)
			if err != nil {
				panic(err)
			}
		}
		getAward(worldState, initiator, accountID, endorsement.Budget)
		return nil
	}

	threshold := float64(len(endorsement.Peers)) * AwardUnlockThreshold
	if float64(len(endorsement.Endorsers)) > threshold {
		for endorser, _ := range endorsement.Endorsers {
			getAward(worldState, initiator, endorser, endorsement.Budget)
		}
		endorsement.Locked = false
		if len(endorsement.Endorsers) == len(endorsement.Peers) {
			err := worldState.Del(key)
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
}
