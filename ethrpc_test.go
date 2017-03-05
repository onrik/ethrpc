package ethrpc

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
)

type EthRPCTestSuite struct {
	suite.Suite
	rpc *EthRPC
}

func (s *EthRPCTestSuite) registerResponse(response string) {
	body := fmt.Sprintf(`{"jsonrpc":"2.0", "id":1, "result": %s}`, response)
	httpmock.RegisterResponder("POST", s.rpc.url, httpmock.NewStringResponder(200, body))
}

func (s *EthRPCTestSuite) SetupSuite() {
	s.rpc = NewEthRPC("http://127.0.0.1:8545")

	httpmock.Activate()
}

func (s *EthRPCTestSuite) TearDownSuite() {
	httpmock.Deactivate()
}

func (s *EthRPCTestSuite) TearDownTest() {
	httpmock.Reset()
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

func TestEthRPCTestSuite(t *testing.T) {
	suite.Run(t, new(EthRPCTestSuite))
}
