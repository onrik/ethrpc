package ethrpc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type EthRPCTestSuite struct {
	suite.Suite
	rpc *EthRPC
}

func (s *EthRPCTestSuite) registerResponse(result string, callback func([]byte)) {
	httpmock.Reset()
	response := fmt.Sprintf(`{"jsonrpc":"2.0", "id":1, "result": %s}`, result)
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		callback(s.getBody(request))
		return httpmock.NewStringResponse(200, response), nil
	})
}

func (s *EthRPCTestSuite) registerResponseError(err error) {
	httpmock.Reset()
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		return nil, err
	})
}

func (s *EthRPCTestSuite) getBody(request *http.Request) []byte {
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	s.Require().Nil(err)

	return body
}

func (s *EthRPCTestSuite) methodEqual(body []byte, expected string) {
	value := gjson.GetBytes(body, "method").String()

	s.Require().Equal(expected, value)
}

func (s *EthRPCTestSuite) paramsEqual(body []byte, expected string) {
	value := gjson.GetBytes(body, "params").Raw
	if expected == "null" {
		s.Require().Equal(expected, value)
	} else {
		s.JSONEq(expected, value)
	}
}

func (s *EthRPCTestSuite) SetupSuite() {
	s.rpc = NewEthRPC("http://127.0.0.1:8545", WithHttpClient(http.DefaultClient), WithLogger(nil), WithDebug(false))

	httpmock.Activate()
}

func (s *EthRPCTestSuite) TearDownSuite() {
	httpmock.Deactivate()
}

func (s *EthRPCTestSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *EthRPCTestSuite) TestURL() {
	s.Require().Equal(s.rpc.url, s.rpc.URL())
}

func (s *EthRPCTestSuite) TestWeb3ClientVersion() {
	response := `{"jsonrpc":"2.0", "id":1, "result": "test client"}`

	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "web3_clientVersion")
		s.paramsEqual(body, `null`)

		return httpmock.NewStringResponse(200, response), nil
	})

	v, err := s.rpc.Web3ClientVersion()
	s.Require().Nil(err)
	s.Require().Equal("test client", v)
}

func (s *EthRPCTestSuite) TestCall() {
	// Test http error
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		return nil, errors.New("Error")
	})

	_, err := s.rpc.Call("test")
	s.Require().NotNil(err)
	httpmock.Reset()

	// Test invalid response format
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(200, "{213"), nil
	})
	_, err = s.rpc.Call("test")
	s.Require().NotNil(err)
	httpmock.Reset()

	// Test eth error
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(200, `{"error": {"code": 21, "message": "eee"}}`), nil
	})
	_, err = s.rpc.Call("test")
	s.Require().NotNil(err)
	ethError, ok := err.(EthError)
	s.Require().True(ok)
	s.Require().Equal(21, ethError.Code)
	s.Require().Equal("eee", ethError.Message)
}

func (s *EthRPCTestSuite) Test_call() {
	// Test http error
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("Error")
	})
	err := s.rpc.call("test", nil)
	s.Require().NotNil(err)

	// Test target is nil
	s.registerResponse(`{"foo": "bar"}`, func([]byte) {})
	err = s.rpc.call("test", nil)
	s.Require().Nil(err)

	// Test invalid target
	target := ""
	err = s.rpc.call("test", &target)
	s.Require().NotNil(err)
}

func (s *EthRPCTestSuite) TestWeb3Sha3() {
	response := `{"jsonrpc":"2.0", "id":1, "result": "sha3result"}`

	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "web3_sha3")
		s.paramsEqual(body, `["0x64617461"]`)

		return httpmock.NewStringResponse(200, response), nil
	})

	result, err := s.rpc.Web3Sha3([]byte("data"))
	s.Require().Nil(err)
	s.Require().Equal("sha3result", result)
}

func (s *EthRPCTestSuite) TestNetVersion() {
	response := `{"jsonrpc":"2.0", "id":1, "result": "v2b3"}`

	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "net_version")
		s.paramsEqual(body, "null")

		return httpmock.NewStringResponse(200, response), nil
	})

	v, err := s.rpc.NetVersion()
	s.Require().Nil(err)
	s.Require().Equal("v2b3", v)
}

func (s *EthRPCTestSuite) TestNetListening() {
	response := `{"jsonrpc":"2.0", "id":1, "result": true}`
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "net_listening")
		s.paramsEqual(body, "null")

		return httpmock.NewStringResponse(200, response), nil
	})

	listening, err := s.rpc.NetListening()
	s.Require().Nil(err)
	s.Require().True(listening)

	httpmock.Reset()
	response = `{"jsonrpc":"2.0", "id":1, "result": false}`
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "net_listening")
		s.paramsEqual(body, "null")

		return httpmock.NewStringResponse(200, response), nil
	})

	listening, err = s.rpc.NetListening()
	s.Require().Nil(err)
	s.Require().False(listening)
}

func (s *EthRPCTestSuite) TestNetPeerCount() {
	// Test error
	s.registerResponseError(errors.New("Error"))
	peerCount, err := s.rpc.NetPeerCount()
	s.Require().NotNil(err)
	s.Require().Equal(0, peerCount)

	// Test success
	s.registerResponse(`"0x22"`, func(body []byte) {
		s.methodEqual(body, "net_peerCount")
		s.paramsEqual(body, "null")
	})

	peerCount, err = s.rpc.NetPeerCount()
	s.Require().Nil(err)
	s.Require().Equal(34, peerCount)
}

func (s *EthRPCTestSuite) TestEthProtocolVersion() {
	s.registerResponse(`"54"`, func(body []byte) {
		s.methodEqual(body, "eth_protocolVersion")
		s.paramsEqual(body, "null")
	})

	protocolVersion, err := s.rpc.EthProtocolVersion()
	s.Require().Nil(err)
	s.Require().Equal("54", protocolVersion)
}

