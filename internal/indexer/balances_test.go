package indexer

import (
	"math/big"
	"testing"

	"github.com/dailycrypto-me/daily-indexer/internal/chain"
	"github.com/dailycrypto-me/daily-indexer/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBalancesInternal(t *testing.T) {
	// Prepare test data
	accounts := storage.Accounts([]storage.Account{
		{
			Address: "0x1111111111111111111111111111111111111111",
			Balance: big.NewInt(100),
		},
		{
			Address: "0x0DC0d841F962759DA25547c686fa440cF6C28C61",
			Balance: big.NewInt(50),
		},
	})
	trx := storage.Transaction{
		From:    "0x1111111111111111111111111111111111111111",
		To:      "0x0DC0d841F962759DA25547c686fa440cF6C28C61",
		GasCost: big.NewInt(1),
		Value:   big.NewInt(20),
	}

	accounts.UpdateBalances(trx.From, trx.To, trx.Value)

	// Validate the updated balances
	{
		acc := accounts.FindBalance("0x1111111111111111111111111111111111111111")
		assert.Equal(t, big.NewInt(80), acc.Balance, "UpdateBalancesInternal failed to update 'from' balance correctly")
	}

	{
		acc := accounts.FindBalance("0x0DC0d841F962759DA25547c686fa440cF6C28C61")
		assert.Equal(t, big.NewInt(70), acc.Balance, "UpdateBalancesInternal failed to update 'to' balance correctly")
	}
}

func TestUpdateBalances(t *testing.T) {
	// Prepare test data
	accounts := storage.Accounts([]storage.Account{
		{
			Address: "0x1111111111111111111111111111111111111111",
			Balance: big.NewInt(100),
		},
		{
			Address: "0x0DC0d841F962759DA25547c686fa440cF6C28C61",
			Balance: big.NewInt(50),
		},
	})
	trx := &chain.Transaction{
		Logs: []chain.EventLog{{
			Address:          "0x00000000000000000000000000000000000000fe",
			Data:             "0x000000000000000000000000000000000000000000000005d9da3b556bb3aa86",
			LogIndex:         "0",
			Removed:          false,
			Topics:           []string{"0x9310ccfcb8de723f578a9e4282ea9f521f05ae40dc08f3068dfad528a65ee3c7", "0x00000000000000000000000021db400dcb1ef3bc3aee4f3d028ec1939b7fadd6", "0x0000000000000000000000004beaf4ce3c239ac7195a1e422725c0465271fb42"},
			TransactionHash:  "0xd8c9296770c696b313128f1cc913b1a5e90ddc62b049ceb8a476b1125d65d3a4",
			TransactionIndex: "1",
		},
			{
				Address:  "0x00000000000000000000000000000000000000fe",
				Data:     "0x000000000000000000000000000000000000000000000000a7a44a964be1f30a",
				LogIndex: "0",
				Removed:  false,
				Topics: []string{
					"0xf0ec9e0f6add850a1738c5822244e26ffc3d1f14da7537aa240582b25af12ad0",
					"0x0000000000000000000000000dc0d841f962759da25547c686fa440cf6c28c61",
					"0x000000000000000000000000ed4d5f4f3641cbc056e466d15dbe2403e38056f8",
				},
				TransactionHash:  "0xe667503bfec2ade69c5e03398aa29a88e035931cadd2caf265c0c85345f3f40e",
				TransactionIndex: "105",
			},
		},
		Transaction: storage.Transaction{
			Value: big.NewInt(30),
			From:  "0x1111111111111111111111111111111111111111",
			To:    "0x0DC0d841F962759DA25547c686fa440cF6C28C61",
		},
	}

	// Invoke the method
	accounts.UpdateBalances(trx.From, trx.To, trx.Value)
	err := accounts.UpdateEvents(trx.ExtractLogs())

	if err != nil {
		t.Error("UpdateBalances failed to update balances correctly. Error: ", err)
	}

	// Validate the updated balances
	{
		acc := accounts.FindBalance("0x1111111111111111111111111111111111111111")
		if acc == nil || acc.Balance.Cmp(big.NewInt(70)) != 0 {
			t.Error("UpdateBalances failed to update 'from' balance correctly. Should be 70 but is ", acc.Balance.String())
		}
	}
	{
		acc := accounts.FindBalance("0x0DC0d841F962759DA25547c686fa440cF6C28C61")
		bigInt, _ := big.NewInt(0).SetString("12079862109893161818", 10)
		if acc == nil || acc.Balance.Cmp(bigInt) != 0 {
			t.Error("UpdateBalances failed to update 'to' balance correctly. Should be 12079862109893161818 but is ", acc.Balance.String())
		}
	}
}
