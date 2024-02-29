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

var (
	ip        string
	ctrlConn  net.Conn
	dataConn  net.Conn
	username  string
	password  string
	sessionId string
	fileMap   map[int]string
)

func main() {
	file, err := os.Open("rustatic.conf")
	if err == nil {
		data := make([]byte, 1024)
		n, _ := file.Read(data)
		arr := strings.Split(string(data[:n]), "\n")
		ip = arr[0]
		username = arr[1]
		password = arr[2]
	}
	fmt.Println("📄Welcome to Rustatic! Tiny and fast file driver for personal using.😎")

	if len(ip) != 0 && len(username) != 0 && len(password) != 0 {
		fmt.Println("🫨Seems that you have saved your server address and login information. Do you want to use them? (Yes/No)")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "Yes" {
			fmt.Println("Please tell me your server address, just ip only.🖥️")
			fmt.Scanln(&ip)
		}
	} else {
		fmt.Println("Please tell me your server address, just ip only.🖥️")
		fmt.Scanln(&ip)
	}

	status := 0
	retry := 0
	wait := make(chan struct{})
	go func() {
		for {
			if status == 1 {
				close(wait)
				break
			}
			if retry > 5 {
				fmt.Println("I'm sorry, but I can't connect to the server. " +
					"Please check your server address and make sure the port 8190 and 8191 are open.🥺")
				os.Exit(1)
			}
			time.Sleep(2 * time.Second)
			fmt.Println("Please waiting, I'm connecting to my server friend.🤗")
			retry += 1
		}
	}()
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	ctrlConn, err = tls.Dial("tcp", ip+":8190", config)
	if err != nil {
		fmt.Println(fmt.Sprintf("control stream connect error: %s", err))
		return
	}
	dataConn, err = net.Dial("tcp", ip+":8191")
	if err != nil {
		fmt.Println(fmt.Sprintf("data stream connect error: %s", err))
		return
	}
	defer ctrlConn.Close()
	defer dataConn.Close()
	status = 1
	<-wait

	if len(username) != 0 && len(password) != 0 {
		sessionId, err = login(username, password, ctrlConn)
		if err != nil {
			fmt.Println(err)
			return
		}
		initDataConn(sessionId, dataConn)
		fmt.Println("Login successfully.")
	}

	fmt.Printf("All supported operations are:\n✈️upload   [up]\n🚚download [dl]\n⛰️list     [ls]\n" +
		"🗑️delete   [de]\n🎮login    [lg]\n🧑‍💻sign     [sg]\n")
	fmt.Println("😘")
	fmt.Println("Please input your operation type, such as upload with 'up' or 'upload'.🖇️")
	var op string
	for {
		fmt.Scanln(&op)
		switch op {
		case "up", "upload":
			fmt.Println("You are uploading a file.")
			up()
		case "dl", "download":
			fmt.Println("You are downloading a file.")
			dl()
		case "ls", "list":
			fmt.Println("You are listing files.")
			ls()
		case "de", "delete":
			fmt.Println("You are deleting a file.")
			fmt.Println("This operation is not supported yet.")
		case "lg", "login":
			fmt.Println("You are logging in.")
			lg()
		case "sg", "sign":
			fmt.Println("You are new here!")
			sg()
		case "re", "remember me":
			fmt.Println("You are remembering me.")
			re()
		default:
			fmt.Println("Invalid operation type. Please input your operation type again.")
			continue
		}
	}
}

func lg() {
	fmt.Println("Please input your username.")
	fmt.Scanln(&username)
	fmt.Println("Please input your password.")
	fmt.Scanln(&password)
	sessionId, err := login(username, password, ctrlConn)
	if err != nil {
		fmt.Println(err)
		return
	}
	initDataConn(sessionId, dataConn)
	fmt.Println("Login successfully.🎮")
}

func sg() {
	fmt.Println("Please input your username.")
	fmt.Scanln(&username)
	fmt.Println("Please input your password.")
	var password string
	fmt.Scanln(&password)
	sessionId, err := sign(username, password, ctrlConn)
	if err != nil {
		fmt.Println(err)
		return
	}
	initDataConn(sessionId, dataConn)
	fmt.Println("Sign up successfully. (automatically login!)🥳")
}

