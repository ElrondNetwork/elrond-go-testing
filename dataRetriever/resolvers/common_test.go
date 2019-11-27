package resolvers_test

import (
	"github.com/ElrondNetwork/elrond-go-testing/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-testing/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go-testing/p2p"
)

func createRequestMsg(dataType dataRetriever.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataType, Value: val})
	return &mock.P2PMessageMock{DataField: buff}
}
