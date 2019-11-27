package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go-testing/crypto"
	"github.com/ElrondNetwork/elrond-go-testing/hashing"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
)

// ArgInterceptedBlockHeader is the argument for the intercepted header
type ArgInterceptedBlockHeader struct {
	HdrBuff           []byte
	Marshalizer       marshal.Marshalizer
	Hasher            hashing.Hasher
	SingleSigVerifier crypto.SingleSigner
	MultiSigVerifier  crypto.MultiSigVerifier
	NodesCoordinator  sharding.NodesCoordinator
	ShardCoordinator  sharding.Coordinator
	KeyGen            crypto.KeyGenerator
}
