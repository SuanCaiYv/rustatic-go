package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	dataConn, err := net.Dial("tcp", "127.0.0.1:8192")
	if err != nil {
		panic(fmt.Sprintf("data server connect error: %s", err))
	}
	defer dataConn.Close()
	buffer := make([]byte, 1024*1024*8)
	size := 705299015
	total := 0
	t := time.Now()
	for {
		n, err := dataConn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		total += n
		if total == size {
			break
		}
	}
	if total != size {
		println("total:", total, "size:", size)
	}
	fmt.Printf("speed: %.4f MB/s\n", float64(size)/1024/1024/time.Now().Sub(t).Seconds())
}
