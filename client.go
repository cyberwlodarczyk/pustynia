package pustynia

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sync"
)

type ClientConfig struct {
	RoomID    Code
	Addr      string
	TLSConfig *tls.Config
}

type Client struct {
	roomID Code
	conn   net.Conn
	once   sync.Once
	quit   chan empty
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	conn, err := tls.Dial("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		roomID: cfg.RoomID,
		conn:   conn,
		quit:   make(chan empty),
	}, nil
}

func (c *Client) Run() error {
	fail := make(chan error)
	join := make(chan empty)
	go func() {
		if _, err := c.conn.Write(c.roomID.Bytes()); err != nil {
			if isClosed(err) {
				err = nil
			}
			fail <- err
			return
		}
		close(join)
		buf := make([]byte, 4096)
		for {
			n, err := c.conn.Read(buf)
			if err != nil {
				if isClosed(err) {
					err = nil
				}
				fail <- err
				return
			}
			fmt.Printf("%s\n", buf[:n])
		}
	}()
	input := make(chan []byte)
	go func() {
		<-join
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			input <- s.Bytes()
		}
		fail <- s.Err()
	}()
	for {
		select {
		case <-c.quit:
			return nil
		case err := <-fail:
			return err
		case msg := <-input:
			if _, err := c.conn.Write(msg); err != nil {
				if isClosed(err) {
					return nil
				}
				return err
			}
		}
	}
}

func (c *Client) Close() error {
	c.once.Do(func() {
		close(c.quit)
	})
	return c.conn.Close()
}
