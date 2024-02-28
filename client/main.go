package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
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
	//filepath := "/Users/slma/Downloads/dotnet-runtime-7.0.13-osx-arm64.pkg"
	//fileId, err := upload(sessionId, filepath, ctrlConn, dataConn)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(fileId)
	//upload0(filepath, dataConn)
	filename, size, err := download(sessionId, "OTcxNjU3MTgtOWEzZS00YWQ5LWE1ZjctNzMwNGVkMjNlMjM1", ctrlConn)
	if err != nil {
		panic(err)
	}
	fmt.Println(size)
	t := time.Now()
	download1(filename, dataConn, size)
	fmt.Printf("speed: %.4f MB/s\n", float64(size)/1024/1024/time.Now().Sub(t).Seconds())
}

func initDataConn(sessionId string, conn net.Conn) {
	conn.Write([]byte(sessionId + "\n"))
}

func sign(username string, password string, conn net.Conn) (string, error) {
	user := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}
	binary.Write(conn, binary.BigEndian, uint16(1))
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

func upload(sessionId string, filepath string, ctrlConn net.Conn, dataConn net.Conn) (string, error) {
	fileStat, err := os.Stat(filepath)
	if err != nil {
		panic(err)
	}
	fileSize := fileStat.Size()
	fileName := fileStat.Name()
	createAt := fileStat.ModTime().UnixMilli()
	file := struct {
		SessionId string `json:"session_id"`
		FileName  string `json:"filename"`
		FileSize  int64  `json:"size"`
		CreateAt  int64  `json:"create_at"`
		UpdateAt  int64  `json:"update_at"`
	}{
		SessionId: sessionId,
		FileName:  fileName,
		FileSize:  fileSize,
		CreateAt:  createAt,
		UpdateAt:  createAt,
	}
	binary.Write(ctrlConn, binary.BigEndian, uint16(3))
	req := jsonMarshal(file)
	binary.Write(ctrlConn, binary.BigEndian, uint16(len(req)))
	ctrlConn.Write(req)
	resp := readToLine(ctrlConn)
	str := string(resp)
	if strings.HasPrefix(str, "err ") {
		return "", fmt.Errorf(string(resp[4:]))
	} else {
		return str[3:], nil
	}
}

func upload0(filepath string, dataConn net.Conn) {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buffer := make([]byte, 4096)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		if n == 0 {
			break
		}
		dataConn.Write(buffer[:n])
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
	binary.Write(ctrlConn, binary.BigEndian, uint16(5))
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

func download0(filename string, dataConn net.Conn, size int) {
	file, err := os.OpenFile("./"+filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buffer := make([]byte, 4096*32)
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
		// fmt.Println(n, total)
		file.Write(buffer[:n])
		if total == size {
			break
		}
	}
}

func download1(_filename string, dataConn net.Conn, size int) {
	buffer := make([]byte, 1024*1024*4)
	total := 0
	go func() {
		for {
			if total == size {
				break
			}
			time.Sleep(time.Second)
			fmt.Printf("process percentage: %6.2f%%\n", float64(total)/float64(size)*100)
		}
	}()
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

func listFiles(username string, ctrlConn net.Conn) ([]map[string]string, error) {
	req := []byte(username)
	binary.Write(ctrlConn, binary.BigEndian, uint16(6))
	binary.Write(ctrlConn, binary.BigEndian, uint16(len(req)))
	ctrlConn.Write(req)
	resp := readToLine(ctrlConn)
	str := string(resp[0:3])
	if strings.HasPrefix(str, "ok ") {
		var files []map[string]string
		ll := bytes.Split(resp[3:], []byte(" "))
		for i := 0; i < len(ll); i += 6 {
			for j := 0; j < 6; j++ {
				files = append(files, map[string]string{string(ll[i+j]): string(ll[i+j+1])})
			}
		}
		return files, nil
	} else {
		return nil, fmt.Errorf(string(resp[4:]))
	}
}
