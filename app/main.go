package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

type entry struct {
	value      string
	expiryTime int64
}

var (
	store = make(map[string]entry)
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

func handleSet(conn net.Conn, key, value string, pxMillis int64) {
	mu.Lock()
	defer mu.Unlock()
	var exp int64
	if pxMillis > 0 {
		exp = time.Now().UnixMilli() + pxMillis
	} else {
		exp = 0 // No expiry
	}
	store[key] = entry{value: value, expiryTime: exp}
	conn.Write([]byte("+OK\r\n"))
}

func handleGet(conn net.Conn, key string) {
	mu.Lock()
	defer mu.Unlock()
	if entry, exists := store[key]; exists {
		if entry.expiryTime == 0 || entry.expiryTime > time.Now().UnixMilli() {
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(entry.value), entry.value)))
		} else {
			delete(store, key)
			conn.Write([]byte("$-1\r\n")) // Key has expired
		}
	} else {
		conn.Write([]byte("$-1\r\n")) // Key does not exist
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
				key := commandArray[i+2]
				value := commandArray[i+4]
				pxMillis := int64(0)
				if len(commandArray) > i+5 && (commandArray[i+5] == "PX" || commandArray[i+5] == "px") {
					pxMillis = int64(0) // Placeholder for actual expiry time logic
				}
				handleSet(conn, key, value, pxMillis)
			case "GET":
				handleGet(conn, commandArray[i+2])
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
