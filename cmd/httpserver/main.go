package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpgo/internal/headers"
	"httpgo/internal/request"
	"httpgo/internal/response"
	"httpgo/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/stream/") {
		// Chunked encoding
		proxyChunked(w, r)
		return
	}
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/range/") {
		// Trailer headers
		proxyTrailers(w, r)
		return
	}

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

func proxyChunked(w *response.Writer, r *request.Request) {
	arg, err := strconv.Atoi(strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/stream/"))
	if err != nil {
		log.Fatal(err)
	}

	err = w.WriteStatusLine(response.Ok)
	if err != nil {
		log.Fatal(err)
	}

	// headers
	headers := response.GetDefaultHeaders(len(response.OkHtmlMsg))
	err = headers.Override("Content-Length", "Transfer-Encoding", "chunked")
	if err != nil {
		log.Fatal(err)
	}
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Fatal(err)
	}

	// CHunked body
	buf := make([]byte, 1024)
	res, err := http.Get(fmt.Sprintf("https://httpbingo.org/stream/%d", arg))
	if err != nil {
		log.Fatal(err)
	}

	var errRead error
	var nRead int
	defer res.Body.Close()
	for errRead != io.EOF {
		nRead, errRead = res.Body.Read(buf)
		if errRead != nil {
			if errRead != io.EOF {
				log.Fatal(errRead)
			}
		}
		if nRead > 0 {
			fmt.Printf("Got %d bytes in response from httpbin\n", nRead)
			n, err := w.WriteChunkedBody(buf[:nRead])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Forwarded %d chunked bytes\n", n) // should be +2 due to rn
		}
		if errRead == io.EOF {
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Fatal(err)
	}
}

func proxyTrailers(w *response.Writer, r *request.Request) {
	arg, err := strconv.Atoi(strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/range/"))
	if err != nil {
		log.Fatal(err)
	}

	err = w.WriteStatusLine(response.Ok)
	if err != nil {
		log.Fatal(err)
	}

	// headers
	myHeaders := response.GetDefaultHeaders(len(response.OkHtmlMsg))
	err = myHeaders.Override("Content-Length", "Transfer-Encoding", "chunked")
	if err != nil {
		log.Fatal(err)
	}

	myHeaders.Add("Trailer", "X-Content-SHA256, X-Content-Length")

	err = w.WriteHeaders(myHeaders)
	if err != nil {
		log.Fatal(err)
	}

	hasher := sha256.New()
	// CHunked body
	buf := make([]byte, 1024)
	res, err := http.Get(fmt.Sprintf("https://httpbingo.org/range/%d", arg))
	if err != nil {
		log.Fatal(err)
	}

	var errRead error
	var nRead int
	var totalRead int
	defer res.Body.Close()
	for errRead != io.EOF {
		nRead, errRead = res.Body.Read(buf)
		if errRead != nil {
			if errRead != io.EOF {
				log.Fatal(errRead)
			}
		}
		if nRead > 0 {
			fmt.Printf("Got %d bytes in response from httpbin\n", nRead)
			hasher.Write(buf[:nRead])
			totalRead += nRead
			n, err := w.WriteChunkedBody(buf[:nRead])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Forwarded %d chunked bytes\n", n) // should be +2 due to rn
		}
		if errRead == io.EOF {
			break
		}
	}
	hash := hasher.Sum(nil)
	fmt.Println(hash)
	trailers := headers.NewHeaders()
	trailers.Add("X-Content-SHA256", hex.EncodeToString(hash))
	trailers.Add("X-Content-Length", strconv.Itoa(totalRead))

	err = w.WriteTrailers(trailers)
	if err != nil {
		log.Fatal(err)
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
