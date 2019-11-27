package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-testing/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-testing/process"
	"github.com/ElrondNetwork/elrond-go-testing/process/mock"
	"github.com/ElrondNetwork/elrond-go-testing/process/unsigned"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedUnsignedTxDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedUnsignedTxDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.AddrConv = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestInterceptedUnsignedTxDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.NotNil(t, imh)
	assert.Nil(t, err)

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &smartContractResult.SmartContractResult{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)
	interceptedData, err := imh.Create(emptyTxBuff)

	_, ok := interceptedData.(*unsigned.InterceptedUnsignedTransaction)
	assert.True(t, ok)
}
