package types

import (
	permissioned "go.dedis.ch/cs438/permissioned-chain"
	"go.dedis.ch/cs438/transport"
)

type BCPrivateMessage struct {
	// Recipients is a bag of recipients
	Recipients map[string]struct{}

	// Msg is the private message to be read by the recipients
	Msg *transport.Message
}

type BCTxnMessag struct {
	Origin string
	Txn    permissioned.SignedTransaction
}

type BCBlkMessage struct {
	Origin string
	Blk    permissioned.Block
}
