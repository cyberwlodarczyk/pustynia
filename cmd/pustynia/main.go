package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conn, err := tls.Dial("tcp", ":3000", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	quit := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		close(quit)
	}()
	go func() {
		defer close(quit)
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				return
			}
			fmt.Println(string(buf[:n]))
		}
	}()
	input := make(chan string)
	go func() {
		defer close(quit)
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			input <- s.Text()
		}
		if err := s.Err(); err != nil {
			log.Println(err)
		}
	}()
	for {
		select {
		case <-quit:
			return
		case msg := <-input:
			if _, err = conn.Write([]byte(msg)); err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				return
			}
		}
	}
}
