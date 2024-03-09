package main

import (
	"fmt"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", "0.0.0.0:8193")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer ln.Close()

	fmt.Println("Server is listening on port ", ln.Addr())

	for {
		// Accept new connection
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 128)

	for {
		total := 0
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				return
			}
			total += n
			if total == 64 {
				break
			}
		}

		for {
			n, err := conn.Write(buffer[:total])
			if err != nil {
				return
			}
			total -= n
			if total == 0 {
				break
			}
		}
	}
}
