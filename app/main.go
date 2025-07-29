package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit
var (
	store = make(map[string]string)
	mu    sync.Mutex
)

func handlePing(conn net.Conn) {
	conn.Write([]byte("+PONG\r\n"))
}

//	func handleParser(commandString string) ([]string, error) {
//		// This function is a placeholder for any command parsing logic you might want to implement.
//		// For now, it simply returns the command string as is.
//		for _, char := range commandString {
//			if char
//		return nil, nil
//	}
func handleEcho(conn net.Conn, message string) {
	conn.Write([]byte(fmt.Sprintf("+%s\r\n", message)))
}

func handleSet(conn net.Conn, key, value string) {
	mu.Lock()
	store[key] = value
	mu.Unlock()
	conn.Write([]byte("+OK\r\n"))
}

func handleGet(conn net.Conn, key string) {
	mu.Lock()
	value, ok := store[key]
	mu.Unlock()
	if ok {
		conn.Write([]byte(fmt.Sprintf("+%s\r\n", value)))
	} else {
		conn.Write([]byte("-ERR key not found\r\n"))
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			os.Exit(1)
		}
		commandString := string(buffer[:n])
		commandArray := strings.Split(commandString, "\r\n")
		for i, command := range commandArray {

			switch command {
			case "PING":
				handlePing(conn)
			case "ECHO":
				handleEcho(conn, commandArray[i+2])
			case "SET":
				handleSet(conn, commandArray[i+3], commandArray[i+4])
			case "GET":
				handleGet(conn, commandArray[i+1])
			}
		}

	}

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handle(conn)
	}
}
