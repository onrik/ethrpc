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

func (s *EthRPCTestSuite) registerResponse(response string) {
	body := fmt.Sprintf(`{"jsonrpc":"2.0", "id":1, "result": %s}`, response)
	httpmock.RegisterResponder("POST", s.rpc.url, httpmock.NewStringResponder(200, body))
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

func (s *EthRPCTestSuite) TestEthSyncing() {
	expected := &Syncing{
		IsSyncing:     false,
		CurrentBlock:  0,
		HighestBlock:  0,
		StartingBlock: 0,
	}
	s.registerResponse(`false`)
	syncing, err := s.rpc.EthSyncing()

	s.Require().Nil(err)
	s.Require().Equal(expected, syncing)

	httpmock.Reset()
	s.registerResponse(`{
		"currentBlock": "0x8c3be",
		"highestBlock": "0x9bb3b",
		"startingBlock": "0x0"
	}`)

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

func (s *EthRPCTestSuite) TestSendTransaction() {
	response := `{"jsonrpc":"2.0", "id":1, "result": "0xea1115eb5"}`

	t := T{
		From:     "0x3cc1a3c082944b9dba70e490e481dd56",
		To:       "0x1bf21cb1dc384d019a885a06973f7308",
		Gas:      24900,
		GasPrice: big.NewInt(5000000000),
		Value:    big.NewInt(1000000000000000000), // 1 ETH
		Data:     "some data",
		Nonce:    98384,
	}

	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
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

		return httpmock.NewStringResponse(200, response), nil
	})

	txid, err := s.rpc.EthSendTransaction(t)
	s.Require().Nil(err)
	s.Require().Equal("0xea1115eb5", txid)

	httpmock.Reset()

	t = T{}
	httpmock.RegisterResponder("POST", s.rpc.url, func(request *http.Request) (*http.Response, error) {
		body := s.getBody(request)
		s.methodEqual(body, "eth_sendTransaction")
		s.paramsEqual(body, `[{
			"from": ""
		}]`)

		return httpmock.NewStringResponse(200, response), nil
	})

	txid, err = s.rpc.EthSendTransaction(t)
	s.Require().Nil(err)
	s.Require().Equal("0xea1115eb5", txid)
}

func TestEthRPCTestSuite(t *testing.T) {
	suite.Run(t, new(EthRPCTestSuite))
}