func (s *EthRPCTestSuite) TestEthSyncing() {
	s.registerResponseError(errors.New("Error"))
	syncing, err := s.rpc.EthSyncing()
	s.Require().NotNil(err)

	expected := &Syncing{
		IsSyncing:     false,
		CurrentBlock:  0,
		HighestBlock:  0,
		StartingBlock: 0,
	}
	s.registerResponse(`false`, func(body []byte) {
		s.methodEqual(body, "eth_syncing")
	})
	syncing, err = s.rpc.EthSyncing()

	s.Require().Nil(err)
	s.Require().Equal(expected, syncing)

	httpmock.Reset()
	s.registerResponse(`{
		"currentBlock": "0x8c3be",
		"highestBlock": "0x9bb3b",
		"startingBlock": "0x0"
	}`, func(body []byte) {})

	expected = &Syncing{
		IsSyncing:     true,
		CurrentBlock:  574398,
		HighestBlock:  637755,
		StartingBlock: 0,
	}
	syncing, err = s.rpc.EthSyncing()
	s.Require().Nil(err)
	s.Require().Equal(expected, syncing)
}

func (s *EthRPCTestSuite) TestEthCoinbase() {
	s.registerResponse(`"0x407d73d8a49eeb85d32cf465507dd71d507100c1"`, func(body []byte) {
		s.methodEqual(body, "eth_coinbase")
		s.paramsEqual(body, "null")
	})

	address, err := s.rpc.EthCoinbase()
	s.Require().Nil(err)
	s.Require().Equal("0x407d73d8a49eeb85d32cf465507dd71d507100c1", address)
}
func (s *EthRPCTestSuite) TestEthMining() {
	s.registerResponse(`true`, func(body []byte) {
		s.methodEqual(body, "eth_mining")
		s.paramsEqual(body, "null")
	})

	mining, err := s.rpc.EthMining()
	s.Require().Nil(err)
	s.Require().True(mining)

	httpmock.Reset()
	s.registerResponse(`false`, func(body []byte) {})

	mining, err = s.rpc.EthMining()
	s.Require().Nil(err)
	s.Require().False(mining)
}

func (s *EthRPCTestSuite) TestEthHashrate() {
	s.registerResponseError(errors.New("Error"))
	hashrate, err := s.rpc.EthHashrate()
	s.Require().NotNil(err)

	s.registerResponse(`"0x38a"`, func(body []byte) {
		s.methodEqual(body, "eth_hashrate")
		s.paramsEqual(body, "null")
	})

	hashrate, err = s.rpc.EthHashrate()
	s.Require().Nil(err)
	s.Require().Equal(906, hashrate)
}

func (s *EthRPCTestSuite) TestEthGasPrice() {
	s.registerResponseError(errors.New("Error"))
	gasPrice, err := s.rpc.EthGasPrice()
	s.Require().NotNil(err)

	s.registerResponse(`"0x09184e72a000"`, func(body []byte) {
		s.methodEqual(body, "eth_gasPrice")
		s.paramsEqual(body, "null")
	})

	expected, _ := big.NewInt(0).SetString("09184e72a000", 16)
	gasPrice, err = s.rpc.EthGasPrice()
	s.Require().Nil(err)
	s.Require().Equal(*expected, gasPrice)
}

func (s *EthRPCTestSuite) TestEthAccounts() {
	s.registerResponse(`["0x407d73d8a49eeb85d32cf465507dd71d507100c1"]`, func(body []byte) {
		s.methodEqual(body, "eth_accounts")
		s.paramsEqual(body, "null")
	})

	accounts, err := s.rpc.EthAccounts()
	s.Require().Nil(err)
	s.Require().Equal([]string{"0x407d73d8a49eeb85d32cf465507dd71d507100c1"}, accounts)
}

func (s *EthRPCTestSuite) TestEthBlockNumber() {
	s.registerResponseError(errors.New("Error"))
	blockBumber, err := s.rpc.EthBlockNumber()
	s.Require().NotNil(err)

	s.registerResponse(`"0x37eb38"`, func(body []byte) {
		s.methodEqual(body, "eth_blockNumber")
		s.paramsEqual(body, "null")
	})

	blockBumber, err = s.rpc.EthBlockNumber()
	s.Require().Nil(err)
	s.Require().Equal(3664696, blockBumber)
}

func (s *EthRPCTestSuite) TestEthGetBalance() {
	address := "0x407d73d8a49eeb85d32cf465507dd71d507100c1"
	s.registerResponseError(errors.New("Error"))
	balance, err := s.rpc.EthGetBalance(address, "latest")
	s.Require().NotNil(err)

	s.registerResponse(`"0x486d06b0d08d05909c4"`, func(body []byte) {
		s.methodEqual(body, "eth_getBalance")
		s.paramsEqual(body, fmt.Sprintf(`["%s", "latest"]`, address))
	})

	expected, _ := big.NewInt(0).SetString("21376347749069564217796", 10)
	balance, err = s.rpc.EthGetBalance(address, "latest")
	s.Require().Nil(err)
	s.Require().Equal(*expected, balance)
}

func (s *EthRPCTestSuite) TestEthGetStorageAt() {
	data := "0x295a70b2de5e3953354a6a8344e616ed314d7251"
	position := 33
	tag := "pending"

	s.registerResponse(`"0x00000000000000000000000000000000000000000000000000000000000004d2"`, func(body []byte) {
		s.methodEqual(body, "eth_getStorageAt")
		s.paramsEqual(body, fmt.Sprintf(`["%s", "0x21", "pending"]`, data))
	})

	result, err := s.rpc.EthGetStorageAt(data, position, tag)
	s.Require().Nil(err)
	s.Require().Equal("0x00000000000000000000000000000000000000000000000000000000000004d2", result)
}

func (s *EthRPCTestSuite) TestEthGetTransactionCount() {
	address := "0x407d73d8a49eeb85d32cf465507dd71d507100c1"
	s.registerResponseError(errors.New("Error"))
	count, err := s.rpc.EthGetTransactionCount(address, "latest")
	s.Require().NotNil(err)

	s.registerResponse(`"0x10"`, func(body []byte) {
		s.methodEqual(body, "eth_getTransactionCount")
		s.paramsEqual(body, fmt.Sprintf(`["%s", "latest"]`, address))
	})

	count, err = s.rpc.EthGetTransactionCount(address, "latest")
	s.Require().Nil(err)
	s.Require().Equal(16, count)
}

