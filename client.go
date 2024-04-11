package pustynia

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

	"golang.org/x/crypto/argon2"
)

var ErrInvalidPassword = errors.New("pustynia: invalid password")

type ClientConfig struct {
	RoomID    Code
	Username  string
	Password  []byte
	Addr      string
	TLSConfig *tls.Config
}

type Client struct {
	roomID Code
	label  []byte
	hash   [sha256.Size]byte
	aead   cipher.AEAD
	conn   net.Conn
	once   sync.Once
	quit   chan empty
	fail   chan error
	join   chan empty
	input  chan []byte
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	defer zero(cfg.Password)
	salt := sha256.Sum256(cfg.RoomID.Bytes())
	key := argon2.IDKey(cfg.Password, salt[:], 1, 1<<16, uint8(runtime.NumCPU()), 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	conn, err := tls.Dial("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		roomID: cfg.RoomID,
		label:  []byte(fmt.Sprintf("%s> ", cfg.Username)),
		hash:   sha256.Sum256(cfg.Password),
		aead:   aead,
		conn:   conn,
		quit:   make(chan empty),
		fail:   make(chan error),
		join:   make(chan empty),
		input:  make(chan []byte),
	}, nil
}

func (c *Client) isError(_ int, err error) bool {
	if err != nil {
		if isClosed(err) {
			c.fail <- nil
		} else {
			c.fail <- err
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
	if !c.write(c.roomID.Bytes()) {
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
		ciphertext := make([]byte, int(binary.BigEndian.Uint32(msg[12:])))
		if !c.read(ciphertext) {
			return
		}
		plaintext, err := c.aead.Open(nil, msg[:12], ciphertext, nil)
		if err != nil {
			c.fail <- err
			return
		}
		for i := 0; i < len(c.label); i++ {
			fmt.Print("\b \b")
		}
		fmt.Printf("%s\n", plaintext)
		fmt.Printf("%s", c.label)
	}
}

func (c *Client) scan() {
	<-c.join
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s", c.label)
		if !s.Scan() {
			c.fail <- s.Err()
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
		case plaintext := <-c.input:
			msg := make([]byte, 16)
			if _, err := io.ReadFull(rand.Reader, msg[:12]); err != nil {
				return err
			}
			msg = c.aead.Seal(msg, msg[:12], plaintext, nil)
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
