# Ethrpc
Golang client for ethereum [JSON RPC API](https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getcompilers).

##### Usage:
```go
package main

import (
	"log"
    
	"github.com/onrik/ethrpc"
)

func main() {
	client := ethrcp.NewEthRPC("http://127.0.0.1:8545")

	version, err := client.Web3ClientVersion()
    if err != nil {
        log.Fatal(err)
    }
}

```
