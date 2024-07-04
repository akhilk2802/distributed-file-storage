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

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOps{
		ListenAddr:     listenAddr,
		HandeShakeFunc: p2p.NOPHandShakeFunc,
		Decoder:        p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		storageRoot:       listenAddr + "_network",
		pathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootStrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)

	// go func() {
	// 	time.Sleep(time.Second * 3)
	// 	s.stop()
	// }()

	// if err := s.start(); err != nil {
	// 	log.Fatal(err)
	// }
	tcpTransport.TCPTransportOpts.OnPeer = s.OnPeer
	return s
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

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		if err := s1.start(); err != nil {
			log.Fatal("unable to start the server : ", err)
		}
	}()
	s2.start()
}
