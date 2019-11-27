package sposFactory

import (
	"github.com/ElrondNetwork/elrond-go-testing/consensus"
	"github.com/ElrondNetwork/elrond-go-testing/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go-testing/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-testing/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go-testing/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-testing/core"
	"github.com/ElrondNetwork/elrond-go-testing/core/indexer"
	"github.com/ElrondNetwork/elrond-go-testing/crypto"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
)

// GetSubroundsFactory returns a subrounds factory depending of the given parameter
func GetSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	consensusType string,
	appStatusHandler core.AppStatusHandler,
	indexer indexer.Indexer,
) (spos.SubroundsFactory, error) {

	switch consensusType {
	case blsConsensusType:
		subRoundFactoryBls, err := bls.NewSubroundsFactory(consensusDataContainer, consensusState, worker)
		if err != nil {
			return nil, err
		}

		err = subRoundFactoryBls.SetAppStatusHandler(appStatusHandler)
		if err != nil {
			return nil, err
		}

		subRoundFactoryBls.SetIndexer(indexer)

		return subRoundFactoryBls, nil
	case bnConsensusType:
		subRoundFactoryBn, err := bn.NewSubroundsFactory(consensusDataContainer, consensusState, worker)
		if err != nil {
			return nil, err
		}

		err = subRoundFactoryBn.SetAppStatusHandler(appStatusHandler)
		if err != nil {
			return nil, err
		}

		subRoundFactoryBn.SetIndexer(indexer)

		return subRoundFactoryBn, nil
	}

	return nil, ErrInvalidConsensusType
}

// GetConsensusCoreFactory returns a consensus service depending of the given parameter
func GetConsensusCoreFactory(consensusType string) (spos.ConsensusService, error) {
	switch consensusType {
	case blsConsensusType:
		return bls.NewConsensusService()
	case bnConsensusType:
		return bn.NewConsensusService()
	}

	return nil, ErrInvalidConsensusType
}

// GetBroadcastMessenger returns a consensus service depending of the given parameter
func GetBroadcastMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	privateKey crypto.PrivateKey,
	singleSigner crypto.SingleSigner,
) (consensus.BroadcastMessenger, error) {

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return broadcast.NewShardChainMessenger(marshalizer, messenger, privateKey, shardCoordinator, singleSigner)
	}

	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return broadcast.NewMetaChainMessenger(marshalizer, messenger, privateKey, shardCoordinator, singleSigner)
	}

	return nil, ErrInvalidShardId
}
