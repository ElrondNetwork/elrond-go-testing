package mock

import (
	"github.com/ElrondNetwork/elrond-go-testing/data"
	"github.com/ElrondNetwork/elrond-go-testing/process"
)

type TxTypeHandlerMock struct {
	ComputeTransactionTypeCalled func(tx data.TransactionHandler) (process.TransactionType, error)
}

func (th *TxTypeHandlerMock) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, error) {
	if th.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, nil
	}

	return th.ComputeTransactionTypeCalled(tx)
}
