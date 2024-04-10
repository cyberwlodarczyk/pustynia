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
}

type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
	once     sync.Once
	quit     chan empty
	rwMutex  sync.RWMutex
	userID   int
	rooms    map[Code]room
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	listener, err := tls.Listen("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Server{
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

func (s *Server) handleConnection(conn net.Conn) {
	var sess *session
	exit := make(chan empty)
	join := make(chan empty)
	check := func(n int, err error) (int, bool) {
		if err != nil {
			if !isClosed(err) {
				log.Println(err)
			}
			return 0, false
		}
		return n, true
	}
	read := func(b []byte) (int, bool) {
		return check(conn.Read(b))
	}
	readFull := func(b []byte) bool {
		_, ok := check(io.ReadFull(conn, b))
		return ok
	}
	write := func(b []byte) bool {
		_, ok := check(conn.Write(b))
		return ok
	}
	leave := sync.OnceFunc(func() {
		if sess != nil {
			s.leaveRoom(sess)
			sess = nil
		}
		close(exit)
	})
	go func() {
		defer leave()
		<-join
		for {
			msg := s.recvMessage(sess)
			if msg == nil {
				return
			}
			if !write(msg) {
				return
			}
		}
	}()
	go func() {
		defer leave()
		buf := make([]byte, 4096)
		for {
			code := make([]byte, CodeSize)
			if !readFull(code) {
				return
			}
			roomID, ok := ParseCode(string(code))
			if !ok {
				return
			}
			var hash [sha256.Size]byte
			if !readFull(hash[:]) {
				return
			}
			sess = s.joinRoom(roomID, hash)
			if sess == nil {
				write([]byte{0})
				return
			}
			if !write([]byte{1}) {
				return
			}
			close(join)
			for {
				n, ok := read(buf)
				if !ok {
					return
				}
				s.sendMessage(sess, buf[:n])
			}
		}
	}()
	select {
	case <-s.quit:
		conn.Close()
		<-exit
	case <-exit:
		conn.Close()
	}
	s.wg.Done()
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
		go s.handleConnection(conn)
	}
}

func (s *Server) Close() error {
	s.once.Do(func() {
		close(s.quit)
	})
	s.wg.Wait()
	return s.listener.Close()
}
