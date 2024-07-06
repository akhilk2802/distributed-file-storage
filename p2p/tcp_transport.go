package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// remote node over TCP connection
type TCPPeer struct {
	// conn is existing connection of the peer
	// TCP Connection
	net.Conn
	// if we dial a connection -> outbound == true
	// if we accept a connection -> outbound == false
	outbound bool
	wg       *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// Close implements the peer interface
func (p *TCPPeer) Close() error {
	return p.Conn.Close()
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
		rpcch:            make(chan RPC, 1024),
	}
}

// Addr implements transport interface returns the address
// Transport accepting the connections
func (t *TCPTransport) Addr() string {
	return t.TCPTransportOpts.ListenAddr
}

// implements the peer interface and will return the remoteAddr of its under lying peer
func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.Conn.RemoteAddr()
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
	log.Printf("TCP Transport listening on port: %s\n", t.TCPTransportOpts.ListenAddr)
	return nil
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	fmt.Println("Address : ", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	fmt.Println("Incoming Connection : ", conn)

	go t.handleConn(conn, true)
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP Accept error : %s\n", err)
		}
		go t.handleConn(conn, false)
	}
}

type Temp struct {
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {

	var err error
	defer func() {
		fmt.Printf("Dropping peer connection : %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.TCPTransportOpts.HandeShakeFunc(peer); err != nil {
		fmt.Printf("TCP Handshake error : %s\n", err)
		return
	}

	if t.TCPTransportOpts.OnPeer != nil {
		if err = t.TCPTransportOpts.OnPeer(peer); err != nil {
			return
		}
	}

	for {
		rpc := RPC{}
		err := t.TCPTransportOpts.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP Read error : %s\n", err)
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("waiting till the stream is done : %s \n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("Stream closed resuming read loop : %s \n", conn.RemoteAddr())
			continue
		}

		t.rpcch <- rpc
	}
}
