package server

import (
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

// type Handler func(w io.Writer, r *request.Request) *HandlerError
type Handler func(w *response.Writer, req *request.Request)

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
				log.Fatalf("unexpected error while listening")
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
	w := &response.Writer{
		NextWrite:    response.StatusLine,
		ResponseType: "text/html",
		Writer:       conn,
	}

	parsedRequest, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.BadRequest)
		headers := response.GetDefaultHeaders(len(response.BadRequestHtmlMsg))
		headers.Set("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(response.BadRequestHtmlMsg))
		return err
	}

	s.handlerFunc(w, parsedRequest)
	return nil
}
