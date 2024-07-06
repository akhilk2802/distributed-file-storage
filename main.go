package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

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

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		if err := s1.start(); err != nil {
			log.Fatal("unable to start the server : ", err)
		}
	}()
	go s2.start()
	time.Sleep(time.Second * 3)

	// data := bytes.NewReader([]byte("my data file is here"))
	// s2.Store("coolPicture.jpg", data)
	// time.Sleep(time.Millisecond * 5)

	r, err := s2.Get("coolPicture.jpg")
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
	select {}
}
