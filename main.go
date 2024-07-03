package main

import (
	"fmt"
	"log"

	"github.com/akhilk2802/distributed-file-storage/p2p"
)

func OnPeer(peer p2p.Peer) error {
	fmt.Println("Doing some logical task with the peer outisde of TCP Transport")
	peer.Close()
	return nil
}

func main() {
	// tcpOpts := p2p.TCPTransportOps{
	// 	ListenAddr:     ":3000",
	// 	HandeShakeFunc: p2p.NOPHandShakeFunc,
	// 	Decoder:        p2p.DefaultDecoder{},
	// 	OnPeer:         OnPeer,
	// }
	// tr := p2p.NewTCPTransport(tcpOpts)

	// go func() {
	// 	for {
	// 		msg := <-tr.Consume()
	// 		fmt.Printf("Message from main function go routine : %+v\n", msg)
	// 	}
	// }()

	// err := tr.ListenAndAccept()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// select {}

	tcpTransportOpts := p2p.TCPTransportOps{
		ListenAddr:     ":3000",
		HandeShakeFunc: p2p.NOPHandShakeFunc,
		Decoder:        p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		storageRoot:       "3000_network",
		pathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
	}
	s := NewFileServer(fileServerOpts)

	if err := s.start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
