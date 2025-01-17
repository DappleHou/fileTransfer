package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

func main() {
	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs) // Root path now serves static files

	// Upload handler
	http.HandleFunc("/upload", uploadHandler)

	fmt.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

	//send()
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // Limit the max memory to 10MB
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Get the IP address
	ip := r.FormValue("ip")
	if ip == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Respond with success message
	response := fmt.Sprintf("File '%s' uploaded successfully with IP '%s'", header.Filename, ip)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, response)))
}

func send() {
	// 连接到接收端
	var ip string
	fmt.Print("Please enter the receiver's IP address: ")
	fmt.Scanln(&ip)
	conn, err := net.Dial("tcp", ip+":8888")
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer conn.Close()

	// 输入文件路径
	fmt.Print("Please enter the file path to send: ")
	var filePath string
	fmt.Scanln(&filePath)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("无法获取文件信息:", err)
		return
	}
	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	// 发送文件名和文件大小
	header := fmt.Sprintf("%s:%d\n", fileName, fileSize)
	_, err = conn.Write([]byte(header))
	if err != nil {
		fmt.Println("发送文件头信息失败:", err)
		return
	}

	// 发送文件内容
	buffer := make([]byte, 4096) // 4KB 缓冲区
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF { // 文件发送完成
				break
			}
			fmt.Println("读取文件内容失败:", err)
			return
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Println("发送文件内容失败:", err)
			return
		}
	}

	fmt.Println("文件发送完成!")
}
