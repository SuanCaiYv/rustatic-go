package benchmark

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
)

func normal() {
	dataConn, err := net.Dial("tcp", "127.0.0.1:8192")
	if err != nil {
		panic(fmt.Sprintf("data server connect error: %s", err))
	}
	defer dataConn.Close()
	buffer := make([]byte, 1024*1024*8)
	size := 705299015
	total := 0
	// t := time.Now()
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
	// fmt.Printf("speed: %.4f MB/s\n", float64(size)/1024/1024/time.Now().Sub(t).Seconds())
}

func rustatic() {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	ctrlConn, err := tls.Dial("tcp", "127.0.0.1:8190", config)
	if err != nil {
		panic(fmt.Sprintf("control server connect error: %s", err))
	}
	dataConn, err := net.Dial("tcp", "127.0.0.1:8191")
	if err != nil {
		panic(fmt.Sprintf("data server connect error: %s", err))
	}
	defer ctrlConn.Close()
	defer dataConn.Close()
	sessionId, err := login("dev-user", "123456", ctrlConn)
	if err != nil {
		panic(err)
	}
	initDataConn(sessionId, dataConn)
	//filepath := "/Users/joker/Downloads/DxO_PureRaw_v3.dmg"
	//fileId, err := upload(sessionId, filepath, ctrlConn, dataConn)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(fileId)
	//upload0(filepath, dataConn)
	filename, size, err := download(sessionId, "M2M2M2FhM2UtZDNjOC00ZDAxLTkyOWEtM2VkMWYwYTRhNWI0", ctrlConn)
	if err != nil {
		panic(err)
	}
	// t := time.Now()
	download1(filename, dataConn, size)
	// fmt.Printf("speed: %.4f MB/s\n", float64(size)/1024/1024/time.Now().Sub(t).Seconds())
}

func initDataConn(sessionId string, conn net.Conn) {
	conn.Write([]byte(sessionId + "\n"))
}

func login(username string, password string, conn net.Conn) (string, error) {
	user := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}
	binary.Write(conn, binary.BigEndian, uint16(2))
	req := jsonMarshal(user)
	binary.Write(conn, binary.BigEndian, uint16(len(req)))
	conn.Write(req)
	resp := readToLine(conn)
	str := string(resp)
	if strings.HasPrefix(str, "err ") {
		return "", fmt.Errorf(string(resp[4:]))
	} else {
		return str[3:], nil
	}
}

func download(sessionId string, fileId string, ctrlConn net.Conn) (string, int, error) {
	file := struct {
		SessionId string `json:"session_id"`
		FileId    string `json:"link"`
	}{
		SessionId: sessionId,
		FileId:    fileId,
	}
	binary.Write(ctrlConn, binary.BigEndian, uint16(4))
	req := jsonMarshal(file)
	binary.Write(ctrlConn, binary.BigEndian, uint16(len(req)))
	ctrlConn.Write(req)
	resp := readToLine(ctrlConn)
	str := string(resp)
	if strings.HasPrefix(str, "err ") {
		return "", 0, fmt.Errorf(string(resp[4:]))
	} else {
		arr := strings.Split(str[3:], " ")
		size, _ := strconv.Atoi(arr[1])
		return arr[0], size, nil
	}
}

func download1(_filename string, dataConn net.Conn, size int) {
	buffer := make([]byte, 1024*1024*4)
	total := 0
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
}

func jsonMarshal(data any) (output []byte) {
	output, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return
}

func readToLine(reader io.Reader) []byte {
	buffer := make([]byte, 4096)
	idx := 0
	for {
		n, err := reader.Read(buffer[idx:])
		if err != nil {
			panic(err)
		}
		if n == 0 {
			return nil
		}
		for i := idx; i < idx+n; i += 1 {
			if buffer[i] == '\n' {
				return buffer[:i]
			}
		}
		idx += n
	}
}

func BenchmarkNormal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		normal()
	}
}

func BenchmarkRustatic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rustatic()
	}
}

// go test -bench=. -benchtime=3s -count=3 .