func (s *EthRPCTestSuite) TestEthGetBlockTransactionCountByHash() {
	hash := "0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238"
	s.registerResponseError(errors.New("Error"))
	count, err := s.rpc.EthGetBlockTransactionCountByHash(hash)
	s.Require().NotNil(err)

	s.registerResponse(`"0xb"`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockTransactionCountByHash")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, hash))
	})

	count, err = s.rpc.EthGetBlockTransactionCountByHash(hash)
	s.Require().Nil(err)
	s.Require().Equal(11, count)
}

func (s *EthRPCTestSuite) TestEthGetBlockTransactionCountByNumber() {
	number := 2384732
	s.registerResponseError(errors.New("Error"))
	count, err := s.rpc.EthGetBlockTransactionCountByNumber(number)
	s.Require().NotNil(err)

	s.registerResponse(`"0xe8"`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockTransactionCountByNumber")
		s.paramsEqual(body, `["0x24635c"]`)
	})

	count, err = s.rpc.EthGetBlockTransactionCountByNumber(number)
	s.Require().Nil(err)
	s.Require().Equal(232, count)
}

func (s *EthRPCTestSuite) TestEthGetUncleCountByBlockHash() {
	hash := "0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238"
	s.registerResponseError(errors.New("Error"))
	count, err := s.rpc.EthGetUncleCountByBlockHash(hash)
	s.Require().NotNil(err)

	s.registerResponse(`"0xa"`, func(body []byte) {
		s.methodEqual(body, "eth_getUncleCountByBlockHash")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, hash))
	})

	count, err = s.rpc.EthGetUncleCountByBlockHash(hash)
	s.Require().Nil(err)
	s.Require().Equal(10, count)
}

func (s *EthRPCTestSuite) TestEthGetUncleCountByBlockNumber() {
	number := 3987434
	s.registerResponseError(errors.New("Error"))
	count, err := s.rpc.EthGetUncleCountByBlockNumber(number)
	s.Require().NotNil(err)

	s.registerResponse(`"0x386"`, func(body []byte) {
		s.methodEqual(body, "eth_getUncleCountByBlockNumber")
		s.paramsEqual(body, `["0x3cd7ea"]`)
	})

	count, err = s.rpc.EthGetUncleCountByBlockNumber(number)
	s.Require().Nil(err)
	s.Require().Equal(902, count)
}

func (s *EthRPCTestSuite) TestEthGetCode() {
	address := "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b"
	result := "0x600160008035811a818181146012578301005b601b6001356025565b8060005260206000f25b600060078202905091905056"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_getCode")
		s.paramsEqual(body, fmt.Sprintf(`["%s", "latest"]`, address))
	})

	code, err := s.rpc.EthGetCode(address, "latest")
	s.Require().Nil(err)
	s.Require().Equal(result, code)
}

func (s *EthRPCTestSuite) TestEthSign() {
	address := "0x9b2055d370f73ec7d8a03e965129118dc8f5bf83"
	data := "0xdeadbeaf"
	result := "0xa3f20717a250c2b0b729b7e5becbff67fdaef7e0699da4de7ca5895b02a170a12d887fd3b17bfdce3481f10bea41f45ba9f709d39ce8325427b57afcfc994cee1b"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_sign")
		s.paramsEqual(body, fmt.Sprintf(`["%s", "%s"]`, address, data))
	})

	signed, err := s.rpc.EthSign(address, data)
	s.Require().Nil(err)
	s.Require().Equal(result, signed)
}

func (s *EthRPCTestSuite) TestSendTransaction() {
	t := T{
		From:     "0x3cc1a3c082944b9dba70e490e481dd56",
		To:       "0x1bf21cb1dc384d019a885a06973f7308",
		Gas:      24900,
		GasPrice: big.NewInt(5000000000),
		Value:    big.NewInt(1000000000000000000), // 1 ETH
		Data:     "some data",
		Nonce:    98384,
	}

	result := "0xea1115eb5"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_sendTransaction")
		s.paramsEqual(body, `[{
			"from": "0x3cc1a3c082944b9dba70e490e481dd56",
			"to": "0x1bf21cb1dc384d019a885a06973f7308",
			"gas": "0x6144",
			"gasPrice": "0x12a05f200",
			"value": "0xde0b6b3a7640000",
			"data": "some data",
			"nonce": "0x18050"
		}]`)
	})

	txid, err := s.rpc.EthSendTransaction(t)
	s.Require().Nil(err)
	s.Require().Equal(result, txid)

	t = T{}
	httpmock.Reset()
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_sendTransaction")
		s.paramsEqual(body, `[{
			"from": ""
		}]`)

	})

	txid, err = s.rpc.EthSendTransaction(t)
	s.Require().Nil(err)
	s.Require().Equal(result, txid)
}

func (s *EthRPCTestSuite) TestEthSendRawTransaction() {
	data := "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"
	result := "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_sendRawTransaction")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, data))
	})

	txid, err := s.rpc.EthSendRawTransaction(data)
	s.Require().Nil(err)
	s.Require().Equal(result, txid)
}

func (s *EthRPCTestSuite) TestEthGetCompilers() {
	s.registerResponse(`["solidity", "some comp"]`, func(body []byte) {
		s.methodEqual(body, "eth_getCompilers")
		s.paramsEqual(body, "null")
	})

	compilers, err := s.rpc.EthGetCompilers()
	s.Require().Nil(err)
	s.Require().Equal([]string{"solidity", "some comp"}, compilers)

}

