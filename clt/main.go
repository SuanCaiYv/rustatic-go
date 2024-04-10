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
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var (
	langFlag  int
	ip        string
	ctrlConn  net.Conn
	dataConn  net.Conn
	username  string
	password  string
	sessionId string
	fileMap   map[int]string
)

func main() {
	inputHandler()
	file, err := os.Open("rustatic.conf")
	if err == nil {
		data := make([]byte, 1024)
		n, _ := file.Read(data)
		arr := strings.Split(string(data[:n]), "\n")
		ip = arr[0]
		username = arr[1]
		password = arr[2]
	}
	fmt.Println("📄 Welcome to Rustatic! Tiny and fast file driver for personal using. 😎")
	fmt.Println("Choose you language: 🇨🇳 中文[1] 🇺🇸 English[2], default is English. 🌍. Input number directly/直接输入数字")
	fmt.Print("📝 ")
	var lang string
	fmt.Scanln(&lang)

	langFlag = 2
	if lang == "1" {
		langFlag = 1
	}

	if len(ip) != 0 && len(username) != 0 && len(password) != 0 {
		if langFlag == 1 {
			fmt.Println("🫨 看起来你已经保存了服务器地址和登录信息。你想使用它们吗？(Yes/No)")
		} else {
			fmt.Println("🫨 Seems that you have saved your server address and login information. Do you want to use them? (Yes/No)")
		}
		fmt.Print("📝 ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "Yes" {
			if langFlag == 1 {
				fmt.Println("🖥️ 请告诉我你的服务器地址，只需要IP即可。")
			} else {
				fmt.Println("🖥️ Please tell me your server address, just ip only.")
			}
			fmt.Scanln(&ip)
		}
	} else {
		if langFlag == 1 {
			fmt.Println("🖥️ 请告诉我你的服务器地址，只需要IP即可。🖥️")
		} else {
			fmt.Println("🖥️ Please tell me your server address, just ip only.🖥️")
		}
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
				if langFlag == 1 {
					fmt.Println("对不起，我无法连接到服务器。请检查你的服务器地址，确保端口 8190 和 8191 是开放的。🥺")
				} else {
					fmt.Println("I'm sorry, but I can't connect to the server. " +
						"Please check your server address and make sure the port 8190 and 8191 are open. 🥺")
				}
				os.Exit(1)
			}
			time.Sleep(2 * time.Second)
			if langFlag == 1 {
				fmt.Println("请稍等，我正在连接到我的服务器朋友。🤗")
			} else {
				fmt.Println("Please waiting, I'm connecting to my server friend. 🤗")
			}
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

	if langFlag == 1 {
		fmt.Printf("📝 所有支持的操作有:\n🛫 上传   [up]\n🚚 下载   [dl]\n📒 列表   [ls]\n" +
			"🪦 删除   [de]\n🚁 登录   [lg]\n🚀 注册   [sg]\n🤙 记住我 [re]\n")
	} else {
		fmt.Printf("📝 All supported operations are:\n🛫 upload     [up]\n🚚 download   [dl]\n📒 list       [ls]\n" +
			"🪦 delete     [de]\n🚁 login      [lg]\n🚀 sign       [sg]\n🤙 remeber me [re]\n")
	}
	fmt.Println("😘")
	if langFlag == 1 {
		fmt.Println("📝 请输入你的操作类型，如上传 'up' 或 'upload'。")
	} else {
		fmt.Println("📝 Please input your operation type, such as upload with 'up' or 'upload'.")
	}
	fmt.Print("📝 ")
	var op string
	for {
		fmt.Scanln(&op)
		switch op {
		case "up", "upload":
			if langFlag == 1 {
				fmt.Println("上传文件")
			} else {
				fmt.Println("You are uploading a file.")
			}
			up()
		case "dl", "download":
			if langFlag == 1 {
				fmt.Println("下载文件")
			} else {
				fmt.Println("You are downloading a file.")
			}
			dl()
		case "ls", "list":
			if langFlag == 1 {
				fmt.Println("查看文件列表")
			} else {
				fmt.Println("You are listing files.")
			}
			ls()
		case "de", "delete":
			if langFlag == 1 {
				fmt.Println("删除文件")
			} else {
				fmt.Println("You are deleting a file.")
			}
			fmt.Println("This operation is not supported yet.")
		case "lg", "login":
			if langFlag == 1 {
				fmt.Println("登录中... ...")
			} else {
				fmt.Println("You are logging in.")
			}
			lg()
		case "sg", "sign":
			if langFlag == 1 {
				fmt.Println("哇哦🎉新来的！")
			} else {
				fmt.Println("You are new here!")
			}
			sg()
		case "re", "remember me":
			if langFlag == 1 {
				fmt.Println("记住账号？")
			} else {
				fmt.Println("You are remembering me.")
			}
			re()
		case "exit":
			fmt.Println("Goodbye! 🥳")
			return
		default:
			if langFlag == 1 {
				fmt.Println("无效的操作类型，请重新输入你的操作类型。")
			} else {
				fmt.Println("Invalid operation type. Please input your operation type again.")
			}
			continue
		}
		if langFlag == 1 {
			fmt.Println("继续吗？😋")
		} else {
			fmt.Println("Let's continue! 😋")
		}
		fmt.Print("👉 ")
	}
}

func inputHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Printf("\nGoodbye! 🥳")
		os.Exit(0)
	}()
}

func lg() {
	if langFlag == 1 {
		fmt.Println("请输入你的用户名。")
	} else {
		fmt.Println("Please input your username.")
	}
	fmt.Print("📝 ")
	fmt.Scanln(&username)
	if langFlag == 1 {
		fmt.Println("请输入你的密码。")
	} else {
		fmt.Println("Please input your password.")
	}
	fmt.Print("📝 ")
	fmt.Scanln(&password)
	var err error
	sessionId, err = login(username, password, ctrlConn)
	if err != nil {
		fmt.Println(err)
		return
	}
	initDataConn(sessionId, dataConn)
	fmt.Println("Login successfully.🎮")
}

func sg() {
	fmt.Println("Please input your username.")
	fmt.Print("📝 ")
	fmt.Scanln(&username)
	fmt.Println("Please input your password.")
	fmt.Print("📝 ")
	var password string
	fmt.Scanln(&password)
	var err error
	sessionId, err = sign(username, password, ctrlConn)
	if err != nil {
		fmt.Println(err)
		return
	}
	initDataConn(sessionId, dataConn)
	fmt.Println("Sign up successfully. (automatically login!) 🥳")
}

func up() {
	if len(sessionId) == 0 {
		fmt.Println("Please login first.")
		return
	}
	fmt.Println("Please input your file path.")
	fmt.Print("📝 ")
	var filepath string
	fmt.Scanln(&filepath)
	if filepath[0] == '"' && filepath[len(filepath)-1] == '"' {
		filepath = filepath[1 : len(filepath)-1]
	}
	fileSize, err := upload(filepath)
	if err != nil {
		return
	}
	upload0(filepath, fileSize)
	fmt.Println("Upload finished. 🛫")
}

func dl() {
	if len(sessionId) == 0 {
		fmt.Println("Please login first.")
		return
	}
	fmt.Println("Please input your target file index. Such as 123, 234 etc...")
	fmt.Print("📝 ")
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
	fmt.Println("File total size:", size)
	t := time.Now()
	download0(filename, dataConn, size)
	fmt.Printf("Download finished 🚚, speed: %.4f MB/s\n", float64(size)/1024/1024/time.Now().Sub(t).Seconds())
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
		tag, _ := strconv.Atoi(file[2])
		if tag == 0 {
			fmt.Printf("File index: %3d, size: %s, filename: %s\n", i, formatSize(size), file[0])
		} else {
			fmt.Printf("File index: %3d, size: %s, filename: %s-[%d]\n", i, formatSize(size), file[0], tag)
		}
		fileMap[i] = file[5]
	}
}

func re() {
	fmt.Println("Please input 'Yes' to confirm.")
	fmt.Print("📝 ")
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
		fmt.Println("File not found. 🙇")
		return 0, err
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
	d := make(chan struct{}, 1)
	go func() {
		t := time.Tick(time.Second)
		for {
			select {
			case <-t:
				fmt.Printf("\rupload percentage: %6.2f%%", float64(total)/float64(size)*100)
			case <-d:
				fmt.Println()
				<-d
				return
			}
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
	if total != size {
		fmt.Println("File size fatal, total:", total, "size:", size)
	}
	d <- struct{}{}
	d <- struct{}{}
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
	d := make(chan struct{}, 1)
	go func() {
		t := time.Tick(time.Second)
		for {
			select {
			case <-t:
				fmt.Printf("\rdownload percentage: %6.2f%%", float64(total)/float64(size)*100)
			case <-d:
				fmt.Println()
				<-d
				return
			}
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
	d <- struct{}{}
	d <- struct{}{}
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
		for i := 0; i < len(ll)-1; i += 7 {
			arr := make([]string, 7)
			for j := 0; j < 7; j++ {
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
