package ethrpc

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHexIntUnmarshal(t *testing.T) {
	test := struct {
		ID hexInt `json:"id"`
	}{}

	data := []byte(`{"id": "0x1cc348"}`)
	err := json.Unmarshal(data, &test)

	require.Nil(t, err)
	require.Equal(t, hexInt(1885000), test.ID)
}

func TestHexBigUnmarshal(t *testing.T) {
	test := struct {
		ID hexBig `json:"id"`
	}{}

	data := []byte(`{"id": "0x51248487c7466b7062d"}`)
	err := json.Unmarshal(data, &test)

	require.Nil(t, err)
	b := big.Int{}
	b.SetString("23949082357483433297453", 10)

	require.Equal(t, hexBig(b), test.ID)
}

func TestSyncingUnmarshal(t *testing.T) {
	syncing := new(Syncing)
	err := json.Unmarshal([]byte("0"), syncing)
	require.NotNil(t, err)

	data := []byte(`{
		"startingBlock": "0x384",
		"currentBlock": "0x386",
		"highestBlock": "0x454"
	}`)

	err = json.Unmarshal(data, syncing)
	require.Nil(t, err)
	require.True(t, syncing.IsSyncing)
	require.Equal(t, 900, syncing.StartingBlock)
	require.Equal(t, 902, syncing.CurrentBlock)
	require.Equal(t, 1108, syncing.HighestBlock)
}

func TestBlockWithoutTxsUnmarshal(t *testing.T) {

	data := []byte(`{
	   "baseFeePerGas":"0x1ed813f503",
	   "difficulty":"0x1dd3d924c8d8ef",
	   "extraData":"0x4554482e4352415a59504f4f4c2e4f5247",
	   "gasLimit":"0x1ca35d3",
	   "gasUsed":"0x80b3a1",
	   "hash":"0x42d1c51129ff905010a668a222798f7e935751b1dd320179d38ecb4d3268c418",
	   "logsBloom":"0x10212082c1062093048220408e02124004099801004004082007294985488e08182c495c21020097c0087f0a004b8912326d84004954a820284413064020888842500200114c8928484022283c004024018280110060901284500441836715f918c09002420210401d88101402441970600014484a0035544b080514040801244d3081310500231386fa842000028014502b0a8b1367004800e200400418402002ea0a042d8c630002214892c0714da501840000e002900100722b02980c28002c08a00249220062204b49018002959528010304052022300008318684a020880030a92109416611881014089639228280e58c10b0008240c100608023b140d1",
	   "miner":"0x4f9bebe3adc3c7f647c0023c60f91ac9dffa52d5",
	   "mixHash":"0x01ecd40c9cba98c2f1c58b581f2b25f8ad40ec1790303cf6529401953813e89e",
	   "nonce":"0xf36ba9063ff9355a",
	   "number":"0xc8679a",
	   "parentHash":"0x7eca2df7c366f578bac81ce35c8f9c0999c72c6684642d04284f7a99b5bbc7fd",
	   "receiptsRoot":"0x29b0b30d0f969a702e37d539f459725610aff1fd306b078b45d70bdb185ab9a2",
	   "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	   "size":"0x9911",
	   "stateRoot":"0xbd4bf582b32e1d7413bfde4fbbe2c7f6d2807b909f14cd47ed4be20910bc7601",
	   "timestamp":"0x612e3674",
	   "totalDifficulty":"0x651197aca755e55d718",
	   "transactions":[
           "0x791a7b7f4adf85bbb3d9cf1862f3a85dde9df646174e455580a99e9a6a32aa37",
           "0xe0dd65e400a7956a58e3c640379bafd5c6d98db751b9ebb07283be4ee5b2321d"
	   ]
}`)

	response := new(proxyBlockWithoutTransactions)
	err := json.Unmarshal(data, response)
	require.Nil(t, err)

	block := response.toBlock()
	require.Equal(t, 13133722, block.Number)
	require.Equal(t, *big.NewInt(132474205443), block.BaseFeePerGas)
}

