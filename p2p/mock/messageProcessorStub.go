package mock

import (
	"github.com/ElrondNetwork/elrond-go-testing/p2p"
)

type MessageProcessorStub struct {
	ProcessMessageCalled func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
}

func (mps *MessageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	return mps.ProcessMessageCalled(message, broadcastHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mps *MessageProcessorStub) IsInterfaceNil() bool {
	if mps == nil {
		return true
	}
	return false
}
