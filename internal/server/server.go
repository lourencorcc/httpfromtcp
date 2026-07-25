package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
)

/* type serverState int

const (
	StateInit serverState = iota
	StateBusy
	StateOpen
	StateClosed
) */

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	// creates a net.Listener and returns a new Server instance.
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("error while creating listener on port %d: %w", port, err)
	}

	server := &Server{
		listener: l, // fine cz closed gets the 0 value because we mustn't copy an atomic bool
	}

	go server.listen()

	return server, err

}

func (s *Server) listen() {

	l := s.listener
	defer l.Close()
	for {

		// waits for a connection.
		conn, err := l.Accept()
		if err != nil {
			if s.closed.Load() {
				log.Println(err)
			} else {
				log.Fatalf("unexpected error while listening: %w", err)
			}
			return
		}
		// handle the connection in a new goroutine.
		// the loop then returns to accepting, so that multiple connections may be served concurrently.
		go func(c net.Conn) {
			s.handle(c)
		}(conn)

	}
}

func (s *Server) Close() {
	s.closed.Store(true)
	s.listener.Close() // could be nil actually if the package is misused but I assume that won't happen
}

func (s *Server) handle(conn net.Conn) {
	// Shut down the connection.
	defer conn.Close()

	n, _ := io.Copy(conn, strings.NewReader("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!\n"))
	// conn.Write([]byte(response)) another way of doing it

	fmt.Printf("received %d bytes\n", n)
}
