package mock

import (
	"github.com/ElrondNetwork/elrond-go-testing/p2p"
)

type PeerDiscovererStub struct {
	BootstrapCalled    func() error
	CloseCalled        func() error
	ApplyContextCalled func(ctxProvider p2p.ContextProvider) error
}

func (pds *PeerDiscovererStub) Bootstrap() error {
	return pds.BootstrapCalled()
}

func (pds *PeerDiscovererStub) Name() string {
	return "PeerDiscovererStub"
}

func (pds *PeerDiscovererStub) ApplyContext(ctxProvider p2p.ContextProvider) error {
	return pds.ApplyContextCalled(ctxProvider)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pds *PeerDiscovererStub) IsInterfaceNil() bool {
	if pds == nil {
		return true
	}
	return false
}
