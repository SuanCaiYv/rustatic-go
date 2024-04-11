package main

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
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
		size := file[1].(int)
		tag := file[2].(int)
		if tag == 0 {
			fmt.Printf("File index: %3d, size: %s, filename: %s\n", i, formatSize(size), file[0])
		} else {
			fmt.Printf("File index: %3d, size: %s, filename: %s-[%d]\n", i, formatSize(size), file[0], tag)
		}
		fileMap[i] = file[6].(string)
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

func login(username string, password string, conn net.Conn) (sessionId string, err error) {
	item1 := []byte(username)
	item2 := []byte(password)
	writeReq(conn, 2, item1, item2)
	resp, err := readResp(conn)
	if err != nil {
		return
	}
	sessionId = string(resp[0])
	return
}

func sign(username string, password string, conn net.Conn) (sessionId string, err error) {
	item1 := []byte(username)
	item2 := []byte(password)
	writeReq(conn, 1, item1, item2)
	resp, err := readResp(conn)
	if err != nil {
		return
	}
	sessionId = string(resp[0])
	return
}

func upload(filepath string) (fileSize int, err error) {
	fileStat, err := os.Stat(filepath)
	if err != nil {
		fmt.Println("File not found. 🙇")
		return 0, err
	}
	fileSize = int(fileStat.Size())
	fileName := fileStat.Name()
	createAt := fileStat.ModTime().UnixMilli()
	item1 := []byte(sessionId)
	item2 := []byte(fileName)
	item3 := make([]byte, 8)
	binary.BigEndian.PutUint64(item3, uint64(fileSize))
	item4 := make([]byte, 8)
	binary.BigEndian.PutUint64(item4, uint64(createAt))
	item5 := make([]byte, 8)
	binary.BigEndian.PutUint64(item5, uint64(createAt))
	writeReq(ctrlConn, 3, item1, item2, item3, item4, item5)
	_, err = readResp(ctrlConn)
	if err != nil {
		return 0, err
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

func download(fileId string) (filename string, size int, err error) {
	item1 := []byte(sessionId)
	item2 := []byte(fileId)
	writeReq(ctrlConn, 4, item1, item2)
	resp, err := readResp(ctrlConn)
	if err != nil {
		return
	} else {
		filename = string(resp[0])
		size = int(binary.BigEndian.Uint64(resp[1]))
		return
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

func listFiles() (files [][]any, err error) {
	item1 := []byte(username)
	writeReq(ctrlConn, 6, item1)
	resp, err := readResp(ctrlConn)
	if err != nil {
		return
	}
	for i := 0; i < len(resp); i += 7 {
		arr := make([]any, 7)
		arr[0] = string(resp[i])
		for j := 1; j < 6; j++ {
			arr[j] = int(binary.BigEndian.Uint64(resp[i+j]))
		}
		arr[6] = string(resp[i+6])
		files = append(files, arr)
	}
	return
}

func initDataConn(sessionId string, conn net.Conn) {
	conn.Write([]byte(sessionId + "\n"))
}

func writeAll(writer io.Writer, data []byte) {
	idx := 0
	l := len(data)
	for {
		n, err := writer.Write(data[idx:])
		if err != nil {
			panic(err)
		}
		idx += n
		if idx == l {
			return
		}
	}
}

func writeReq(writer io.Writer, opCode int, req ...[]byte) {
	binary.Write(writer, binary.BigEndian, uint16(opCode))
	l := 0
	for i := range req {
		l += 2
		l += len(req[i])
	}
	binary.Write(writer, binary.BigEndian, uint16(l))
	for i := range req {
		binary.Write(writer, binary.BigEndian, uint16(len(req[i])))
		writeAll(writer, req[i])
	}
}

func readAll(reader io.Reader, size int) []byte {
	buffer := make([]byte, size)
	idx := 0
	for {
		n, err := reader.Read(buffer[idx:])
		if err != nil {
			panic(err)
		}
		idx += n
		if idx == size {
			return buffer
		}
	}
}

// readResp/ on ok: [ok data1 data2 data3 ...], on err: [err errorStr]
func readResp(reader io.Reader) (resp [][]byte, err error) {
	l := uint16(0)
	binary.Read(reader, binary.BigEndian, &l)
	if l == 0 {
		return
	}
	for {
		size := uint16(0)
		binary.Read(reader, binary.BigEndian, &size)
		resp = append(resp, readAll(reader, int(size)))
		l -= 2
		l -= size
		if l == 0 {
			break
		}
	}
	if string(resp[0]) == "err" {
		resp = make([][]byte, 0)
		err = fmt.Errorf(string(resp[1]))
	} else {
		resp = resp[1:]
	}
	return
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
