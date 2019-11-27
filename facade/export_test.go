package facade

import (
	"github.com/ElrondNetwork/elrond-go-testing/ntp"
)

// GetSyncer returns the current syncer
func (ef *ElrondNodeFacade) GetSyncer() ntp.SyncTimer {
	return ef.syncer
}
