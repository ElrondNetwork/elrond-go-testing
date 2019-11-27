package broadcast

import (
	"github.com/ElrondNetwork/elrond-go-testing/consensus"
	"github.com/ElrondNetwork/elrond-go-testing/crypto"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
)

func (cm *commonMessenger) SignMessage(message *consensus.Message) ([]byte, error) {
	return cm.signMessage(message)
}

func NewCommonMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) (*commonMessenger, error) {

	return &commonMessenger{
		marshalizer:      marshalizer,
		messenger:        messenger,
		privateKey:       privateKey,
		shardCoordinator: shardCoordinator,
		singleSigner:     singleSigner,
	}, nil
}
