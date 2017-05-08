package ethrpc

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type EthRPCTestSuite struct {
	suite.Suite
	rpc *EthRPC
}

func (s *EthRPCTestSuite) registerResponse(result string, callback func([]byte)) {
	response := fmt.Sprintf(`{"jsonrpc":"2.0", "id":1, "result": %s}`, result)
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		callback(s.getBody(request))
		return httpmock.NewStringResponse(200, response), nil
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
	s.rpc = NewEthRPC("http://127.0.0.1:8545")
	// s.rpc.Debug = true

	httpmock.Activate()
}

func (s *EthRPCTestSuite) TearDownSuite() {
	httpmock.Deactivate()
}

func (s *EthRPCTestSuite) TearDownTest() {
	httpmock.Reset()
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
	s.registerResponse(`"0x22"`, func(body []byte) {
		s.methodEqual(body, "net_peerCount")
		s.paramsEqual(body, "null")
	})

	peerCount, err := s.rpc.NetPeerCount()
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
	expected := &Syncing{
		IsSyncing:     false,
		CurrentBlock:  0,
		HighestBlock:  0,
		StartingBlock: 0,
	}
	s.registerResponse(`false`, func(body []byte) {
		s.methodEqual(body, "eth_syncing")
	})
	syncing, err := s.rpc.EthSyncing()

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
	s.registerResponse(`"0x38a"`, func(body []byte) {
		s.methodEqual(body, "eth_hashrate")
		s.paramsEqual(body, "null")
	})

	hashrate, err := s.rpc.EthHashrate()
	s.Require().Nil(err)
	s.Require().Equal(906, hashrate)
}

func (s *EthRPCTestSuite) TestEthGasPrice() {
	s.registerResponse(`"0x09184e72a000"`, func(body []byte) {
		s.methodEqual(body, "eth_gasPrice")
		s.paramsEqual(body, "null")
	})

	expected, _ := big.NewInt(0).SetString("09184e72a000", 16)
	gasPrice, err := s.rpc.EthGasPrice()
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
	s.registerResponse(`"0x37eb38"`, func(body []byte) {
		s.methodEqual(body, "eth_blockNumber")
		s.paramsEqual(body, "null")
	})

	blockBumber, err := s.rpc.EthBlockNumber()
	s.Require().Nil(err)
	s.Require().Equal(3664696, blockBumber)
}

func (s *EthRPCTestSuite) TestEthGetBalance() {
	address := "0x407d73d8a49eeb85d32cf465507dd71d507100c1"
	s.registerResponse(`"0x486d06b0d08d05909c4"`, func(body []byte) {
		s.methodEqual(body, "eth_getBalance")
		s.paramsEqual(body, fmt.Sprintf(`["%s", "latest"]`, address))
	})

	expected, _ := big.NewInt(0).SetString("21376347749069564217796", 10)
	balance, err := s.rpc.EthGetBalance(address, "latest")
	s.Require().Nil(err)
	s.Require().Equal(*expected, balance)
}

func (s *EthRPCTestSuite) TestEthGetTransactionCount() {
	address := "0x407d73d8a49eeb85d32cf465507dd71d507100c1"
	s.registerResponse(`"0x10"`, func(body []byte) {
		s.methodEqual(body, "eth_getTransactionCount")
		s.paramsEqual(body, fmt.Sprintf(`["%s", "latest"]`, address))
	})

	count, err := s.rpc.EthGetTransactionCount(address, "latest")
	s.Require().Nil(err)
	s.Require().Equal(16, count)
}

func (s *EthRPCTestSuite) TestEthGetBlockTransactionCountByHash() {
	hash := "0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238"
	s.registerResponse(`"0xb"`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockTransactionCountByHash")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, hash))
	})

	count, err := s.rpc.EthGetBlockTransactionCountByHash(hash)
	s.Require().Nil(err)
	s.Require().Equal(11, count)
}

func (s *EthRPCTestSuite) TestEthGetBlockTransactionCountByNumber() {
	number := 2384732
	s.registerResponse(`"0xe8"`, func(body []byte) {
		s.methodEqual(body, "eth_getBlockTransactionCountByNumber")
		s.paramsEqual(body, `["0x24635c"]`)
	})

	count, err := s.rpc.EthGetBlockTransactionCountByNumber(number)
	s.Require().Nil(err)
	s.Require().Equal(232, count)
}

func (s *EthRPCTestSuite) TestEthGetUncleCountByBlockHash() {
	hash := "0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238"
	s.registerResponse(`"0xa"`, func(body []byte) {
		s.methodEqual(body, "eth_getUncleCountByBlockHash")
		s.paramsEqual(body, fmt.Sprintf(`["%s"]`, hash))
	})

	count, err := s.rpc.EthGetUncleCountByBlockHash(hash)
	s.Require().Nil(err)
	s.Require().Equal(10, count)
}

func (s *EthRPCTestSuite) TestEthGetUncleCountByBlockNumber() {
	number := 3987434
	s.registerResponse(`"0x386"`, func(body []byte) {
		s.methodEqual(body, "eth_getUncleCountByBlockNumber")
		s.paramsEqual(body, `["0x3cd7ea"]`)
	})

	count, err := s.rpc.EthGetUncleCountByBlockNumber(number)
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

func TestEthRPCTestSuite(t *testing.T) {
	suite.Run(t, new(EthRPCTestSuite))
}
