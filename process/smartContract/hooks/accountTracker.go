package hooks

import "github.com/ElrondNetwork/elrond-go-testing/data/state"

type TempAccountTracker struct {
}

func (t *TempAccountTracker) SaveAccount(accountHandler state.AccountHandler) error {
	return nil
}

func (t *TempAccountTracker) Journalize(entry state.JournalEntry) {
}

func (t *TempAccountTracker) IsInterfaceNil() bool {
	return t == nil
}
