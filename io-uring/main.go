package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "192.168.64.25:8190")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Get user input from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Send user input to server
		_, err := conn.Write(scanner.Bytes())
		if err != nil {
			fmt.Println(err)
			return
		}

		// Read echo response from server
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Print server response
		fmt.Println("resp: " + string(buf[:n]))
	}
}
