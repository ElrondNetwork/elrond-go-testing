package processor

import (
	"github.com/ElrondNetwork/elrond-go-testing/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-testing/process"
	"github.com/ElrondNetwork/elrond-go-testing/storage"
)

// ArgHdrInterceptorProcessor is the argument for the interceptor processor used for headers (shard, meta and so on)
type ArgHdrInterceptorProcessor struct {
	Headers       storage.Cacher
	HeadersNonces dataRetriever.Uint64SyncMapCacher
	HdrValidator  process.HeaderValidator
	BlackList     process.BlackListHandler
}
