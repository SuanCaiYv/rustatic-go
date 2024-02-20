package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024*1024*32)
	total := 0

	for {
		n, _ := conn.Read(buf)
		total += n
		if n == 0 {
			break
		}
		fmt.Println("total:", total)
	}
}
