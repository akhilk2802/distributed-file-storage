package p2p

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {

	opts := TCPTransportOps{
		ListenAddr:     ":3000",
		HandeShakeFunc: NOPHandShakeFunc,
		Decoder:        DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	fmt.Println("Listen Address : ", tr.TCPTransportOpts.ListenAddr)

	assert.Equal(t, ":3000", tr.TCPTransportOpts.ListenAddr)
	assert.Nil(t, tr.ListenAndAccept())
}