func (s *EthRPCTestSuite) TestGetBlock() {
	s.registerResponseError(errors.New("Error"))
	block, err := s.rpc.getBlock("eth_getBlockByHash", true)
	s.Require().NotNil(err)

	// Test with transactions
	result := ` {
        "difficulty": "0x81299d4dbde29",
        "extraData": "0x706f6f6c2e65746866616e732e6f726720284d4e323729",
        "gasLimit": "0x667900",
        "gasUsed": "0x639fa0",
        "hash": "0x2bdda43f649c564642101fc990f569dd855e60f88bf83e931f509a92c62700f9",
        "logsBloom": "0x111",
        "miner": "0x1e9939daaad6924ad004c2560e90804164900341",
        "mixHash": "0xa6b69fa82eaea8674236170a2d8ea41d80c176315a579138b718f3bcaa4c39ab",
        "nonce": "0xefd7ef000d0b78b8",
        "number": "0x4055d5",
        "parentHash": "0x913f938dcb4ff83b2b6b42a0cf6517d438a3ce95174e9342c780fd20c84dfd03",
        "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
        "size": "0x2fc6",
        "stateRoot": "0xab9287d3b8864338892d1d572198933979e39bfcfbde569ea52be15a9691b4c1",
        "timestamp": "0x59a556bd",
        "totalDifficulty": "0x2b5f79e86aaf701c81",
        "transactions": [
			{
				"blockHash": "0x2bdda43f649c564642101fc990f569dd855e60f88bf83e931f509a92c62700f9",
				"blockNumber": "0x4055d5",
				"from": "0xa95350d70b18fa29f6b5eb8d627ceeeee499340d",
				"gas": "0x5208",
				"gasPrice": "0x6edf2a079e",
				"hash": "0xf519ca0e9ceeb0405dfeb95544179f557e3221213f07e33709af7ced60ab61b9",
				"input": "0x",
				"nonce": "0x289b",
				"to": "0xb595f3390fcec074237c8264b908fc73d4aedc93",
				"transactionIndex": "0x0",
				"value": "0xdbd2fc137a30000"
			},
			{
				"blockHash": "0x2bdda43f649c564642101fc990f569dd855e60f88bf83e931f509a92c62700f9",
				"blockNumber": "0x4055d5",
				"from": "0x0f1b76410215ed963ea2c3d3eaddd4a56350b422",
				"gas": "0x3d090",
				"gasPrice": "0x1176592e00",
				"hash": "0xa72743a3608e2ae7b3d1cc1f0e3ceed9a1c78d803eba5f28d5d6908adfaa211c",
				"input": "0x278b8c0e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004551ce090d138000000000000000000000000000006bea43baa3f7a6f765f14f10a1a1b08334ef4500000000000000000000000000000000000000000000003627e8f712373c000000000000000000000000000000000000000000000000000000000000004059b200000000000000000000000000000000000000000000000000000000418e8e7d000000000000000000000000000000000000000000000000000000000000001b64b1fee882b69969c9395a095e45e4b0abb3b19806ba040a6765194f966ae64e24a3d44a837de95e014b6b0f7eea075e30cca0414a18c0a27a7f349271689f3d",
				"nonce": "0x1c2",
				"to": "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819",
				"transactionIndex": "0x1",
				"value": "0x0"
			}
		],
        "transactionsRoot": "0x97849642410701c38f904912238eb78d3aa854e72c5ae39394c7217f4f9474bc",
        "uncles": ["0xf14cdb8a75de31dcf3da7a3a52c1fffcbaa3d56de9f50f86767fa411c10f4397"]
	}`
	hash := "0x2bdda43f649c564642101fc990f569dd855e60f88bf83e931f509a92c62700f9"
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "eth_getBlockByHash")
	})

	block, err = s.rpc.getBlock("eth_getBlockByHash", true)
	s.Require().Nil(err)
	s.Require().NotNil(block)
	s.Require().Equal(hash, block.Hash)
	s.Require().Equal(4216277, block.Number)
	s.Require().Equal("0x913f938dcb4ff83b2b6b42a0cf6517d438a3ce95174e9342c780fd20c84dfd03", block.ParentHash)
	s.Require().Equal("0xefd7ef000d0b78b8", block.Nonce)
	s.Require().Equal("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", block.Sha3Uncles)
	s.Require().Equal("0x111", block.LogsBloom)
	s.Require().Equal("0x97849642410701c38f904912238eb78d3aa854e72c5ae39394c7217f4f9474bc", block.TransactionsRoot)
	s.Require().Equal("0xab9287d3b8864338892d1d572198933979e39bfcfbde569ea52be15a9691b4c1", block.StateRoot)
	s.Require().Equal("0x1e9939daaad6924ad004c2560e90804164900341", block.Miner)
	s.Require().Equal(newBigInt("2272251724160553"), block.Difficulty)
	s.Require().Equal(newBigInt("800089780620203400321"), block.TotalDifficulty)
	s.Require().Equal("0x706f6f6c2e65746866616e732e6f726720284d4e323729", block.ExtraData)
	s.Require().Equal(12230, block.Size)
	s.Require().Equal(6715648, block.GasLimit)
	s.Require().Equal(6528928, block.GasUsed)
	s.Require().Equal(1504007869, block.Timestamp)
	s.Require().Equal([]string{"0xf14cdb8a75de31dcf3da7a3a52c1fffcbaa3d56de9f50f86767fa411c10f4397"}, block.Uncles)
	s.Require().Equal(2, len(block.Transactions))

	s.Require().Equal(Transaction{
		Hash:             "0xf519ca0e9ceeb0405dfeb95544179f557e3221213f07e33709af7ced60ab61b9",
		Nonce:            10395,
		BlockHash:        block.Hash,
		BlockNumber:      &block.Number,
		TransactionIndex: ptrInt(0),
		From:             "0xa95350d70b18fa29f6b5eb8d627ceeeee499340d",
		To:               "0xb595f3390fcec074237c8264b908fc73d4aedc93",
		Value:            newBigInt("990000000000000000"),
		Gas:              21000,
		GasPrice:         newBigInt("476190476190"),
		Input:            "0x",
	}, block.Transactions[0])

	s.Require().Equal(Transaction{
		Hash:             "0xa72743a3608e2ae7b3d1cc1f0e3ceed9a1c78d803eba5f28d5d6908adfaa211c",
		Nonce:            450,
		BlockHash:        block.Hash,
		BlockNumber:      &block.Number,
		TransactionIndex: ptrInt(1),
		From:             "0x0f1b76410215ed963ea2c3d3eaddd4a56350b422",
		To:               "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819",
		Value:            newBigInt("0"),
		Gas:              250000,
		GasPrice:         newBigInt("75000000000"),
		Input:            "0x278b8c0e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004551ce090d138000000000000000000000000000006bea43baa3f7a6f765f14f10a1a1b08334ef4500000000000000000000000000000000000000000000003627e8f712373c000000000000000000000000000000000000000000000000000000000000004059b200000000000000000000000000000000000000000000000000000000418e8e7d000000000000000000000000000000000000000000000000000000000000001b64b1fee882b69969c9395a095e45e4b0abb3b19806ba040a6765194f966ae64e24a3d44a837de95e014b6b0f7eea075e30cca0414a18c0a27a7f349271689f3d",
	}, block.Transactions[1])

	httpmock.Reset()
	// Test without transactions
	result = `{
		"difficulty": "0x7feab8ef4d978",
		"extraData": "0xd58301050b8650617269747986312e31352e31826c69",
		"gasLimit": "0x665f6b",
		"gasUsed": "0x1d71b",
		"hash": "0x23be1464d0e805fe3cec49039a9cf7fae7c09d2efacbed2abb10ef7ddae960ab",
		"logsBloom": "0x222",
		"miner": "0x6a7a43be33ba930fe58f34e07d0ad6ba7adb9b1f",
		"mixHash": "0xa8f339af405f7f3a7b7c163f8889f44343abfbbeda13c41e06923de349ea6483",
		"nonce": "0x19a48ee424b5088f",
		"number": "0x4105f3",
		"parentHash": "0xbc3e37984a619008d75e7f73865247fb420ae5ed2c921599d099ab5f20519396",
		"receiptsRoot": "0xa1384524d42ff86fdf4e44eeea853aba4e772a52240037cbfddc22782bad017e",
		"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
		"size": "0x490",
		"stateRoot": "0xbe7e86ee05a5d49ba64b3d9f3f0129bab90308032e42307a1a2ef5c8971c5f5c",
		"timestamp": "0x59b62713",
		"totalDifficulty": "0x30e3d47fb9d7a43f7c",
		"transactions": [
			"0x160e19780a24f3d78492c7ac7228e0220d4b96878fec19daf182e1d8c4b3d94e"
		],
		"transactionsRoot": "0x1bcd58c2420d63c5e8ed3182afd33c01737be38a4a8c10a81dfb70b692e8f286",
		"uncles": []
	}`

	hash = "0x23be1464d0e805fe3cec49039a9cf7fae7c09d2efacbed2abb10ef7ddae960ab"
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "eth_getBlockByHash")
	})

	block, err = s.rpc.getBlock("eth_getBlockByHash", false)
	s.Require().Nil(err)
	s.Require().NotNil(block)
	s.Require().Equal(hash, block.Hash)
	s.Require().Equal(4261363, block.Number)
	s.Require().Equal("0xbc3e37984a619008d75e7f73865247fb420ae5ed2c921599d099ab5f20519396", block.ParentHash)
	s.Require().Equal("0x19a48ee424b5088f", block.Nonce)
	s.Require().Equal("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", block.Sha3Uncles)
	s.Require().Equal("0x222", block.LogsBloom)
	s.Require().Equal("0x1bcd58c2420d63c5e8ed3182afd33c01737be38a4a8c10a81dfb70b692e8f286", block.TransactionsRoot)
	s.Require().Equal("0xbe7e86ee05a5d49ba64b3d9f3f0129bab90308032e42307a1a2ef5c8971c5f5c", block.StateRoot)
	s.Require().Equal("0x6a7a43be33ba930fe58f34e07d0ad6ba7adb9b1f", block.Miner)
	s.Require().Equal(newBigInt("2250337628248440"), block.Difficulty)
	s.Require().Equal(newBigInt("901860602515894321020"), block.TotalDifficulty)
	s.Require().Equal("0xd58301050b8650617269747986312e31352e31826c69", block.ExtraData)
	s.Require().Equal(1168, block.Size)
	s.Require().Equal(6709099, block.GasLimit)
	s.Require().Equal(120603, block.GasUsed)
	s.Require().Equal(1505109779, block.Timestamp)
	s.Require().Equal([]string{}, block.Uncles)
	s.Require().Equal(1, len(block.Transactions))
	s.Require().Equal(Transaction{
		Hash:             "0x160e19780a24f3d78492c7ac7228e0220d4b96878fec19daf182e1d8c4b3d94e",
		Nonce:            0,
		BlockHash:        "",
		BlockNumber:      nil,
		TransactionIndex: nil,
		From:             "",
		To:               "",
		Value:            big.Int{},
		Gas:              0,
		GasPrice:         big.Int{},
		Input:            "",
	}, block.Transactions[0])

	s.registerResponse("null", func(body []byte) {})

	block, err = s.rpc.getBlock("eth_getBlockByHash", false)
	s.Require().Nil(block)
	s.Require().Nil(err)
}

