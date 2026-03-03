package ipc

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
)

type HandlerFunc func(req *Request) (any, error)

type Server struct {
	path     string
	handlers map[string]HandlerFunc
	ln       net.Listener
}

func NewServer(socketPath string) *Server {
	return &Server{
		path:     socketPath,
		handlers: make(map[string]HandlerFunc),
	}
}

func (s *Server) Handle(method string, fn HandlerFunc) {
	s.handlers[method] = fn
}

func (s *Server) Listen() error {
	_ = os.Remove(s.path)
	ln, err := net.Listen("unix", s.path)
	if err != nil {
		return err
	}
	_ = os.Chmod(s.path, 0600)
	s.ln = ln
	return nil
}

func (s *Server) Serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) Close() {
	if s.ln != nil {
		_ = s.ln.Close()
	}
	_ = os.Remove(s.path)
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	enc := json.NewEncoder(conn)

	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			_ = enc.Encode(Response{Error: "invalid json"})
			continue
		}
		fn, ok := s.handlers[req.Method]
		if !ok {
			_ = enc.Encode(Response{ID: req.ID, Error: "unknown method: " + req.Method})
			continue
		}
		data, err := fn(&req)
		if err != nil {
			_ = enc.Encode(Response{ID: req.ID, Error: err.Error()})
			continue
		}
		_ = enc.Encode(Response{ID: req.ID, Data: data})
	}
}
