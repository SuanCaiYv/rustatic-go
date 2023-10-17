package main

import (
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "192.168.64.25:8190")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buffer := make([]byte, 1024*1024*32)
	curr := time.Now()
	total := 0
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		total += n
		time.Sleep(time.Millisecond * 10)
	}
	println("total:", total, "time:", time.Now().Sub(curr).String())
}
