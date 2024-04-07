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
	addr := flag.String("addr", ":8888", "server address")
	room := flag.String("room", "", "room code")
	flag.Parse()
	if *room == "" {
		log.Fatalln("please specify the --room flag")
	}
	roomID, ok := pustynia.ParseCode(*room)
	if !ok {
		log.Fatalln("please specify the valid --room flag")
	}
	client, err := pustynia.NewClient(&pustynia.ClientConfig{
		RoomID:    roomID,
		Addr:      *addr,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		client.Close()
	}()
	if err = client.Run(); err != nil {
		log.Fatalln(err)
	}
}
