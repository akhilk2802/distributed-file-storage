package p2p

const IncomingStream = 0x2
const IncomingMessage = 0x1

// Holds the arbitrary message that is sent between each nodes in the TCP Connections
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
