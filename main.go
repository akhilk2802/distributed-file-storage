package main

import (
	"fmt"
	"log"

	"github.com/akhilk2802/distributed-file-storage/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOps{
		ListenAddr:     ":3000",
		HandeShakeFunc: p2p.NOPHandShakeFunc,
		Decoder:        p2p.DefaultDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("Message from main function go routine : %+v\n", msg)
		}
	}()

	err := tr.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}
	select {}
}