func (s *EthRPCTestSuite) TestEthGetBlockByHash() {
	// Test with transactions
	hash := "0x111"
	s.registerResponse(`{}`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockByHash")
		s.paramsEqual(body, `["0x111", true]`)
	})

	_, err := s.rpc.EthGetBlockByHash(hash, true)
	s.Require().Nil(err)

	httpmock.Reset()

	// Test without transactions
	hash = "0x222"
	s.registerResponse(`{}`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockByHash")
		s.paramsEqual(body, `["0x222", false]`)
	})

	_, err = s.rpc.EthGetBlockByHash(hash, false)
	s.Require().Nil(err)
}

func (s *EthRPCTestSuite) TestEthGetBlockByNumber() {
	// Test with transactions
	number := 3274863
	s.registerResponse(`{}`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockByNumber")
		s.paramsEqual(body, `["0x31f86f", true]`)
	})

	_, err := s.rpc.EthGetBlockByNumber(number, true)
	s.Require().Nil(err)

	httpmock.Reset()

	// Test without transactions
	number = 14322
	s.registerResponse(`{}`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockByNumber")
		s.paramsEqual(body, `["0x37f2", false]`)
	})

	_, err = s.rpc.EthGetBlockByNumber(number, false)
	s.Require().Nil(err)
}

