package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/scrye/squic/src/network"
)

var (
	mtu         = flag.Int("mtu", 1200, "mtu for link between client and server")
	statsPath   = flag.String("stats", "stats.dat", "file to dump statistics to")
	inFilePath  = flag.String("input", "input.dat", "file sent by the client")
	outFilePath = flag.String("output", "output.dat", "file received by the client")
)

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}
}

func realMain() error {
	flag.Parse()
	quic.InjectMTU(*mtu)

	clientWriteLog := new(bytes.Buffer)
	serverWriteLog := new(bytes.Buffer) // Discarded
	defer func() {
		if err := ioutil.WriteFile(*statsPath, clientWriteLog.Bytes(), 0644); err != nil {
			fmt.Printf("collector error, err = %v", err)
		}
	}()

	fmt.Fprintf(clientWriteLog, "MTU: %v\n", *mtu)

	clientConn, serverConn, err := network.NewMTUConnPair("127.0.0.1:0", "127.0.0.1:6062", *mtu, clientWriteLog, serverWriteLog)
	if err != nil {
		return fmt.Errorf("udp error, err = %v", err)
	}

	server := &network.Server{PacketConn: serverConn, Out: serverWriteLog}
	client := &network.Client{PacketConn: clientConn, RemoteAddress: "127.0.0.1:6062", Out: clientWriteLog}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("server error, err = %v", err)
		}
	}()
	time.Sleep(200 * time.Millisecond) // wait for srv

	inFile, err := os.Open(*inFilePath)
	if err != nil {
		return fmt.Errorf("unable to open input file, err = %v", err)
	}
	outFile, err := os.Create(*outFilePath)
	if err != nil {
		return fmt.Errorf("unable to open output file, err = %v", err)
	}
	if err := client.Run(inFile, outFile); err != nil {
		return fmt.Errorf("client error, err = %v", err)
	}
	return nil
}
