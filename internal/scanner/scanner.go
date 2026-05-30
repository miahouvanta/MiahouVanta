package scanner

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Host struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Ports    []Port `json:"ports"`
	Status   string `json:"status"`
}

type Port struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	State    string `json:"state"`
}

var (
	scanResults []*Host
	scanStatus  = "idle"
	scanTime    string
	mu          sync.RWMutex
)

var commonPorts = []int{21, 22, 23, 25, 53, 80, 110, 135, 139, 143, 443, 445, 993, 995, 1433, 1521, 3306, 3389, 5432, 5900, 8080, 8443, 27017}
var portServices = map[int]string{
	21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
	80: "http", 110: "pop3", 135: "msrpc", 139: "netbios", 143: "imap",
	443: "https", 445: "smb", 993: "imaps", 995: "pop3s",
	1433: "mssql", 1521: "oracle", 3306: "mysql", 3389: "rdp",
	5432: "postgres", 5900: "vnc", 8080: "http-proxy", 8443: "https-alt", 27017: "mongodb",
}

func Init() {}

func HandleScan(w http.ResponseWriter, r *http.Request) {
	go runScan()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func HandleResults(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": scanStatus, "results": scanResults, "time": scanTime,
	})
}

func runScan() {
	mu.Lock()
	scanStatus = "scanning"
	scanResults = nil
	mu.Unlock()
	hosts := scanNetwork()
	mu.Lock()
	scanResults = hosts
	scanStatus = "complete"
	scanTime = time.Now().UTC().Format(time.RFC3339)
	mu.Unlock()
}

func scanNetwork() []*Host {
	var hosts []*Host
	localIP := getLocalIP()
	if localIP == "" {
		return hosts
	}
	hosts = append(hosts, scanHost(localIP))
	parts := strings.Split(localIP, ".")
	if len(parts) != 4 {
		return hosts
	}
	type result struct{ host *Host }
	ch := make(chan *result, 254)
	var wg sync.WaitGroup
	for i := 1; i < 255; i++ {
		ip := fmt.Sprintf("%s.%s.%s.%d", parts[0], parts[1], parts[2], i)
		if ip == localIP {
			continue
		}
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			if h := scanHost(addr); h != nil {
				ch <- &result{host: h}
			}
		}(ip)
	}
	go func() { wg.Wait(); close(ch) }()
	for r := range ch {
		hosts = append(hosts, r.host)
	}
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].IP < hosts[j].IP })
	return hosts
}

func scanHost(ip string) *Host {
	var openPorts []Port
	var wg sync.WaitGroup
	portCh := make(chan Port, len(commonPorts))
	for _, port := range commonPorts {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%d", ip, p)
			conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			if err != nil {
				return
			}
			conn.Close()
			svc := portServices[p]
			if svc == "" { svc = "unknown" }
			portCh <- Port{Number: p, Protocol: "tcp", Service: svc, State: "open"}
		}(port)
	}
	go func() { wg.Wait(); close(portCh) }()
	for p := range portCh {
		openPorts = append(openPorts, p)
	}
	if len(openPorts) == 0 {
		return nil
	}
	hostname, _ := net.LookupAddr(ip)
	h := &Host{IP: ip, Ports: openPorts, Status: "up"}
	if len(hostname) > 0 { h.Hostname = hostname[0] }
	return h
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil { return "" }
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