func (s *EthRPCTestSuite) TestEthCall() {
	s.registerResponse(`"0x11"`, func(body []byte) {
		s.methodEqual(body, "eth_call")
		s.paramsEqual(body, `[{"from":"0x111","to":"0x222"}, "ttt"]`)
	})

	result, err := s.rpc.EthCall(T{
		From: "0x111",
		To:   "0x222",
	}, "ttt")
	s.Require().Nil(err)
	s.Require().Equal("0x11", result)
}

func (s *EthRPCTestSuite) TestEthEstimateGas() {
	s.registerResponseError(errors.New("error"))
	result, err := s.rpc.EthEstimateGas(T{
		From: "0x111",
		To:   "0x222",
	})
	s.Require().NotNil(err)

	s.registerResponse(`"0x5022"`, func(body []byte) {
		s.methodEqual(body, "eth_estimateGas")
		s.paramsEqual(body, `[{"from":"0x111","to":"0x222"}]`)
	})
	result, err = s.rpc.EthEstimateGas(T{
		From: "0x111",
		To:   "0x222",
	})
	s.Require().Nil(err)
	s.Require().Equal(20514, result)
}

func (s *EthRPCTestSuite) TestEthGetTransactionReceipt() {
	hash := "0x9c17afa5336d3cfd47e2e795520959b92e627e123e538fd4d5d7ece9025a8dce"
	s.registerResponseError(errors.New("error"))
	receipt, err := s.rpc.EthGetTransactionReceipt(hash)
	s.Require().NotNil(err)

	result := `{
        "blockHash": "0x11537af16aec572bb72d6d52e2c801dbfc10f42ab6ea849fd8e31b57d7099eea",
        "blockNumber": "0x3919d3",
        "contractAddress": null,
        "cumulativeGasUsed": "0x1677f1",
        "gasUsed": "0x10148",
        "logs": [{
            "address": "0xcd111aa492a9c77a367c36e6d6af8e6f212e0c8e",
            "topics": ["0x78e4fc71ff7e525b3b4660a76336a2046232fd9bba9c65abb22fa3d07d6e7066"],
            "data": "0x9da86521f54f8e4747f86593145f7ec22f2ab4c8e32288c378ed503f253b6426",
            "blockNumber": "0x3919d3",
            "transactionHash": "0x9c17afa5336d3cfd47e2e795520959b92e627e123e538fd4d5d7ece9025a8dce",
            "transactionIndex": "0x13",
            "blockHash": "0x11537af16aec572bb72d6d52e2c801dbfc10f42ab6ea849fd8e31b57d7099eea",
            "logIndex": "0xc",
            "removed": false
        }],
        "logsBloom": "0x001",
        "root": "0x55b68780caee96e686eb398371bb679574d4b995614ae94243da4886059a47ee",
        "transactionHash": "0x9c17afa5336d3cfd47e2e795520959b92e627e123e538fd4d5d7ece9025a8dce",
		"transactionIndex": "0x13",
		"status": "0x1"
	}`
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "eth_getTransactionReceipt")
		s.paramsEqual(body, `["0x9c17afa5336d3cfd47e2e795520959b92e627e123e538fd4d5d7ece9025a8dce"]`)
	})

	receipt, err = s.rpc.EthGetTransactionReceipt(hash)
	s.Require().Nil(err)
	s.Require().NotNil(receipt)
	s.Require().Equal(hash, receipt.TransactionHash)
	s.Require().Equal(19, receipt.TransactionIndex)
	s.Require().Equal("0x11537af16aec572bb72d6d52e2c801dbfc10f42ab6ea849fd8e31b57d7099eea", receipt.BlockHash)
	s.Require().Equal(3742163, receipt.BlockNumber)
	s.Require().Equal(1472497, receipt.CumulativeGasUsed)
	s.Require().Equal(65864, receipt.GasUsed)
	s.Require().Equal("", receipt.ContractAddress)
	s.Require().Equal("0x001", receipt.LogsBloom)
	s.Require().Equal("0x55b68780caee96e686eb398371bb679574d4b995614ae94243da4886059a47ee", receipt.Root)
	s.Require().Equal("0x1", receipt.Status)
	s.Require().Equal(1, len(receipt.Logs))
	s.Require().Equal(Log{
		Removed:          false,
		LogIndex:         12,
		TransactionIndex: receipt.TransactionIndex,
		TransactionHash:  receipt.TransactionHash,
		BlockNumber:      receipt.BlockNumber,
		BlockHash:        receipt.BlockHash,
		Address:          "0xcd111aa492a9c77a367c36e6d6af8e6f212e0c8e",
		Data:             "0x9da86521f54f8e4747f86593145f7ec22f2ab4c8e32288c378ed503f253b6426",
		Topics:           []string{"0x78e4fc71ff7e525b3b4660a76336a2046232fd9bba9c65abb22fa3d07d6e7066"},
	}, receipt.Logs[0])
}

