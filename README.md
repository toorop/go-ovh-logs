# go-ovh-logs
Golang package for sending logs to the [OVH logs PAAS](https://www.ovh.com/fr/data-platforms/logs/)

[![GoDoc](https://godoc.org/github.com/toorop/go-ovh-logs?status.svg)](https://godoc.org/github.com/toorop/go-ovh-logs)

It's completely API compatible with the standard library logger.

It implements GELF format (TCP or UDP, compressed or not).

Example:

```go
package main

import (
	ovhlogs "github.com/toorop/go-ovh-logs"
)

func main() {
	OVHLogger := ovhlogs.New("STREAM_TOKEN", ovhlogs.GelfUDP, ovhlogs.CompressNone, false)
	OVHLogger.Print("Hello World !")
}

```
See [Examples folder](examples) for more... examples ;)

## Support this project & open-source
If this project is useful for you, please consider making a donation.

### Bitcoin

Address: 1JvMRNRxiTiN9H7LyZTq4yzR7ez86M7ND6

![Bitcoin QR code](https://raw.githubusercontent.com/toorop/wallets/master/btc.png)


### Ethereum

Address: 0xA84684B45969efbD54fd25A1e2eD8C7790A0C497

![ETH QR code](https://raw.githubusercontent.com/toorop/wallets/master/eth.png)
