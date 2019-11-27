package processor

import (
	"github.com/ElrondNetwork/elrond-go-testing/hashing"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
	"github.com/ElrondNetwork/elrond-go-testing/storage"
)

// ArgTxBodyInterceptorProcessor is the argument for the interceptor processor used for tx block body
type ArgTxBodyInterceptorProcessor struct {
	MiniblockCache   storage.Cacher
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}
