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