func TestBlockWithTxsUnmarshal(t *testing.T) {

	data := []byte(`{
	   "baseFeePerGas":"0x1ed813f503",
	   "difficulty":"0x1dd3d924c8d8ef",
	   "extraData":"0x4554482e4352415a59504f4f4c2e4f5247",
	   "gasLimit":"0x1ca35d3",
	   "gasUsed":"0x80b3a1",
	   "hash":"0x42d1c51129ff905010a668a222798f7e935751b1dd320179d38ecb4d3268c418",
	   "logsBloom":"0x10212082c1062093048220408e02124004099801004004082007294985488e08182c495c21020097c0087f0a004b8912326d84004954a820284413064020888842500200114c8928484022283c004024018280110060901284500441836715f918c09002420210401d88101402441970600014484a0035544b080514040801244d3081310500231386fa842000028014502b0a8b1367004800e200400418402002ea0a042d8c630002214892c0714da501840000e002900100722b02980c28002c08a00249220062204b49018002959528010304052022300008318684a020880030a92109416611881014089639228280e58c10b0008240c100608023b140d1",
	   "miner":"0x4f9bebe3adc3c7f647c0023c60f91ac9dffa52d5",
	   "mixHash":"0x01ecd40c9cba98c2f1c58b581f2b25f8ad40ec1790303cf6529401953813e89e",
	   "nonce":"0xf36ba9063ff9355a",
	   "number":"0xc8679a",
	   "parentHash":"0x7eca2df7c366f578bac81ce35c8f9c0999c72c6684642d04284f7a99b5bbc7fd",
	   "receiptsRoot":"0x29b0b30d0f969a702e37d539f459725610aff1fd306b078b45d70bdb185ab9a2",
	   "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	   "size":"0x9911",
	   "stateRoot":"0xbd4bf582b32e1d7413bfde4fbbe2c7f6d2807b909f14cd47ed4be20910bc7601",
	   "timestamp":"0x612e3674",
	   "totalDifficulty":"0x651197aca755e55d718",
	   "transactions":[
			{
				"blockHash": "0x3003694478c108eaec173afcb55eafbb754a0b204567329f623438727ffa90d8",
				"blockNumber": "0x83319",
				"from": "0x201354729f8d0f8b64e9a0c353c672c6a66b3857",
				"gas": "0x15f90",
				"gasPrice": "0x4a817c800",
				"hash": "0xfc7dcd42eb0b7898af2f52f7c5af3bd03cdf71ab8b3ed5b3d3a3ff0d91343cbe",
				"input": "0xe1fa8e8425f1af44eb895e4900b8be35d9fdc28744a6ef491c46ec8601990e12a58af0ed",
				"nonce": "0x6ba1",
				"to": "0xd10e3be2bc8f959bc8c41cf65f60de721cf89adf",
				"transactionIndex": "0x3",
				"value": "0x0"
    		}
	   ]
}`)

	response := new(proxyBlockWithTransactions)
	err := json.Unmarshal(data, response)
	require.Nil(t, err)

	block := response.toBlock()
	require.Equal(t, 13133722, block.Number)
	require.Equal(t, *big.NewInt(132474205443), block.BaseFeePerGas)
}

func TestTransactionUnmarshal(t *testing.T) {
	tx := new(Transaction)
	err := json.Unmarshal([]byte("111"), tx)
	require.NotNil(t, err)

	data := []byte(`{
        "blockHash": "0x3003694478c108eaec173afcb55eafbb754a0b204567329f623438727ffa90d8",
        "blockNumber": "0x83319",
        "from": "0x201354729f8d0f8b64e9a0c353c672c6a66b3857",
        "gas": "0x15f90",
        "gasPrice": "0x4a817c800",
        "hash": "0xfc7dcd42eb0b7898af2f52f7c5af3bd03cdf71ab8b3ed5b3d3a3ff0d91343cbe",
        "input": "0xe1fa8e8425f1af44eb895e4900b8be35d9fdc28744a6ef491c46ec8601990e12a58af0ed",
        "nonce": "0x6ba1",
        "to": "0xd10e3be2bc8f959bc8c41cf65f60de721cf89adf",
        "transactionIndex": "0x3",
        "value": "0x0"
    }`)

	err = json.Unmarshal(data, tx)

	require.Nil(t, err)
	require.Equal(t, "0x3003694478c108eaec173afcb55eafbb754a0b204567329f623438727ffa90d8", tx.BlockHash)
	require.Equal(t, 537369, *tx.BlockNumber)
	require.Equal(t, "0x201354729f8d0f8b64e9a0c353c672c6a66b3857", tx.From)
	require.Equal(t, 90000, tx.Gas)
	require.Equal(t, *big.NewInt(20000000000), tx.GasPrice)
	require.Equal(t, "0xfc7dcd42eb0b7898af2f52f7c5af3bd03cdf71ab8b3ed5b3d3a3ff0d91343cbe", tx.Hash)
	require.Equal(t, "0xe1fa8e8425f1af44eb895e4900b8be35d9fdc28744a6ef491c46ec8601990e12a58af0ed", tx.Input)
	require.Equal(t, 27553, tx.Nonce)
	require.Equal(t, "0xd10e3be2bc8f959bc8c41cf65f60de721cf89adf", tx.To)
	require.Equal(t, 3, *tx.TransactionIndex)
	require.Equal(t, *big.NewInt(0), tx.Value)
}

