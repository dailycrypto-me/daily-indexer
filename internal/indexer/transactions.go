package indexer

import (
	"math/big"

	"github.com/dailycrypto-me/daily-indexer/internal/chain"
	"github.com/dailycrypto-me/daily-indexer/internal/common"
	"github.com/dailycrypto-me/daily-indexer/internal/storage"
	"github.com/ethereum/go-ethereum/rlp"
	log "github.com/sirupsen/logrus"
)

func (bc *blockContext) processTransactions() (err error) {
	if len(bc.Block.Pbft.Transactions) == 0 {
		return
	}

	if len(bc.Block.Pbft.Transactions) != len(bc.Block.Transactions) {
		log.WithFields(log.Fields{"in_block": len(bc.Block.Pbft.Transactions), "transactions": len(bc.Block.Transactions), "traces": len(bc.Block.Traces)}).Error("Transactions count mismatch")
	}
	feeReward := big.NewInt(0)
	for t_idx := 0; t_idx < len(bc.Block.Transactions); t_idx++ {
		bc.Block.Transactions[t_idx].SetTimestamp(bc.Block.Pbft.Timestamp)

		bc.SaveTransaction(bc.Block.Transactions[t_idx].GetStorage(), false)

		trx_fee := bc.Block.Transactions[t_idx].GetFee()
		feeReward.Add(feeReward, trx_fee)
		// Remove fee from sender balance
		bc.accounts.AddToBalance(bc.Block.Transactions[t_idx].From, big.NewInt(0).Neg(trx_fee))
		if !bc.Block.Transactions[t_idx].Status {
			continue
		}
		// remove value from sender and add to receiver
		receiver := bc.Block.Transactions[t_idx].To
		// handle contract creation
		if receiver == "" {
			receiver = bc.Block.Transactions[t_idx].ContractAddress
		}
		bc.accounts.UpdateBalances(bc.Block.Transactions[t_idx].From, receiver, bc.Block.Transactions[t_idx].Value)

		// process logs
		err = bc.processTransactionLogs(bc.Block.Transactions[t_idx])
		if err != nil {
			return
		}
		if len(bc.Block.Traces) > 0 {
			if internal_transactions := bc.processInternalTransactions(bc.Block.Traces[t_idx], t_idx, bc.Block.Transactions[t_idx].GasPrice); internal_transactions != nil {
				bc.Batch.AddSingleKey(internal_transactions, bc.Block.Transactions[t_idx].Hash)
			}
		}
	}
	// add total fee to the block producer balance before the magnolia hardfork
	if bc.Config.Chain != nil && (bc.Block.Pbft.Number < bc.Config.Chain.Hardforks.MagnoliaHf.BlockNum) {
		bc.accounts.AddToBalance(bc.Block.Pbft.Author, feeReward)
	}
	return
}

func (bc *blockContext) processInternalTransactions(trace chain.TransactionTrace, t_idx int, gasPrice uint64) (internal_transactions *storage.InternalTransactionsResponse) {
	if len(trace.Trace) <= 1 {
		return
	}
	internal_transactions = new(storage.InternalTransactionsResponse)
	internal_transactions.Data = make([]storage.Transaction, 0, len(trace.Trace)-1)

	for e_idx, entry := range trace.Trace {
		if e_idx == 0 {
			continue
		}
		internal := makeInternal(bc.Block.Transactions[t_idx].GetStorage(), entry, gasPrice)
		internal_transactions.Data = append(internal_transactions.Data, internal)

		bc.SaveTransaction(internal, true)
		if entry.Action.CallType != "delegatecall" {
			bc.accounts.UpdateBalances(internal.From, internal.To, internal.Value)
		}
	}
	return
}

func makeInternal(trx storage.Transaction, entry chain.TraceEntry, gasPrice uint64) (internal storage.Transaction) {
	internal = trx
	internal.From = entry.Action.From
	internal.To = chain.GetInternalTransactionTarget(entry)
	internal.Value = common.ParseStringToBigInt(entry.Action.Value)
	internal.GasCost = chain.GetTransactionFee(common.ParseUInt(entry.Result.GasUsed), gasPrice)
	internal.Type = chain.GetTransactionType(trx.To, entry.Action.Input, entry.Type, true)
	internal.BlockNumber = trx.BlockNumber
	return
}

func (bc *blockContext) SaveTransaction(trx storage.Transaction, internal bool) {
	log.WithFields(log.Fields{"from": trx.From, "to": trx.To, "hash": trx.Hash}).Trace("Saving transaction")

	// As the same data is saved with a different keys, it is better to serialize it only once
	trx_bytes, err := rlp.EncodeToBytes(trx)
	if err != nil {
		log.WithFields(log.Fields{"from": trx.From, "to": trx.To, "hash": trx.Hash}).Error("Failed to encode transaction")
	}
	from_index := bc.addressStats.GetAddress(bc.Storage, trx.From).AddTransaction(trx.Timestamp)
	bc.Batch.AddSerialized(trx, trx_bytes, trx.From, from_index)
	if trx.To != "" {
		to_index := bc.addressStats.GetAddress(bc.Storage, trx.To).AddTransaction(trx.Timestamp)
		bc.Batch.AddSerialized(trx, trx_bytes, trx.To, to_index)
	}

	if !internal {
		bc.Batch.AddSerializedSingleKey(trx, trx_bytes, trx.Hash)
	}
}
