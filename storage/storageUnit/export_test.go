package storageUnit

import (
	"github.com/ElrondNetwork/elrond-go-testing/storage"
)

func (s *Unit) GetBlomFilter() storage.BloomFilter {
	return s.bloomFilter
}
