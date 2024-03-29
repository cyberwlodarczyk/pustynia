package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	stop := make(chan struct{})
	go func() {
		b := make([]byte, 4096)
		for {
			select {
			case <-stop:
				return
			default:
				n, err := conn.Read(b)
				if err != nil {
					log.Println(err)
					return
				}
				fmt.Println(string(b[:n]))
			}
		}
	}()
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		msg := s.Bytes()
		if bytes.Equal(msg, []byte("exit")) {
			close(stop)
			break
		}
		if _, err = conn.Write(msg); err != nil {
			log.Println(err)
			break
		}
	}
	if err = s.Err(); err != nil {
		log.Fatalln(err)
	}
}
