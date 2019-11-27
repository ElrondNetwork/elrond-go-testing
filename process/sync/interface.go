package sync

import (
	"github.com/ElrondNetwork/elrond-go-testing/data"
	"github.com/ElrondNetwork/elrond-go-testing/storage"
)

// blockBootstrapper is the interface needed by base sync to deal with shards and meta nodes while they bootstrap
type blockBootstrapper interface {
	getCurrHeader() (data.HeaderHandler, error)
	getPrevHeader(data.HeaderHandler, storage.Storer) (data.HeaderHandler, error)
	getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	getHeaderWithHashRequestingIfMissing(hash []byte) (data.HeaderHandler, error)
	getHeaderWithNonceRequestingIfMissing(nonce uint64) (data.HeaderHandler, error)
	haveHeaderInPoolWithNonce(nonce uint64) bool
	getBlockBodyRequestingIfMissing(headerHandler data.HeaderHandler) (data.BodyHandler, error)
	isForkTriggeredByMeta() bool
}

// syncStarter defines the behavior of component that can start sync-ing blocks
type syncStarter interface {
	SyncBlock() error
}
