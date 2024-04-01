package main

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalln(err)
	}
	listener, err := tls.Listen("tcp", ":3000", &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()
	var wg sync.WaitGroup
	defer wg.Wait()
	quit := make(chan struct{})
	defer close(quit)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		listener.Close()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Println(err)
			}
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			done := make(chan struct{})
			defer func() {
				<-done
			}()
			defer conn.Close()
			go func() {
				defer close(done)
				buf := make([]byte, 4096)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						if err != io.EOF && !errors.Is(err, net.ErrClosed) {
							log.Println(err)
						}
						return
					}
					items := strings.SplitN(strings.TrimSpace(string(buf[:n])), " ", 2)
					switch items[0] {
					case "ECHO":
						if _, err := conn.Write([]byte(items[1])); err != nil {
							if err != io.EOF && !errors.Is(err, net.ErrClosed) {
								log.Println(err)
							}
							return
						}
					case "QUIT":
						return
					}
				}
			}()
			select {
			case <-quit:
				return
			case <-done:
				return
			}
		}()
	}
}
