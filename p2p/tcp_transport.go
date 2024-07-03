package p2p

import (
	"fmt"
	"net"
	"reflect"
)

// remote node over TCP connection
type TCPPeer struct {
	// conn is existing connection of the peer
	conn net.Conn
	// if we dial a connection -> outbound == true
	// if we accept a connection -> outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// Close implements the peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOps struct {
	ListenAddr     string
	HandeShakeFunc HandeShakeFunc
	Decoder        Decoder
	OnPeer         func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts TCPTransportOps
	listener         net.Listener
	rpcch            chan RPC

	// mu    sync.RWMutex
	// peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements the transport interface, which will return read only channel
// for reading the incoming messages recieved from another peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.TCPTransportOpts.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP Accept error : %s\n", err)
		}
		go t.handleConn(conn)
	}
}

type Temp struct {
}

func (t *TCPTransport) handleConn(conn net.Conn) {

	var err error
	defer func() {
		fmt.Printf("Dropping peer connection : %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err = t.TCPTransportOpts.HandeShakeFunc(peer); err != nil {
		fmt.Printf("TCP Handshake error : %s\n", err)
		return
	}

	if t.TCPTransportOpts.OnPeer != nil {
		if err = t.TCPTransportOpts.OnPeer(peer); err != nil {
			return
		}
	}

	rpc := RPC{}
	for {
		err := t.TCPTransportOpts.Decoder.Decode(conn, &rpc)
		fmt.Println("Type of Error : ", reflect.TypeOf(err))
		// panic(err)

		if err != nil {
			fmt.Printf("TCP Read error : %s\n", err)
			return
		}
		rpc.From = conn.RemoteAddr()
		fmt.Printf("message : %+v\n", rpc)
		t.rpcch <- rpc
	}
}
