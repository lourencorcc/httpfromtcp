package main

import (
	"fmt"
	"log"
	"net"

	"httpgo/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		rl, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("error parsing request: %v\n", err)
		} else {
			fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", rl.RequestLine.Method, rl.RequestLine.RequestTarget, rl.RequestLine.HttpVersion)
			fmt.Printf("Headers:\n")
			for k, v := range rl.Headers {
				fmt.Printf("- %s: %s\n", k, v)
			}
			fmt.Printf("Body:\n%s\n", string(rl.Body))
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
