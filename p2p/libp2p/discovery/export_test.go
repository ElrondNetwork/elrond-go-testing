package discovery

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-testing/p2p/libp2p"
)

func (kdd *KadDhtDiscoverer) RefreshInterval() time.Duration {
	return kdd.refreshInterval
}

func (kdd *KadDhtDiscoverer) InitialPeersList() []string {
	return kdd.initialPeersList
}

func (kdd *KadDhtDiscoverer) RandezVous() string {
	return kdd.randezVous
}

func (kdd *KadDhtDiscoverer) ContextProvider() *libp2p.Libp2pContext {
	return kdd.contextProvider
}

func (kdd *KadDhtDiscoverer) ConnectToOnePeerFromInitialPeersList(
	durationBetweenAttempts time.Duration,
	initialPeersList []string) <-chan struct{} {

	return kdd.connectToOnePeerFromInitialPeersList(durationBetweenAttempts, initialPeersList)
}
