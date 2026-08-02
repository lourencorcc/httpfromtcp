package main

import (
	"fmt"
	"httpgo/internal/request"
	"httpgo/internal/response"
	"httpgo/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handleRequest)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handleRequest(w *response.Writer, r *request.Request) {

	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		err := w.WriteStatusLine(response.BadRequest)
		if err != nil {
			log.Fatal(err)
		}

		err = w.WriteHeaders(response.GetDefaultHeaders(len(response.BadRequestHtmlMsg)))
		if err != nil {
			log.Fatal(err)
		}

		n, err := w.WriteBody([]byte(response.BadRequestHtmlMsg))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("responded with %d bytes on /yourproblem \n", n)
	case "/myproblem":
		err := w.WriteStatusLine(response.InternalErr)
		if err != nil {
			log.Fatal(err)
		}

		err = w.WriteHeaders(response.GetDefaultHeaders(len(response.InternalErrHtmlMsg)))
		if err != nil {
			log.Fatal(err)
		}

		n, err := w.WriteBody([]byte(response.InternalErrHtmlMsg))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("responded with %d bytes on /myproblem\n", n)

	default:
		err := w.WriteStatusLine(response.Ok)
		if err != nil {
			log.Fatal(err)
		}

		err = w.WriteHeaders(response.GetDefaultHeaders(len(response.OkHtmlMsg)))
		if err != nil {
			log.Fatal(err)
		}

		n, err := w.WriteBody([]byte(response.OkHtmlMsg))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("responded with %d bytes on %s\n", n, r.RequestLine.RequestTarget)
	}
}

func handleRequestOld(w io.Writer, r *request.Request) *server.HandlerError {
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.BadRequest,
			Msg:        "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.InternalErr,
			Msg:        "Woopsie, my bad\n",
		}
	default:
		w.Write([]byte("All good, frfr\n"))
		return &server.HandlerError{
			StatusCode: response.Ok,
		}
	}
}
