package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == "PING" {
			conn.Write([]byte("+PONG\r\n"))
		} else if strings.Contains(strings.ToUpper(text), "ECHO") {
			// Extract the message after ECHO

			lines := strings.Split(text, "\n")
			var message string
			echoFound := false

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if echoFound && line != "" && !strings.HasPrefix(line, "$") {
					message = line
					break
				}
				if strings.ToUpper(line) == "ECHO" {
					echoFound = true
				}
			}

			if message != "" {
				response := fmt.Sprintf("$%d\r\n%s\r\n", len(message), message)
				conn.Write([]byte(response))
			}
		}

	}
}
