package p2p

import "net"

// Holds the arbitrary message that is sent between each nodes in the TCP Connections
type RPC struct {
	From    net.Addr
	Payload []byte
}
