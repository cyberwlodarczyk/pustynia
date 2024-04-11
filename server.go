package pustynia

import (
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

type room struct {
	hash  [sha256.Size]byte
	users map[int]chan []byte
}

type ServerConfig struct {
	Addr      string
	TLSConfig *tls.Config
	ErrorLog  *log.Logger
}

type Server struct {
	errorLog *log.Logger
	listener net.Listener
	wg       sync.WaitGroup
	once     sync.Once
	quit     chan empty
	rwMutex  sync.RWMutex
	userID   int
	rooms    map[Code]room
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	if cfg.ErrorLog == nil {
		cfg.ErrorLog = log.Default()
	}
	listener, err := tls.Listen("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Server{
		errorLog: cfg.ErrorLog,
		listener: listener,
		quit:     make(chan empty),
		rooms:    make(map[Code]room),
	}, nil
}

type session struct {
	roomID Code
	userID int
}

func (s *Server) joinRoom(roomID Code, hash [sha256.Size]byte) *session {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	r, ok := s.rooms[roomID]
	if ok {
		if r.hash != hash {
			return nil
		}
	} else {
		r = room{hash, make(map[int]chan []byte)}
		s.rooms[roomID] = r
	}
	userID := s.userID
	s.userID++
	r.users[userID] = make(chan []byte)
	return &session{roomID, userID}
}

func (s *Server) leaveRoom(sess *session) {
	s.rwMutex.RLock()
	r := s.rooms[sess.roomID]
	ch := r.users[sess.userID]
	close(ch)
	s.rwMutex.RUnlock()
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	delete(r.users, sess.userID)
	if len(r.users) == 0 {
		delete(s.rooms, sess.roomID)
	}
}

func (s *Server) sendMessage(sess *session, msg []byte) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	r := s.rooms[sess.roomID]
	for userID, ch := range r.users {
		if userID != sess.userID {
			ch <- msg
		}
	}
}

func (s *Server) recvMessage(sess *session) []byte {
	s.rwMutex.RLock()
	r := s.rooms[sess.roomID]
	ch := r.users[sess.userID]
	s.rwMutex.RUnlock()
	return <-ch
}

func (s *Server) isError(n int, err error) (int, bool) {
	if err != nil {
		if !isClosed(err) {
			s.errorLog.Println(err)
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
	exit chan empty
	join chan empty
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
			p.srv.leaveRoom(p.sess)
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
	buf := make([]byte, 4096)
	for {
		code := make([]byte, CodeSize)
		if !p.readFull(code) {
			return
		}
		roomID, ok := ParseCode(string(code))
		if !ok {
			return
		}
		var hash [sha256.Size]byte
		if !p.readFull(hash[:]) {
			return
		}
		p.sess = p.srv.joinRoom(roomID, hash)
		if p.sess == nil {
			p.write([]byte{0})
			return
		}
		if !p.write([]byte{1}) {
			return
		}
		close(p.join)
		for {
			n, ok := p.read(buf)
			if !ok {
				return
			}
			p.srv.sendMessage(p.sess, buf[:n])
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
			return err
		}
		s.wg.Add(1)
		p := &peer{
			srv:  s,
			conn: conn,
			exit: make(chan empty),
			join: make(chan empty),
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
