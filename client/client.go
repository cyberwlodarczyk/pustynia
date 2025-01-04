package client

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"sync"

	"github.com/cyberwlodarczyk/pustynia/code"
	"github.com/cyberwlodarczyk/pustynia/server"
	"golang.org/x/crypto/argon2"
)

var ErrInvalidPassword = errors.New("invalid password")

type Config struct {
	RoomCode  code.Code
	Username  string
	Password  []byte
	Addr      string
	TLSConfig *tls.Config
}

type Client struct {
	roomCode code.Code
	label    []byte
	hash     [sha256.Size]byte
	aead     cipher.AEAD
	conn     net.Conn
	once     sync.Once
	quit     chan struct{}
	fail     chan error
	join     chan struct{}
	input    chan []byte
}

func New(cfg *Config) (*Client, error) {
	defer func() {
		for i := range cfg.Password {
			cfg.Password[i] = 0
		}
	}()
	if cfg.Addr == "" {
		cfg.Addr = server.DefaultAddr
	}
	salt := sha256.Sum256(cfg.RoomCode.Bytes())
	key := argon2.IDKey(cfg.Password, salt[:], 1, 1<<16, uint8(runtime.NumCPU()), 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error initializing AES: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error initializing AES GCM: %w", err)
	}
	conn, err := tls.Dial("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the server: %w", err)
	}
	return &Client{
		roomCode: cfg.RoomCode,
		label:    []byte(fmt.Sprintf("%s> ", cfg.Username)),
		hash:     sha256.Sum256(cfg.Password),
		aead:     gcm,
		conn:     conn,
		quit:     make(chan struct{}),
		fail:     make(chan error),
		join:     make(chan struct{}),
		input:    make(chan []byte),
	}, nil
}

func (c *Client) isError(_ int, err error) bool {
	if err != nil {
		if err == io.EOF || errors.Is(err, net.ErrClosed) {
			close(c.quit)
		} else {
			c.fail <- fmt.Errorf("error interacting with the connection: %w", err)
		}
		return false
	}
	return true
}

func (c *Client) read(b []byte) bool {
	return c.isError(io.ReadFull(c.conn, b))
}

func (c *Client) write(b []byte) bool {
	return c.isError(c.conn.Write(b))
}

func (c *Client) recv() {
	if !c.write(c.roomCode.Bytes()) {
		return
	}
	if !c.write(c.hash[:]) {
		return
	}
	var ok [1]byte
	if !c.read(ok[:]) {
		return
	}
	if ok[0] == 0 {
		c.fail <- ErrInvalidPassword
		return
	}
	close(c.join)
	for {
		msg := make([]byte, 16)
		if !c.read(msg) {
			return
		}
		ct := make([]byte, int(binary.BigEndian.Uint32(msg[12:])))
		if !c.read(ct) {
			return
		}
		pt, err := c.aead.Open(nil, msg[:12], ct, nil)
		if err != nil {
			c.fail <- fmt.Errorf("error decrypting and authenticating the message: %w", err)
			return
		}
		for i := 0; i < len(c.label); i++ {
			fmt.Print("\b \b")
		}
		fmt.Printf("%s\n", pt)
		fmt.Printf("%s", c.label)
	}
}

func (c *Client) scan() {
	<-c.join
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s", c.label)
		if !s.Scan() {
			if err := s.Err(); err != nil {
				c.fail <- fmt.Errorf("error reading from standard input: %w", err)
			} else {
				close(c.quit)
			}
			return
		}
		var b bytes.Buffer
		b.Write(c.label)
		b.Write(s.Bytes())
		c.input <- b.Bytes()
	}
}

func (c *Client) Run() error {
	go c.recv()
	go c.scan()
	for {
		select {
		case <-c.quit:
			return nil
		case err := <-c.fail:
			return err
		case pt := <-c.input:
			msg := make([]byte, 16)
			if _, err := io.ReadFull(rand.Reader, msg[:12]); err != nil {
				return fmt.Errorf("error generating new nonce: %w", err)
			}
			msg = c.aead.Seal(msg, msg[:12], pt, nil)
			binary.BigEndian.PutUint32(msg[12:16], uint32(len(msg[16:])))
			c.write(msg)
		}
	}
}

func (c *Client) Close() error {
	c.once.Do(func() {
		close(c.quit)
	})
	return c.conn.Close()
}
