package storage

import (
	"fmt"

	"github.com/dailycrypto-me/daily-indexer/models"
	"github.com/ethereum/go-ethereum/rlp"
	log "github.com/sirupsen/logrus"
)

type Storage interface {
	Clean() error
	Close() error
	ForEach(o interface{}, key_prefix string, start *uint64, fn func(key, res []byte) (stop bool))
	ForEachBackwards(o interface{}, key_prefix string, start *uint64, fn func(key, res []byte) (stop bool))
	ForEachFromKey(prefix, start_key []byte, fn func(key, res []byte) (stop bool))
	ForEachFromKeyBackwards(prefix, start_key []byte, fn func(key, res []byte) (stop bool))
	NewBatch() Batch
	GetTotalSupply() *TotalSupply
	GetAccounts() Accounts
	GetWeekStats(year, week int32) WeekStats
	GetFinalizationData() *FinalizationData
	GetAddressStats(addr string) *AddressStats
	GenesisHashExist() bool
	GetGenesisHash() GenesisHash
	GetTransactionByHash(hash string) Transaction
	GetInternalTransactions(hash string) InternalTransactionsResponse
	GetTransactionLogs(hash string) models.TransactionLogsResponse
	GetValidatorYield(validator string, block uint64) (res Yield)
	GetTotalYield(block uint64) (res Yield)
}

func GetTotal[T Paginated](s Storage, address string) (r uint64) {
	stats := s.GetAddressStats(address)

	var o T
	switch t := any(o).(type) {
	case models.Dag:
		r = stats.DagsCount
	case models.Pbft:
		r = stats.PbftCount
	case Transaction:
		r = stats.TransactionsCount
	default:
		log.WithField("type", t).Fatal("GetCount incorrect type passed")
	}
	return
}

func GetObjectsPage[T Paginated](s Storage, address string, from, count uint64) (ret []T, pagination *models.PaginatedResponse) {

	pagination = new(models.PaginatedResponse)
	pagination.Start = from
	pagination.Total = GetTotal[T](s, address)
	end := from + count
	pagination.HasNext = (end < pagination.Total)
	if end > pagination.Total {
		end = pagination.Total
	}
	pagination.End = end
	ret = make([]T, 0, count)
	start := pagination.Total - from
	s.ForEachBackwards(new(T), address, &start, func(_, res []byte) (stop bool) {
		var o T
		err := rlp.DecodeBytes(res, &o)
		if err != nil {
			log.WithFields(log.Fields{"type": GetTypeName[T](), "error": err}).Fatal("Error decoding data from db")
		}
		ret = append(ret, o)
		if uint64(len(ret)) == count {
			return true
		}
		return
	})
	return
}

func GetHoldersPage(s Storage, from, count uint64) (ret []models.Account, pagination *models.PaginatedResponse) {
	holders := s.GetAccounts()
	pagination = new(models.PaginatedResponse)
	pagination.Start = from
	pagination.Total = uint64(len(holders))
	end := from + count
	pagination.HasNext = (end < pagination.Total)
	if end > pagination.Total {
		end = pagination.Total
	}
	pagination.End = end

	ret = make([]models.Account, 0, count)
	for i := from; i < end; i++ {
		ret = append(ret, holders[i].ToModel())
	}
	return
}

func ProcessIntervalData[T Yields](s Storage, start uint64, fn func([]byte, T) (stop bool)) {
	s.ForEach(new(T), "", &start, func(key, res []byte) bool {
		var o T
		err := rlp.DecodeBytes(res, &o)
		if err != nil {
			log.WithFields(log.Fields{"type": GetTypeName[T](), "error": err}).Fatal("Error decoding data from db")
		}
		return fn(key, o)
	})
}

func GetUIntKey(key uint64) string {
	return fmt.Sprintf("%020d", key)
}
