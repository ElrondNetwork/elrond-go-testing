package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go-testing/hashing"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
)

// ArgInterceptedTxBlockBody is the argument for the intercepted transaction block body
type ArgInterceptedTxBlockBody struct {
	TxBlockBodyBuff  []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}
