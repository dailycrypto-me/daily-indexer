package common

import (
	"encoding/json"
	"testing"

	"github.com/dailycrypto-me/daily-indexer/models"
	"github.com/stretchr/testify/assert"
)

func TestProcessTransaction(t *testing.T) {
	tx := models.Transaction{
		Hash:  "abc123",
		From:  "0x99a2d5feaecb1a729d4f9af4197cc03bb9a37bc3",
		To:    "0x00000000000000000000000000000000000000fe",
		Input: "0x5c19a95c000000000000000000000000ed4d5f4f3641cbc056e466d15dbe2403e38056f8",
	}

	err := ProcessTransaction(&tx)
	t.Log(tx.Calldata)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tx.Calldata == nil {
		t.Error("Expected Calldata to be set")
	}
	assert.Equal(t, tx.Calldata.Params, []any{"0xed4d5f4f3641cbc056e466d15dbe2403e38056f8"})
}

func TestDecodeTransaction(t *testing.T) {
	// Correct sample input data
	transaction_json := `{
		"blockHash": "0xa3505936a13302b1bdab9cbbff23c9d922842b738c0f18094717d072ae88be59",
		"blockNumber": "0x5487c",
		"from": "0x3aef315b67c2e23aae9e9c1174d05371e2c9b144",
		"gas": "0x5b8d80",
		"gasPrice": "0x0",
		"gasUsed": "0x5af868",
		"hash": "0x9388dfe7caa6cbf19bd6b9508b94e563a32d23d303bf019974bdbcc324c953b7",
		"nonce": "0x0",
		"r": "0xe41fe3baf6ef792fa3d33af556767c84528ee3b8387fcd94d33fb3f12bff2d03",
		"s": "0x4d4462271713d1b1808368952cf44bef70fdd8304246cf5ee0ffa38d771b55d9",
		"to": "0xf3b803a8f4c4fc3fbe454b6438dc0ed22735f01b",
		"transactionIndex": "0x0",
		"v": "0x0",
		"value": "0x82f79cd9000",
		"input": "0x2ada8a320000000000000000000000003aef315b67c2e23aae9e9c1174d05371e2c9b1440000000000000000000000000000000000000000000000206f62c0f445450000000000000000000000000000000000000000000000000000000000000000719900000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000041dd69f77d6f94270a009a6917015acd44eab669684a2f24004d8807d9cae501de57126b392a58df3b4bd5591d0547d0f5b2aaf192f668aae7616ce86f71c861571b00000000000000000000000000000000000000000000000000000000000000"
	}`

	dpos_transaction_json := `{
		"blockHash": "0x3c8813fdae9b0bcdb2fb2a8bffd7c89ed06f84229a5e968bb9ba534a6606e606",
		"blockNumber": "0x5487c",
		"from": "0x0dc0d841f962759da25547c686fa440cf6c28c61",
		"gas": "0x5b8d80",
		"gasPrice": "0x0",
		"gasUsed": "0x5af868",
		"hash": "0x31254d1bd2d5ae8f5bcedeb1e9042720c82348404db97993b6a34c25307fe771",
		"nonce": "0x0",
		"r": "0xe41fe3baf6ef792fa3d33af556767c84528ee3b8387fcd94d33fb3f12bff2d03",
		"s": "0x4d4462271713d1b1808368952cf44bef70fdd8304246cf5ee0ffa38d771b55d9",
		"to": "0x00000000000000000000000000000000000000fe",
		"transactionIndex": "0x0",
		"v": "0x0",
		"value": "0x82f79cda",
		"input": "0x5c19a95c000000000000000000000000ed4d5f4f3641cbc056e466d15dbe2403e38056f8"
	}`

	multisend_transaction_json := `{
		"blockHash": "0xc89430f694b602f62a41b041b2e09e5b035e2243bf57d148ba9b10686fd8d026",
		"blockNumber": "0x5487c",
		"from": "0x99a2d5feaecb1a729d4f9af4197cc03bb9a37bc3",
		"gas": "0x5b8d80",
		"gasPrice": "0x0",
		"gasUsed": "0x5af868",
		"hash": "0xe400436607c11b022d25b7cd363578a406c83ababff74ab10e47b29ffbecb9bc",
		"nonce": "0x0",
		"r": "0xe41fe3baf6ef792fa3d33af556767c84528ee3b8387fcd94d33fb3f12bff2d03",
		"s": "0x4d4462271713d1b1808368952cf44bef70fdd8304246cf5ee0ffa38d771b55d9",
		"to": "0x1578f035581f664efa85a6da822464bd9edd8850",
		"transactionIndex": "0x0",
		"v": "0x0",
		"value": "0x82f79cda",
		"input": "0xd2d745b100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000340000000000000000000000000000000000000000000000000000000000000001700000000000000000000000068ecf99daab887c221f4fe5ff325de9558153b99000000000000000000000000ab9baef3c6cc92fcb37e4b67dab58bbee1673590000000000000000000000000933059d2cd510896758c5aeda7d0e6449d576783000000000000000000000000360e1c8ad2d461e1e711ff7ca27666ba3e338dfd000000000000000000000000a6f3c76f18e2d006dfc05ec49298af5ad5b131150000000000000000000000007bcaf39fef1018e79ceabde60ce4db995687ccaf000000000000000000000000336ba65ebffd13acd91aad33e1f4315f8811d692000000000000000000000000f9188f1f88e50f137074d85c9d59c693cb4b1e0a0000000000000000000000000dc0d841f962759da25547c686fa440cf6c28c61000000000000000000000000f1c587a22fbf80af80446fa17e7322952f18456c000000000000000000000000cddb0d484ca1c625ffca0882396ef34ffff242e3000000000000000000000000e749754ec8cf03b1b9e10b81270b73af66e40f5400000000000000000000000010ce6f9c7c22f82214c40755b3eea5f126a7148d000000000000000000000000d42eaa28c5eafee9a0040a7ac74dd3f4b57678bd0000000000000000000000009184eec720322f3a1bee38aa2879e602ead46486000000000000000000000000ec15db470db85cc75b0e3fa5b6a0c607a5e8c64a00000000000000000000000007dbc3e2ca7894dc00c11cf460e4f4c43e888888000000000000000000000000a4195def477491ef7f00b8688c9b8032cd71bb2a0000000000000000000000008f1567bb4381f4ed53dbeb3c0dca5c4f189a111000000000000000000000000052124d5982576507dd4a18d6607225e64be168bb0000000000000000000000002b34ec1e751bbd345e8b2686f16044425fcc525b000000000000000000000000b56f7b1cdecca0a4915979dbdec29ec63315cdfd0000000000000000000000009574d76c2df7373c162715f9bf5ab93170746ec100000000000000000000000000000000000000000000000000000000000000170000000000000000000000000000000000000000001b929b9ee181f57c15000000000000000000000000000000000000000000000000eb49743ab7882ecd800000000000000000000000000000000000000000000001d692e8756f105d9b00000000000000000000000000000000000000000000000583b8b963daafbd978000000000000000000000000000000000000000000000008168665351115e4d8000000000000000000000000000000000000000000000040b43329e160997328000000000000000000000000000000000000000000000005e1d61afaa6a373600000000000000000000000000000000000000000000000205a1994d444579360000000000000000000000000000000000000000000000005e1d61afaa6a37360000000000000000000000000000000000000000000000005e1d61afaa6a3736000000000000000000000000000000000000000000000002c1dc5cb026988c6880000000000000000000000000000000000000000000000046960943bfcfa9688000000000000000000000000000000000000000000000005e1d61afaa6a373600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000eb49743ab7882ecd8000000000000000000000000000000000000000000000008168665351115e4d80000000000000000000000000000000000000000000000930de8a4b2b51d40700000000000000000000000000000000000000000000000075a4ba1f228369ca00000000000000000000000000000000000000000000000075a4ba1f228369ca00000000000000000000000000000000000000000000000930de8a4b2b51d40700000000000000000000000000000000000000000000000930de8a4b2b51d40700000000000000000000000000000000000000000000003d070d58f68278b7f20000"
	}`

	var trx models.Transaction
	var dpos_trx models.Transaction
	var multisend_trx models.Transaction
	_ = json.Unmarshal([]byte(transaction_json), &trx)
	_ = json.Unmarshal([]byte(dpos_transaction_json), &dpos_trx)
	_ = json.Unmarshal([]byte(multisend_transaction_json), &multisend_trx)

	{
		functionName, params, err := DecodeTransaction(trx)
		assert.NoError(t, err)
		assert.Equal(t, "claim(address,uint256,uint256,bytes)", functionName)
		assert.Equal(t, []any{
			"0x3aef315b67c2e23aae9e9c1174d05371e2c9b144",
			"598322000000000000000",
			"29081",
			"0xdd69f77d6f94270a009a6917015acd44eab669684a2f24004d8807d9cae501de57126b392a58df3b4bd5591d0547d0f5b2aaf192f668aae7616ce86f71c861571b",
		}, params)
	}
	{
		functionName, params, err := DecodeTransaction(dpos_trx)
		assert.NoError(t, err)
		assert.Equal(t, "delegate(address)", functionName)
		assert.Equal(t, []any{
			"0xed4d5f4f3641cbc056e466d15dbe2403e38056f8",
		}, params)
	}

	{
		functionName, params, err := DecodeTransaction(multisend_trx)
		assert.NoError(t, err)
		assert.Equal(t, "multisendToken(address[],uint256[])", functionName)
		expected := []any{
			[]any{"0x68ecf99daab887c221f4fe5ff325de9558153b99", "0xab9baef3c6cc92fcb37e4b67dab58bbee1673590", "0x933059d2cd510896758c5aeda7d0e6449d576783", "0x360e1c8ad2d461e1e711ff7ca27666ba3e338dfd", "0xa6f3c76f18e2d006dfc05ec49298af5ad5b13115", "0x7bcaf39fef1018e79ceabde60ce4db995687ccaf", "0x336ba65ebffd13acd91aad33e1f4315f8811d692", "0xf9188f1f88e50f137074d85c9d59c693cb4b1e0a", "0x0dc0d841f962759da25547c686fa440cf6c28c61", "0xf1c587a22fbf80af80446fa17e7322952f18456c", "0xcddb0d484ca1c625ffca0882396ef34ffff242e3", "0xe749754ec8cf03b1b9e10b81270b73af66e40f54", "0x10ce6f9c7c22f82214c40755b3eea5f126a7148d", "0xd42eaa28c5eafee9a0040a7ac74dd3f4b57678bd", "0x9184eec720322f3a1bee38aa2879e602ead46486", "0xec15db470db85cc75b0e3fa5b6a0c607a5e8c64a", "0x07dbc3e2ca7894dc00c11cf460e4f4c43e888888", "0xa4195def477491ef7f00b8688c9b8032cd71bb2a", "0x8f1567bb4381f4ed53dbeb3c0dca5c4f189a1110", "0x52124d5982576507dd4a18d6607225e64be168bb", "0x2b34ec1e751bbd345e8b2686f16044425fcc525b", "0xb56f7b1cdecca0a4915979dbdec29ec63315cdfd", "0x9574d76c2df7373c162715f9bf5ab93170746ec1"},
			[]any{"33333333330000000000000000", "1111111111000000000000000", "2222222222000000000000000", "6666666667000000000000000", "611111111000000000000000", "4888888889000000000000000", "444444444000000000000000", "2444444444000000000000000", "444444444000000000000000", "444444444000000000000000", "3333333333000000000000000", "333333333000000000000000", "444444444000000000000000", "0", "0", "1111111111000000000000000", "611111111000000000000000", "11111111110000000000000000", "555555556000000000000000", "555555556000000000000000", "11111111110000000000000000", "11111111110000000000000000", "73777777780000000000000000"},
		}
		assert.Equal(t, expected, params)
		expected_json, _ := json.Marshal(expected)
		params_json, _ := json.Marshal(params)
		assert.Equal(t, string(expected_json), string(params_json))
		t.Log(string(params_json))
	}
}
