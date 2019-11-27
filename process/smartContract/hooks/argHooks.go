package hooks

import (
	"github.com/ElrondNetwork/elrond-go-testing/data"
	"github.com/ElrondNetwork/elrond-go-testing/data/state"
	"github.com/ElrondNetwork/elrond-go-testing/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-testing/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
)

// ArgBlockChainHook represents the arguments structure for the blockchain hook
type ArgBlockChainHook struct {
	Accounts         state.AccountsAdapter
	AddrConv         state.AddressConverter
	StorageService   dataRetriever.StorageService
	BlockChain       data.ChainHandler
	ShardCoordinator sharding.Coordinator
	Marshalizer      marshal.Marshalizer
	Uint64Converter  typeConverters.Uint64ByteSliceConverter
}
