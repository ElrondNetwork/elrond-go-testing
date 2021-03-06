package metachain

import (
	"github.com/ElrondNetwork/elrond-go-testing/core/check"
	"github.com/ElrondNetwork/elrond-go-testing/data/block"
	"github.com/ElrondNetwork/elrond-go-testing/data/state"
	"github.com/ElrondNetwork/elrond-go-testing/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-testing/hashing"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/process"
	"github.com/ElrondNetwork/elrond-go-testing/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go-testing/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
)

type preProcessorsContainerFactory struct {
	shardCoordinator    sharding.Coordinator
	store               dataRetriever.StorageService
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	dataPool            dataRetriever.MetaPoolsHolder
	txProcessor         process.TransactionProcessor
	accounts            state.AccountsAdapter
	requestHandler      process.RequestHandler
	economicsFee        process.FeeHandler
	miniBlocksCompacter process.MiniBlocksCompacter
	gasHandler          process.GasHandler
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	dataPool dataRetriever.MetaPoolsHolder,
	accounts state.AccountsAdapter,
	requestHandler process.RequestHandler,
	txProcessor process.TransactionProcessor,
	economicsFee process.FeeHandler,
	miniBlocksCompacter process.MiniBlocksCompacter,
	gasHandler process.GasHandler,
) (*preProcessorsContainerFactory, error) {

	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStore
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(dataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(requestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(miniBlocksCompacter) {
		return nil, process.ErrNilMiniBlocksCompacter
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:    shardCoordinator,
		store:               store,
		marshalizer:         marshalizer,
		hasher:              hasher,
		dataPool:            dataPool,
		txProcessor:         txProcessor,
		accounts:            accounts,
		requestHandler:      requestHandler,
		economicsFee:        economicsFee,
		miniBlocksCompacter: miniBlocksCompacter,
		gasHandler:          gasHandler,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcm *preProcessorsContainerFactory) Create() (process.PreProcessorsContainer, error) {
	container := containers.NewPreProcessorsContainer()

	preproc, err := ppcm.createTxPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.TxBlock, preproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcm *preProcessorsContainerFactory) createTxPreProcessor() (process.PreProcessor, error) {
	txPreprocessor, err := preprocess.NewTransactionPreprocessor(
		ppcm.dataPool.Transactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.txProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestTransaction,
		ppcm.economicsFee,
		ppcm.miniBlocksCompacter,
		ppcm.gasHandler,
	)

	return txPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	if ppcm == nil {
		return true
	}
	return false
}
