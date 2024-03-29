package main

import (
	"io"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			defer conn.Close()
			b := make([]byte, 4096)
			for {
				n, err := conn.Read(b)
				if err != nil {
					if err != io.EOF {
						log.Println(err)
					}
					return
				}
				if _, err = conn.Write(b[:n]); err != nil {
					log.Println(err)
					return
				}
			}
		}()
	}
}
