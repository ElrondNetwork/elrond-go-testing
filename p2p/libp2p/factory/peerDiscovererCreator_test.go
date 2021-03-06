package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-testing/config"
	"github.com/ElrondNetwork/elrond-go-testing/p2p"
	"github.com/ElrondNetwork/elrond-go-testing/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-testing/p2p/libp2p/factory"
	"github.com/stretchr/testify/assert"
)

func TestPeerDiscovererCreator_CreatePeerDiscovererNoDiscoveryEnabledShouldRetNullDiscoverer(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled: false,
		},
	}

	f := factory.NewPeerDiscovererCreator(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	_, ok := pDiscoverer.(*discovery.NullDiscoverer)

	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestPeerDiscovererCreator_CreatePeerDiscovererKadIntervalLessThanZeroShouldErr(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: -1,
		},
	}

	f := factory.NewPeerDiscovererCreator(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	assert.Nil(t, pDiscoverer)
	assert.Equal(t, p2p.ErrNegativeOrZeroPeersRefreshInterval, err)
}

func TestPeerDiscovererCreator_CreatePeerDiscovererKadOkValsShouldWork(t *testing.T) {
	p2pConfig := config.P2PConfig{
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
		},
	}

	f := factory.NewPeerDiscovererCreator(p2pConfig)
	pDiscoverer, err := f.CreatePeerDiscoverer()

	_, ok := pDiscoverer.(*discovery.KadDhtDiscoverer)

	assert.NotNil(t, pDiscoverer)
	assert.True(t, ok)
	assert.Nil(t, err)
}
