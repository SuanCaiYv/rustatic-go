package main

import (
	"os"
	"time"
)

func main() {
	file, err := os.Open("/Users/slma/Downloads/Apifox-macOS-arm64-latest.zip")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buffer := make([]byte, 1024*1024*32)
	total := 0
	t := time.Now()
	for {
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		total += n
	}
	println("total:", total, "time:", time.Now().Sub(t).String())
}
