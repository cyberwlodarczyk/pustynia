package pustynia

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

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
	rooms    map[Code]map[int]chan []byte
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	listener, err := tls.Listen("tcp", cfg.Addr, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: listener,
		quit:     make(chan empty),
		rooms:    make(map[Code]map[int]chan []byte),
	}, nil
}

type session struct {
	roomID Code
	userID int
}

func (s *Server) joinRoom(roomID Code) *session {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	room, ok := s.rooms[roomID]
	if !ok {
		room = make(map[int]chan []byte)
		s.rooms[roomID] = room
	}
	userID := s.userID
	s.userID++
	room[userID] = make(chan []byte)
	return &session{roomID, userID}
}

func (s *Server) leaveRoom(sess *session) {
	s.rwMutex.RLock()
	room := s.rooms[sess.roomID]
	ch := room[sess.userID]
	ch <- nil
	s.rwMutex.RUnlock()
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	delete(room, sess.userID)
	if len(room) == 0 {
		delete(s.rooms, sess.roomID)
	}
}

func (s *Server) sendMessage(sess *session, msg []byte) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	for userID, ch := range s.rooms[sess.roomID] {
		if userID != sess.userID {
			ch <- msg
		}
	}
}

func (s *Server) recvMessage(sess *session) []byte {
	s.rwMutex.RLock()
	ch := s.rooms[sess.roomID][sess.userID]
	s.rwMutex.RUnlock()
	return <-ch
}

func (s *Server) handleConnection(conn net.Conn) {
	var sess *session
	exit := make(chan empty)
	join := make(chan empty)
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
			if _, err := conn.Write(msg); err != nil {
				if !isClosed(err) {
					log.Println(err)
				}
				return
			}
		}
	}()
	go func() {
		defer leave()
		buf := make([]byte, 4096)
		for {
			code := make([]byte, CodeSize)
			if _, err := io.ReadFull(conn, code); err != nil {
				if !isClosed(err) {
					log.Println(err)
				}
				return
			}
			roomID, ok := ParseCode(string(code))
			if !ok {
				return
			}
			sess = s.joinRoom(roomID)
			close(join)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					if !isClosed(err) {
						log.Println(err)
					}
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