func TestLogUnmarshal(t *testing.T) {
	log := new(Log)
	err := json.Unmarshal([]byte("111"), log)
	require.NotNil(t, err)

	data := []byte(`{
        "address": "0xd10e3be2bc8f959bc8c41cf65f60de721cf89adf",
        "topics": ["0x78e4fc71ff7e525b3b4660a76336a2046232fd9bba9c65abb22fa3d07d6e7066"],
        "data": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "blockNumber": "0x7f2cd",
        "blockHash": "0x3757b6efd7f82e3a832f0ec229b2fa36e622033ae7bad76b95763055a69374f7",
        "transactionIndex": "0x1",
        "transactionHash": "0xecd8a21609fa852c08249f6c767b7097481da34b9f8d2aae70067918955b4e69",
        "logIndex": "0x6",
        "removed": false
    }`)

	err = json.Unmarshal(data, log)

	require.Nil(t, err)
	require.Equal(t, "0xd10e3be2bc8f959bc8c41cf65f60de721cf89adf", log.Address)
	require.Equal(t, []string{"0x78e4fc71ff7e525b3b4660a76336a2046232fd9bba9c65abb22fa3d07d6e7066"}, log.Topics)
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", log.Data)
	require.Equal(t, 520909, log.BlockNumber)
	require.Equal(t, "0x3757b6efd7f82e3a832f0ec229b2fa36e622033ae7bad76b95763055a69374f7", log.BlockHash)
	require.Equal(t, 1, log.TransactionIndex)
	require.Equal(t, "0xecd8a21609fa852c08249f6c767b7097481da34b9f8d2aae70067918955b4e69", log.TransactionHash)
	require.Equal(t, 6, log.LogIndex)
	require.Equal(t, false, log.Removed)
}

func TestTransactionReceiptUnmarshal(t *testing.T) {
	receipt := new(TransactionReceipt)
	err := json.Unmarshal([]byte("[1]"), receipt)
	require.NotNil(t, err)

	data := []byte(`{
        "blockHash": "0x3757b6efd7f82e3a832f0ec229b2fa36e622033ae7bad76b95763055a69374f7",
        "blockNumber": "0x7f2cd",
        "contractAddress": null,
        "cumulativeGasUsed": "0x13356",
        "gasUsed": "0x6384",
        "logs": [{
            "address": "0xd10e3be2bc8f959bc8c41cf65f60de721cf89adf",
            "topics": ["0x78e4fc71ff7e525b3b4660a76336a2046232fd9bba9c65abb22fa3d07d6e7066"],
            "data": "0x0000000000000000000000000000000000000000000000000000000000000000",
            "blockNumber": "0x7f2cd",
            "blockHash": "0x3757b6efd7f82e3a832f0ec229b2fa36e622033ae7bad76b95763055a69374f7",
            "transactionIndex": "0x1",
            "transactionHash": "0xecd8a21609fa852c08249f6c767b7097481da34b9f8d2aae70067918955b4e69",
            "logIndex": "0x6",
            "removed": false
        }],
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000020000000000000000000000000040000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000",
        "root": "0xe367ea197d629892e7b25ea246fba93cd8ae053d468cc5997a816cc85d660321",
        "transactionHash": "0xecd8a21609fa852c08249f6c767b7097481da34b9f8d2aae70067918955b4e69",
        "transactionIndex": "0x1"
    }`)

	err = json.Unmarshal(data, receipt)

	require.Nil(t, err)
	require.Equal(t, 1, len(receipt.Logs))
	require.Equal(t, "0x3757b6efd7f82e3a832f0ec229b2fa36e622033ae7bad76b95763055a69374f7", receipt.BlockHash)
	require.Equal(t, 520909, receipt.BlockNumber)
	require.Equal(t, "", receipt.ContractAddress)
	require.Equal(t, 78678, receipt.CumulativeGasUsed)
	require.Equal(t, 25476, receipt.GasUsed)
	require.Equal(t, "0x00000000000000000000000000000000000000000000000000000000000020000000000000000000000000040000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000", receipt.LogsBloom)
	require.Equal(t, "0xe367ea197d629892e7b25ea246fba93cd8ae053d468cc5997a816cc85d660321", receipt.Root)
	require.Equal(t, "0xecd8a21609fa852c08249f6c767b7097481da34b9f8d2aae70067918955b4e69", receipt.TransactionHash)
	require.Equal(t, 1, receipt.TransactionIndex)

	require.Equal(t, "0xd10e3be2bc8f959bc8c41cf65f60de721cf89adf", receipt.Logs[0].Address)
	require.Equal(t, []string{"0x78e4fc71ff7e525b3b4660a76336a2046232fd9bba9c65abb22fa3d07d6e7066"}, receipt.Logs[0].Topics)
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", receipt.Logs[0].Data)
	require.Equal(t, 520909, receipt.Logs[0].BlockNumber)
	require.Equal(t, "0x3757b6efd7f82e3a832f0ec229b2fa36e622033ae7bad76b95763055a69374f7", receipt.Logs[0].BlockHash)
	require.Equal(t, 1, receipt.Logs[0].TransactionIndex)
	require.Equal(t, "0xecd8a21609fa852c08249f6c767b7097481da34b9f8d2aae70067918955b4e69", receipt.Logs[0].TransactionHash)
	require.Equal(t, 6, receipt.Logs[0].LogIndex)
	require.Equal(t, false, receipt.Logs[0].Removed)
}
