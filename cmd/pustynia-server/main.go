package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberwlodarczyk/pustynia"
	"github.com/sirupsen/logrus"
)

func run() error {
	addr := flag.String("addr", ":8888", "listen address")
	certFile := flag.String("tls-cert", "", "tls certificate file location")
	keyFile := flag.String("tls-key", "", "tls key file location")
	flag.Parse()
	if *certFile == "" {
		return errors.New("please specify the --tls-cert flag")
	}
	if *keyFile == "" {
		return errors.New("please specify the --tls-key flag")
	}
	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return fmt.Errorf("error loading X509 key pair: %w", err)
	}
	server, err := pustynia.NewServer(&pustynia.ServerConfig{
		Addr: *addr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
		},
	})
	if err != nil {
		return fmt.Errorf("error starting the server: %w", err)
	}
	defer server.Close()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		server.Close()
	}()
	if err = server.Run(); err != nil {
		return fmt.Errorf("error running the server: %w", err)
	}
	return nil
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}
