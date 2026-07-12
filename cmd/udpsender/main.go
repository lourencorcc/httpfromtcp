package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		panic(err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("UDP connection successful on port %d\n", udpAddr.Port)

	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := reader.ReadString(10)
		if err != nil {
			fmt.Printf("Error while reading from stdin: %v\n", err)
		}

		n, err := udpConn.Write([]byte(line))
		if err != nil {
			log.Printf("Error writing to UDP: %v", err)
		}
		fmt.Printf("Sent %d bytes to server\n", n)

	}

}
