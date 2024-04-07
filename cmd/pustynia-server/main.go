package main

import (
	"crypto/tls"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberwlodarczyk/pustynia"
)

func main() {
	addr := flag.String("addr", ":8888", "listen address")
	certFile := flag.String("tls-cert", "", "tls certificate file location")
	keyFile := flag.String("tls-key", "", "tls key file location")
	flag.Parse()
	if *certFile == "" {
		log.Fatalln("please specify the --tls-cert flag")
	}
	if *keyFile == "" {
		log.Fatalln("please specify the --tls-key flag")
	}
	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalln(err)
	}
	server, err := pustynia.NewServer(&pustynia.ServerConfig{
		Addr: *addr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		server.Close()
	}()
	if err = server.Run(); err != nil {
		log.Fatalln(err)
	}
}