func (s *EthRPCTestSuite) TestGetTransaction() {
	result := `{
        "blockHash": "0x8b0404b2e5173e7abdbfc98f521d50808486ccaff3cd0a6344e0bb6c7aa8cef0",
        "blockNumber": "0x4109ed",
        "from": "0xe3a7ca9d2306b0dc900ea618648bed9ec6cb1106",
        "gas": "0x3d090",
        "gasPrice": "0xee6b2800",
        "hash": "0x3068bb24a6c65a80eb350b89b2ef2f4d0605f59e5d07fd3467eb76511c4408e7",
        "input": "0x522",
        "nonce": "0xa8",
        "to": "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819",
        "transactionIndex": "0x98",
        "value": "0x9184e72a000"
    }`
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "ggg")
	})

	transaction, err := s.rpc.getTransaction("ggg")
	s.Require().Nil(err)
	s.Require().NotNil(transaction)
	s.Require().Equal("0x3068bb24a6c65a80eb350b89b2ef2f4d0605f59e5d07fd3467eb76511c4408e7", transaction.Hash)
	s.Require().Equal(168, transaction.Nonce)
	s.Require().Equal("0x8b0404b2e5173e7abdbfc98f521d50808486ccaff3cd0a6344e0bb6c7aa8cef0", transaction.BlockHash)
	s.Require().Equal(4262381, *transaction.BlockNumber)
	s.Require().Equal(152, *transaction.TransactionIndex)
	s.Require().Equal("0xe3a7ca9d2306b0dc900ea618648bed9ec6cb1106", transaction.From)
	s.Require().Equal("0x8d12a197cb00d4747a1fe03395095ce2a5cc6819", transaction.To)
	s.Require().Equal(newBigInt("10000000000000"), transaction.Value)
	s.Require().Equal(250000, transaction.Gas)
	s.Require().Equal(newBigInt("4000000000"), transaction.GasPrice)
	s.Require().Equal("0x522", transaction.Input)
}

func (s *EthRPCTestSuite) TestEthGetTransactionByHash() {
	s.registerResponse(`{}`, func(body []byte) {
		s.methodEqual(body, "eth_getTransactionByHash")
		s.paramsEqual(body, `["0x123"]`)
	})

	t, err := s.rpc.EthGetTransactionByHash("0x123")
	s.Require().Nil(err)
	s.Require().NotNil(t)
}

func (s *EthRPCTestSuite) TestEthGetTransactionByBlockHashAndIndex() {
	s.registerResponse(`{}`, func(body []byte) {
		s.methodEqual(body, "eth_getTransactionByBlockHashAndIndex")
		s.paramsEqual(body, `["0x623", "0x12"]`)
	})

	t, err := s.rpc.EthGetTransactionByBlockHashAndIndex("0x623", 18)
	s.Require().Nil(err)
	s.Require().NotNil(t)
}

func (s *EthRPCTestSuite) TestEthGetTransactionByBlockNumberAndIndex() {
	s.registerResponse(`{}`, func(body []byte) {
		s.methodEqual(body, "eth_getTransactionByBlockNumberAndIndex")
		s.paramsEqual(body, `["0x1f537da", "0xa"]`)
	})

	t, err := s.rpc.EthGetTransactionByBlockNumberAndIndex(32847834, 10)
	s.Require().Nil(err)
	s.Require().NotNil(t)
}

func (s *EthRPCTestSuite) TestEthNewFilterWithAddress() {
	address := []string{"0xb2b2eeeee341e560da3d439ef5e5309d78a22a66"}
	filterData := FilterParams{Address: address}
	result := "0x6996a3a4788d4f2067108d1f536d4330"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_newFilter")
		s.paramsEqual(body, fmt.Sprintf(`[{"address": ["%s"]}]`, address[0]))
	})

	filterID, err := s.rpc.EthNewFilter(filterData)
	s.Require().Nil(err)
	s.Require().Equal(result, filterID)
}

func (s *EthRPCTestSuite) TestEthNewFilterWithTopics() {
	topics := [][]string{
		{
			"0xb2b2eeeee341e560da3d439ef5e5309d78a22a66",
			"0xb2b2fffff341e560da3d439ef5e5309d78a22a66",
		},
	}
	filterData := FilterParams{Topics: topics}
	result := "0x6996a3a4788d4f2067108d1f536d4330"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_newFilter")
		s.paramsEqual(body, fmt.Sprintf(`[{"topics": [["%s", "%s"]]}]`, topics[0][0], topics[0][1]))
	})

	filterID, err := s.rpc.EthNewFilter(filterData)
	s.Require().Nil(err)
	s.Require().Equal(result, filterID)
}

func (s *EthRPCTestSuite) TestEthNewFilterWithAddressAndTopics() {
	topics := [][]string{
		{"0xb2b2eeeee341e560da3d439ef5e5309d78a22a66"},
		{"0xb2b2fffff341e560da3d439ef5e5309d78a22a66"},
	}
	address := []string{"0xb2b2eeeee341e560da3d439ef5e5309d78a22a66"}
	filterData := FilterParams{Address: address, Topics: topics}
	result := "0x6996a3a4788d4f2067108d1f536d4330"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_newFilter")
		s.paramsEqual(body, fmt.Sprintf(`[{"address": ["%s"], "topics": [["%s"], ["%s"]]}]`, address[0], topics[0][0], topics[1][0]))
	})

	filterID, err := s.rpc.EthNewFilter(filterData)
	s.Require().Nil(err)
	s.Require().Equal(result, filterID)
}

func (s *EthRPCTestSuite) TestEthNewBlockFilter() {
	result := "0x6996a3a4788d4f2067108d1f536d4330"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_newBlockFilter")
	})

	filterID, err := s.rpc.EthNewBlockFilter()
	s.Require().Nil(err)
	s.Require().Equal(result, filterID)
}

func (s *EthRPCTestSuite) TestEthNewPendingTransactionFilter() {
	result := "0x153"
	s.registerResponse(fmt.Sprintf(`"%s"`, result), func(body []byte) {
		s.methodEqual(body, "eth_newPendingTransactionFilter")
	})

	filterID, err := s.rpc.EthNewPendingTransactionFilter()
	s.Require().Nil(err)
	s.Require().Equal(result, filterID)
}