func up() {
	if len(sessionId) == 0 {
		fmt.Println("Please login first.")
		return
	}
	fmt.Println("Please input your file path.")
	var filepath string
	fmt.Scanln(&filepath)
	fileSize, err := upload(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}
	upload0(filepath, fileSize)
	fmt.Println("Upload finished.✈️")
}

func dl() {
	if len(sessionId) == 0 {
		fmt.Println("Please login first.")
		return
	}
	fmt.Println("Please input your target file index. Such as 123, 234 etc...")
	var fileId int
	fmt.Scanln(&fileId)
	if _, ok := fileMap[fileId]; !ok {
		fmt.Println("File index not found.")
	}
	filename, size, err := download(fileMap[fileId])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(size)
	t := time.Now()
	download0(filename, dataConn, size)
	fmt.Printf("Download finished🚚, speed: %.4f MB/s\n", float64(size)/1024/1024/time.Now().Sub(t).Seconds())
}

func ls() {
	if len(username) == 0 {
		fmt.Println("Please login first.")
		return
	}
	files, err := listFiles()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileMap = make(map[int]string)
	for i, file := range files {
		size, _ := strconv.Atoi(file[1])
		fmt.Printf("File index: %3d, size: %s, filename: %s\n", i, formatSize(size), file[0])
		fileMap[i] = file[5]
	}
}

func re() {
	fmt.Println("Please input 'Yes' to confirm.")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "Yes" {
		fmt.Println("Operation canceled.")
		return
	}
	file, _ := os.Create("rustatic.conf")
	file.WriteString(ip + "\n" + username + "\n" + password)
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

func upload(filepath string) (int, error) {
	fileStat, err := os.Stat(filepath)
	if err != nil {
		panic(err)
	}
	fileSize := int(fileStat.Size())
	fileName := fileStat.Name()
	createAt := fileStat.ModTime().UnixMilli()
	file := struct {
		SessionId string `json:"session_id"`
		FileName  string `json:"filename"`
		FileSize  int    `json:"size"`
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
		return 0, fmt.Errorf(string(resp[4:]))
	} else {
		return fileSize, nil
	}
}

func upload0(filepath string, size int) {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buffer := make([]byte, 4096)
	total := 0
	go func() {
		for {
			time.Sleep(time.Second)
			if total == size {
				break
			}
			fmt.Printf("upload percentage: %6.2f%%\n", float64(total)/float64(size)*100)
		}
	}()
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		total += n
		if n == 0 {
			break
		}
		dataConn.Write(buffer[:n])
	}
}

func download(fileId string) (string, int, error) {
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

func download0(filename string, dataConn net.Conn, size int) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	buffer := make([]byte, 1024*1024*4)
	total := 0
	go func() {
		for {
			time.Sleep(time.Second)
			if total == size {
				break
			}
			fmt.Printf("download percentage: %6.2f%%\n", float64(total)/float64(size)*100)
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
		file.Write(buffer[:n])
		total += n
		if total == size {
			break
		}
	}
	if total != size {
		fmt.Println("File size fatal, total:", total, "size:", size)
	}
}

func listFiles() ([][]string, error) {
	req := []byte(username)
	binary.Write(ctrlConn, binary.BigEndian, uint16(6))
	binary.Write(ctrlConn, binary.BigEndian, uint16(len(req)))
	ctrlConn.Write(req)
	resp := readToLine(ctrlConn)
	str := string(resp[0:3])
	if strings.HasPrefix(str, "ok ") {
		var files [][]string
		ll := bytes.Split(resp[3:], []byte(" "))
		for i := 0; i < len(ll)-1; i += 6 {
			arr := make([]string, 6)
			for j := 0; j < 6; j++ {
				arr[j] = string(ll[i+j])
			}
			files = append(files, arr)
		}
		return files, nil
	} else {
		return nil, fmt.Errorf(string(resp[4:]))
	}
}

func initDataConn(sessionId string, conn net.Conn) {
	conn.Write([]byte(sessionId + "\n"))
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

func formatSize(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%6.2d  B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%6.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%6.2f MB", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%6.2f GB", float64(size)/1024/1024/1024)
	}
}
