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
	"github.com/cyberwlodarczyk/pustynia/server"
	"golang.org/x/term"
)

func run() error {
	addr := flag.String("addr", server.DefaultAddr, "server address")
	room := flag.String("room", "", "room code")
	username := flag.String("user", "anonymous", "room username")
	flag.Parse()
	if *room == "" {
		return errors.New("please specify the --room flag")
	}
	roomCode, ok := code.Parse(*room)
	if !ok {
		return errors.New("please specify the valid --room flag")
	}
	fmt.Print("password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("error reading the password: %w", err)
	}
	fmt.Print("\n")
	c, err := client.New(&client.Config{
		RoomCode:  roomCode,
		Username:  *username,
		Password:  password,
		Addr:      *addr,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})
	if err != nil {
		return fmt.Errorf("error starting the client: %w", err)
	}
	defer c.Close()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		c.Close()
	}()
	if err := c.Run(); err != nil {
		if errors.Is(err, client.ErrInvalidPassword) {
			return err
		}
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
