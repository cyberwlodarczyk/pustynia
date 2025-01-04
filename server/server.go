package server

import (
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/cyberwlodarczyk/pustynia/code"
	"github.com/sirupsen/logrus"
)

var DefaultAddr = ":1984"

type Config struct {
	Addr      string
	TLSConfig *tls.Config
}

type user struct {
	id       int
	messages chan []byte
}

type room struct {
	id    int
	code  code.Code
	hash  [sha256.Size]byte
	users map[int]user
}

type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
	once     sync.Once
	quit     chan struct{}
	rwMutex  sync.RWMutex
	roomID   int
	userID   int
	rooms    map[int]room
	codes    map[code.Code]int
}

func New(cfg *Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = DefaultAddr
	}
	l, err := tls.Listen("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: l,
		quit:     make(chan struct{}),
		rooms:    make(map[int]room),
		codes:    make(map[code.Code]int),
	}, nil
}

type session struct {
	roomID int
	userID int
}

func (s *Server) createSession(c code.Code, h [sha256.Size]byte) *session {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	var r room
	rid, ok := s.codes[c]
	if ok {
		r = s.rooms[rid]
		if r.hash != h {
			logrus.WithField("roomId", r.id).Info("user authentication failed")
			return nil
		}
	} else {
		rid = s.roomID
		s.roomID++
		s.codes[c] = rid
		r = room{rid, c, h, make(map[int]user)}
		s.rooms[rid] = r
		logrus.WithField("roomId", r.id).WithField("roomCode", r.code.String()).Info("room created")
	}
	uid := s.userID
	s.userID++
	u := user{uid, make(chan []byte)}
	r.users[uid] = u
	logrus.WithField("roomId", r.id).WithField("userId", u.id).Info("user created")
	return &session{r.id, u.id}
}

func (s *Server) deleteSession(sess *session) {
	s.rwMutex.RLock()
	r := s.rooms[sess.roomID]
	u := r.users[sess.userID]
	close(u.messages)
	s.rwMutex.RUnlock()
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	delete(r.users, u.id)
	logrus.WithField("roomId", r.id).WithField("userId", u.id).Info("user deleted")
	if len(r.users) == 0 {
		delete(s.rooms, r.id)
		delete(s.codes, r.code)
		logrus.WithField("roomId", r.id).Info("room deleted")
	}
}

func (s *Server) sendMessage(sess *session, msg []byte) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	r := s.rooms[sess.roomID]
	for _, u := range r.users {
		if u.id != sess.userID {
			u.messages <- msg
		}
	}
}

func (s *Server) recvMessage(sess *session) []byte {
	s.rwMutex.RLock()
	r := s.rooms[sess.roomID]
	u := r.users[sess.userID]
	s.rwMutex.RUnlock()
	return <-u.messages
}

func (s *Server) isError(n int, err error) (int, bool) {
	if err != nil {
		if err != io.EOF && !errors.Is(err, net.ErrClosed) {
			logrus.Error(err)
		}
		return 0, false
	}
	return n, true
}

type peer struct {
	srv  *Server
	conn net.Conn
	sess *session
	once sync.Once
	exit chan struct{}
	join chan struct{}
}

func (p *peer) read(b []byte) (int, bool) {
	return p.srv.isError(p.conn.Read(b))
}

func (p *peer) readFull(b []byte) bool {
	_, ok := p.srv.isError(io.ReadFull(p.conn, b))
	return ok
}

func (p *peer) write(b []byte) bool {
	_, ok := p.srv.isError(p.conn.Write(b))
	return ok
}

func (p *peer) leave() {
	p.once.Do(func() {
		if p.sess != nil {
			p.srv.deleteSession(p.sess)
			p.sess = nil
		}
		close(p.exit)
	})
}

func (p *peer) recv() {
	defer p.leave()
	<-p.join
	for {
		msg := p.srv.recvMessage(p.sess)
		if msg == nil {
			return
		}
		if !p.write(msg) {
			return
		}
	}
}

func (p *peer) send() {
	defer p.leave()
	b := make([]byte, 4096)
	for {
		cb := make([]byte, code.Size)
		if !p.readFull(cb) {
			return
		}
		c, ok := code.Parse(string(cb))
		if !ok {
			return
		}
		var h [sha256.Size]byte
		if !p.readFull(h[:]) {
			return
		}
		p.sess = p.srv.createSession(c, h)
		if p.sess == nil {
			p.write([]byte{0})
			return
		}
		if !p.write([]byte{1}) {
			return
		}
		close(p.join)
		for {
			n, ok := p.read(b)
			if !ok {
				return
			}
			p.srv.sendMessage(p.sess, b[:n])
		}
	}
}

func (p *peer) run() {
	go p.recv()
	go p.send()
	select {
	case <-p.srv.quit:
		p.conn.Close()
		<-p.exit
	case <-p.exit:
		p.conn.Close()
	}
	p.srv.wg.Done()
}

func (s *Server) Run() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("error accepting connections: %w", err)
		}
		s.wg.Add(1)
		p := &peer{
			srv:  s,
			conn: conn,
			exit: make(chan struct{}),
			join: make(chan struct{}),
		}
		go p.run()
	}
}

func (s *Server) Close() error {
	s.once.Do(func() {
		close(s.quit)
	})
	s.wg.Wait()
	return s.listener.Close()
}
