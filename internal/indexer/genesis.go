package indexer

import (
	"math/big"

	"github.com/dailycrypto-me/daily-indexer/internal/chain"
	"github.com/dailycrypto-me/daily-indexer/internal/common"
	"github.com/dailycrypto-me/daily-indexer/internal/storage"
	log "github.com/sirupsen/logrus"
)

type Genesis struct {
	storage storage.Storage
	genesis chain.GenesisObject
	bc      blockContext
	hash    string
}

func MakeGenesis(s storage.Storage, c chain.Client, gen_obj chain.GenesisObject, genesisHash storage.GenesisHash) *Genesis {
	var genesis Genesis
	genesis.storage = s
	genesis.genesis = gen_obj
	genesis.hash = string(genesisHash)
	genesis.bc = *MakeBlockContext(s, c, &common.Config{Chain: gen_obj.ToChainConfig()})

	return &genesis
}

func (g *Genesis) makeInitBalanceTrx(addr string, value *big.Int) (trx storage.Transaction) {
	trx.Hash = "GENESIS_" + addr
	trx.From = "GENESIS"
	trx.To = addr
	trx.Value = value
	trx.BlockNumber = 0
	trx.Timestamp = g.genesis.DagGenesisBlock.Timestamp
	trx.Status = true
	return
}

func (g *Genesis) process() {
	genesisSupply := big.NewInt(0)
	accounts := g.storage.GetAccounts()
	for addr, value := range g.genesis.InitialBalances {
		value := common.ParseStringToBigInt(value)
		trx := g.makeInitBalanceTrx(addr, value)
		g.bc.SaveTransaction(trx, false)
		genesisSupply.Add(genesisSupply, value)
		accounts.AddToBalance(trx.To, value)
	}
	for _, validator := range g.genesis.Dpos.InitialValidators {
		for addr, value := range validator.Delegations {
			delegation := common.ParseStringToBigInt(value)
			accounts.AddToBalance(addr, big.NewInt(0).Neg(delegation))
			accounts.AddToBalance(common.DposContractAddress, delegation)
		}
	}
	log.WithField("count", len(g.genesis.InitialBalances)).Info("Genesis: Init balance transactions parsed")

	// Genesis transactions isn't real transactions, so don't count it here
	g.bc.Batch.SaveAccounts(accounts)
	g.bc.finalized.TrxCount = 0
	g.bc.Batch.SetGenesisHash(storage.GenesisHash(g.hash))
	g.bc.Batch.SetTotalSupply(genesisSupply)
	g.bc.commit()
}
