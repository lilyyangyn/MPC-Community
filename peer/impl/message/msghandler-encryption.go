package message

import (
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

/** Message Handler **/

// ProcessPubkeyMsg is a callback function to handle the received pubkey message
func (m *EncryptionModule) ProcessPubkeyMsg(msg types.Message, pkt transport.Packet) error {
	pubkeyMsg, ok := msg.(*types.PubkeyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	log.Info().Msgf("%s received a pubkey message from %s", m.conf.Socket.GetAddress(), pubkeyMsg.Origin)

	// store pubkey into pubkeyStore
	m.pubkeyStore.add(pubkeyMsg.Origin, &pubkeyMsg.Pubkey)

	return nil
}

// ProcessEntryptedMsg is a callback function to handle the received encrypted message
func (m *EncryptionModule) ProcessEntryptedMsg(msg types.Message, pkt transport.Packet) error {
	encryptedMsg, ok := msg.(*types.EncryptedMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	log.Info().Msgf("%s received a pubkey message from %s", m.conf.Socket.GetAddress(), pkt.Header.Source)

	// decrypt message
	ptxt, err := m.decryptWithPrivkey(*encryptedMsg)
	if err != nil {
		return err
	}

	// process the message locally
	newPkt := transport.Packet{
		Header: pkt.Header,
		Msg:    ptxt,
	}
	err = m.conf.MessageRegistry.ProcessPacket(newPkt)

	return err
}
