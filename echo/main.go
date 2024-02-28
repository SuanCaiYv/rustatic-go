package main

import (
	"fmt"
	"net"
	"time"
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

	t := time.Now()
	for {
		n, _ := conn.Read(buf[0:])
		total += n
		fmt.Println("total:", total)
		if total == 48845931 {
			break
		}
	}
	fmt.Println("time:", time.Now().Sub(t).String())
}
