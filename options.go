package ethrpc

import (
	"io"
	"net/http"
)

type httpClient interface {
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
}

// WithClient add custom http client
func WithClient(client httpClient) func(rpc *EthRPC) {
	return func(rpc *EthRPC) {
		rpc.client = client
	}
}

// WithDebug set debug flag
func WithDebug(enabled bool) func(rpc *EthRPC) {
	return func(rpc *EthRPC) {
		rpc.Debug = enabled
	}
}
