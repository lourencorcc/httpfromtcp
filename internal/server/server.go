package server

import (
	"bytes"
	"fmt"
	"httpgo/internal/request"
	"httpgo/internal/response"
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
	listener    net.Listener
	closed      atomic.Bool
	handlerFunc Handler
}

type Handler func(w io.Writer, r *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

func (hErr *HandlerError) writeErr(w io.Writer) error {

	headers := response.GetDefaultHeaders(len(hErr.Msg))

	err := response.WriteStatusLine(w, hErr.StatusCode)
	if err != nil {
		return err
	}

	err = response.WriteHeaders(w, headers)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, strings.NewReader(hErr.Msg))
	if err != nil {
		return err
	}

	return nil
}

func Serve(port int, handler Handler) (*Server, error) {
	// creates a net.Listener and returns a new Server instance.
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("error while creating listener on port %d: %w", port, err)
	}

	server := &Server{
		listener:    l, // fine cz closed gets the 0 value because we mustn't copy an atomic bool
		handlerFunc: handler,
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
			s.handle(c) // not sure what to do with the errors
		}(conn)

	}
}

func (s *Server) Close() {
	s.closed.Store(true)
	s.listener.Close() // could be nil actually if the package is misused but I assume that won't happen
}

func (s *Server) handle(conn net.Conn) error {
	// Shut down the connection.
	defer conn.Close()
	parsedRequest, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.BadRequest,
			Msg:        err.Error(),
		}
		hErr.writeErr(conn)
		return err
	}

	var b bytes.Buffer

	hErr := s.handlerFunc(&b, parsedRequest)
	if hErr.StatusCode != response.Ok {
		hErr.writeErr(conn)
		return nil
	}

	headers := response.GetDefaultHeaders(b.Len())

	err = response.WriteStatusLine(conn, response.Ok)
	if err != nil {
		return err
	}

	response.WriteHeaders(conn, headers)
	_, err = io.Copy(conn, &b) // TODO: vs conn.Write ? what is the diff
	if err != nil {
		return err
	}

	return nil

}
