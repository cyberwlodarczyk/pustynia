package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberwlodarczyk/pustynia"
	"golang.org/x/term"
)

func main() {
	addr := flag.String("addr", ":8888", "server address")
	room := flag.String("room", "", "room code")
	username := flag.String("user", "anonymous", "room username")
	flag.Parse()
	if *room == "" {
		log.Fatalln("please specify the --room flag")
	}
	roomID, ok := pustynia.ParseCode(*room)
	if !ok {
		log.Fatalln("please specify the valid --room flag")
	}
	fmt.Print("password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print("\n")
	client, err := pustynia.NewClient(&pustynia.ClientConfig{
		RoomID:    roomID,
		Username:  *username,
		Password:  password,
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
		if errors.Is(err, pustynia.ErrInvalidPassword) {
			fmt.Println("invalid password")
		} else {
			log.Fatalln(err)
		}
	}
}
