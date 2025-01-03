package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberwlodarczyk/pustynia/client"
	"github.com/cyberwlodarczyk/pustynia/code"
	"golang.org/x/term"
)

func run() error {
	addr := flag.String("addr", ":8888", "server address")
	room := flag.String("room", "", "room code")
	username := flag.String("user", "anonymous", "room username")
	flag.Parse()
	if *room == "" {
		return errors.New("please specify the --room flag")
	}
	roomID, ok := code.Parse(*room)
	if !ok {
		return errors.New("please specify the valid --room flag")
	}
	fmt.Print("password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("error reading the password: %w", err)
	}
	fmt.Print("\n")
	client, err := client.New(&client.Config{
		RoomID:    roomID,
		Username:  *username,
		Password:  password,
		Addr:      *addr,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})
	if err != nil {
		return fmt.Errorf("error starting the client: %w", err)
	}
	defer client.Close()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		client.Close()
	}()
	if err := client.Run(); err != nil {
		return fmt.Errorf("error running the client: %w", err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
