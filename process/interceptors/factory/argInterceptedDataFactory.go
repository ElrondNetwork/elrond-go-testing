package factory

import (
	"github.com/ElrondNetwork/elrond-go-testing/crypto"
	"github.com/ElrondNetwork/elrond-go-testing/data/state"
	"github.com/ElrondNetwork/elrond-go-testing/hashing"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/process"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
)

// ArgInterceptedDataFactory holds all dependencies required by the shard and meta intercepted data factory in order to create
// new instances
type ArgInterceptedDataFactory struct {
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
	MultiSigVerifier crypto.MultiSigVerifier
	NodesCoordinator sharding.NodesCoordinator
	KeyGen           crypto.KeyGenerator
	BlockKeyGen      crypto.KeyGenerator
	Signer           crypto.SingleSigner
	BlockSigner      crypto.SingleSigner
	AddrConv         state.AddressConverter
	FeeHandler       process.FeeHandler
}
