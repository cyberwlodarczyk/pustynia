package pustynia

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"sync"

	"golang.org/x/crypto/argon2"
)

type ClientConfig struct {
	RoomID    Code
	Password  []byte
	Addr      string
	TLSConfig *tls.Config
}

type Client struct {
	roomID Code
	gcm    cipher.AEAD
	conn   net.Conn
	once   sync.Once
	quit   chan empty
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	defer zero(cfg.Password)
	salt := sha256.Sum256(cfg.RoomID.Bytes())
	key := argon2.IDKey(cfg.Password, salt[:], 1, 1<<16, uint8(runtime.NumCPU()), 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	conn, err := tls.Dial("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		roomID: cfg.RoomID,
		gcm:    gcm,
		conn:   conn,
		quit:   make(chan empty),
	}, nil
}

func (c *Client) Run() error {
	fail := make(chan error)
	join := make(chan empty)
	check := func(_ int, err error) bool {
		if err != nil {
			if isClosed(err) {
				fail <- nil
			} else {
				fail <- err
			}
			return false
		}
		return true
	}
	read := func(b []byte) bool {
		return check(io.ReadFull(c.conn, b))
	}
	write := func(b []byte) bool {
		return check(c.conn.Write(b))
	}
	go func() {
		if !write(c.roomID.Bytes()) {
			return
		}
		close(join)
		for {
			msg := make([]byte, 16)
			if !read(msg) {
				return
			}
			ciphertext := make([]byte, int(binary.BigEndian.Uint32(msg[12:])))
			if !read(ciphertext) {
				return
			}
			plaintext, err := c.gcm.Open(nil, msg[:12], ciphertext, nil)
			if err != nil {
				fail <- err
				return
			}
			fmt.Printf("%s\n", plaintext)
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
		case plaintext := <-input:
			msg := make([]byte, 16)
			if _, err := io.ReadFull(rand.Reader, msg[:12]); err != nil {
				return err
			}
			msg = c.gcm.Seal(msg, msg[:12], plaintext, nil)
			binary.BigEndian.PutUint32(msg[12:16], uint32(len(msg[16:])))
			write(msg)
		}
	}
}

func (c *Client) Close() error {
	c.once.Do(func() {
		close(c.quit)
	})
	return c.conn.Close()
}