func (s *EthRPCTestSuite) TestEthGetFilterChanges() {
	filterID := "0x6996a3a4788d4f2067108d1f536d4330"
	result := `[{
		"address":"0xaca0cc3a6bf9552f2866ccc67801d4e6aa6a70f2",
		"blockHash":"0x9d9838090bb7f6194f62acea788688435b79cc44c62dcf1479abd9f2c72a7d5c",
		"blockNumber":1,
		"data":"0x000000000000000000000000000000000000000000000000000000112c905320",
		"logIndex":0,
		"removed":false,
		"topics":["0x581d416ae9dff30c9305c2b35cb09ed5991897ab97804db29ccf92678e953160"]
	}]`
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "eth_getFilterChanges")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, filterID))
	})

	logs, err := s.rpc.EthGetFilterChanges(filterID)
	s.Require().Nil(err)
	s.Require().Equal([]Log{
		{
			Address:     "0xaca0cc3a6bf9552f2866ccc67801d4e6aa6a70f2",
			BlockHash:   "0x9d9838090bb7f6194f62acea788688435b79cc44c62dcf1479abd9f2c72a7d5c",
			BlockNumber: 1,
			Data:        "0x000000000000000000000000000000000000000000000000000000112c905320",
			LogIndex:    0,
			Removed:     false,
			Topics:      []string{"0x581d416ae9dff30c9305c2b35cb09ed5991897ab97804db29ccf92678e953160"},
		},
	}, logs)
}

func (s *EthRPCTestSuite) TestEthGetFilterLogs() {
	filterID := "0x6996a3a4788d4f2067108d1f536d4330"
	result := `[{
		"address": "0xaca0cc3a6bf9552f2866ccc67801d4e6aa6a70f2",
		"blockHash": "0x9d9838090bb7f6194f62acea788688435b79cc44c62dcf1479abd9f2c72a7d5c",
		"blockNumber": 1,
		"data": "0x000000000000000000000000000000000000000000000000000000112c905320",
		"logIndex": 0,
		"removed": false,
		"topics": ["0x581d416ae9dff30c9305c2b35cb09ed5991897ab97804db29ccf92678e953160"]
	}]`
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "eth_getFilterLogs")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, filterID))
	})

	logs, err := s.rpc.EthGetFilterLogs(filterID)
	s.Require().Nil(err)
	s.Require().Equal([]Log{
		{
			Address:     "0xaca0cc3a6bf9552f2866ccc67801d4e6aa6a70f2",
			BlockHash:   "0x9d9838090bb7f6194f62acea788688435b79cc44c62dcf1479abd9f2c72a7d5c",
			BlockNumber: 1,
			Data:        "0x000000000000000000000000000000000000000000000000000000112c905320",
			LogIndex:    0,
			Removed:     false,
			Topics:      []string{"0x581d416ae9dff30c9305c2b35cb09ed5991897ab97804db29ccf92678e953160"},
		},
	}, logs)
}

func (s *EthRPCTestSuite) TestEthGetLogs() {
	params := FilterParams{
		FromBlock: "0x1",
		ToBlock:   "0x10",
		Address:   []string{"0x8888f1f195afa192cfee860698584c030f4c9db1"},
		Topics: [][]string{
			{"0x111"},
			nil,
		},
	}
	result := `[{
		"address": "0xaca0cc3a6bf9552f2866ccc67801d4e6aa6a70f2",
		"blockHash": "0x9d9838090bb7f6194f62acea788688435b79cc44c62dcf1479abd9f2c72a7d5c",
		"blockNumber": 1,
		"data": "0x000000000000000000000000000000000000000000000000000000112c905320",
		"logIndex": 0,
		"removed": false,
		"topics": ["0x581d416ae9dff30c9305c2b35cb09ed5991897ab97804db29ccf92678e953160"]
	}]`
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "eth_getLogs")
		s.paramsEqual(body, fmt.Sprintf(`[{
			"fromBlock": "0x1",
			"toBlock": "0x10",
			"address": ["0x8888f1f195afa192cfee860698584c030f4c9db1"],
			"topics": [["0x111"], null]
		}]`))
	})

	logs, err := s.rpc.EthGetLogs(params)
	s.Require().Nil(err)
	s.Require().Equal([]Log{
		{
			Address:     "0xaca0cc3a6bf9552f2866ccc67801d4e6aa6a70f2",
			BlockHash:   "0x9d9838090bb7f6194f62acea788688435b79cc44c62dcf1479abd9f2c72a7d5c",
			BlockNumber: 1,
			Data:        "0x000000000000000000000000000000000000000000000000000000112c905320",
			LogIndex:    0,
			Removed:     false,
			Topics:      []string{"0x581d416ae9dff30c9305c2b35cb09ed5991897ab97804db29ccf92678e953160"},
		},
	}, logs)
}

func (s *EthRPCTestSuite) TestEthUninstallFilter() {
	filterID := "0x6996a3a4788d4f2067108d1f536d4330"
	result := "true"
	s.registerResponse(result, func(body []byte) {
		s.methodEqual(body, "eth_uninstallFilter")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, filterID))
	})

	uninstall, err := s.rpc.EthUninstallFilter(filterID)
	s.Require().Nil(err)
	boolRes, _ := strconv.ParseBool(result)
	s.Require().Equal(boolRes, uninstall)
}

func TestEthRPCTestSuite(t *testing.T) {
	suite.Run(t, new(EthRPCTestSuite))
}

func TestEthError(t *testing.T) {
	var err error
	err = EthError{-32555, "Messg"}
	require.Equal(t, "Error -32555 (Messg)", err.Error())

	err = EthError{32847, "Kuku"}
	require.Equal(t, "Error 32847 (Kuku)", err.Error())
}

func TestEth1(t *testing.T) {
	client := NewEthRPC("")
	require.Equal(t, int64(1000000000000000000), Eth1().Int64())
	require.Equal(t, int64(1000000000000000000), client.Eth1().Int64())
}

func ptrInt(i int) *int {
	return &i
}

func newBigInt(s string) big.Int {
	i, _ := new(big.Int).SetString(s, 10)
	return *i
}
