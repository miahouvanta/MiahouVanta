package attackmap

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Attack struct {
	ID            string  `json:"id"`
	SrcLat        float64 `json:"src_lat"`
	SrcLng        float64 `json:"src_lng"`
	TgtLat        float64 `json:"tgt_lat"`
	TgtLng        float64 `json:"tgt_lng"`
	SrcCity       string  `json:"src_city"`
	SrcCountry    string  `json:"src_country"`
	TgtCity       string  `json:"tgt_city"`
	TgtCountry    string  `json:"tgt_country"`
	City          string  `json:"city"`
	Country       string  `json:"country"`
	Type          string  `json:"type"`
	Severity      string  `json:"severity"`
	SrcIP         string  `json:"src_ip"`
	TgtIP         string  `json:"tgt_ip"`
	Port          int     `json:"port"`
	Protocol      string  `json:"protocol"`
	TargetService string  `json:"target_service"`
	Timestamp     string  `json:"timestamp"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Attack)
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mu        sync.Mutex
)

var attackTypes = []struct {
	Type     string
	Severity string
	Ports    []int
	Protocol string
	Service  string
}{
	{"DDoS", "critical", []int{80, 443, 8080}, "UDP", "Web Server"},
	{"BruteForce", "high", []int{22, 3389, 21, 23}, "TCP", "SSH/RDP"},
	{"Exploit", "critical", []int{445, 135, 139, 3389}, "TCP", "SMB/RDP"},
	{"Malware", "high", []int{4444, 5555, 8080, 9999}, "TCP", "C2 Channel"},
	{"Ransomware", "critical", []int{445, 3389, 5985}, "TCP", "SMB/WinRM"},
	{"Phishing", "medium", []int{80, 443, 8443}, "TCP", "Web Server"},
	{"SQLi", "high", []int{3306, 1433, 5432}, "TCP", "Database"},
	{"XSS", "medium", []int{80, 443, 8080}, "TCP", "Web Server"},
	{"Recon", "low", []int{21, 22, 23, 25, 53, 80, 443}, "TCP", "Multiple"},
	{"Privilege", "high", []int{0}, "TCP", "System"},
}

var realCits = []struct {
	City, Country, Code string
	Lat, Lng            float64
}{
	{"New York", "United States", "US", 40.7128, -74.0060},
	{"London", "United Kingdom", "GB", 51.5074, -0.1278},
	{"Tokyo", "Japan", "JP", 35.6762, 139.6503},
	{"Frankfurt", "Germany", "DE", 50.1109, 8.6821},
	{"Singapore", "Singapore", "SG", 1.3521, 103.8198},
	{"Sydney", "Australia", "AU", -33.8688, 151.2093},
	{"Ashburn", "United States", "US", 39.0438, -77.4874},
	{"Mumbai", "India", "IN", 19.0760, 72.8777},
	{"Sao Paulo", "Brazil", "BR", -23.5505, -46.6333},
	{"Moscow", "Russia", "RU", 55.7558, 37.6173},
	{"Paris", "France", "FR", 48.8566, 2.3522},
	{"Amsterdam", "Netherlands", "NL", 52.3676, 4.9041},
	{"Seoul", "South Korea", "KR", 37.5665, 126.9780},
	{"Toronto", "Canada", "CA", 43.6532, -79.3832},
	{"Lagos", "Nigeria", "NG", 6.5244, 3.3792},
	{"Jakarta", "Indonesia", "ID", -6.2088, 106.8456},
	{"Dubai", "UAE", "AE", 25.2048, 55.2708},
	{"Beijing", "China", "CN", 39.9042, 116.4074},
}

// Realistic source IP pools by region (common scanner/botnet IPs)
var sourceIPPools = map[string][]string{
	"CN": {"61.135.", "116.228.", "220.181.", "183.222."},
	"RU": {"5.18.", "31.13.", "95.165.", "185.220."},
	"US": {"72.21.", "104.16.", "151.101.", "185.199."},
	"DE": {"89.163.", "178.63.", "94.130.", "136.243."},
	"BR": {"177.54.", "179.188.", "187.45.", "191.252."},
	"IN": {"103.21.", "117.197.", "49.36.", "45.248."},
	"JP": {"126.32.", "133.11.", "150.95.", "202.218."},
	"KP": {"175.45.", "210.56."},
}

func Init() {
	go generateAttacks()
	go hub()
}

func hub() {
	for {
		select {
		case attack := <-broadcast:
			data, _ := json.Marshal(attack)
			mu.Lock()
			for c := range clients {
				if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
					c.Close()
					delete(clients, c)
				}
			}
			mu.Unlock()
		}
	}
}

func generateAttacks() {
	for i := 0; ; i++ {
		atk := attackTypes[rand.Intn(len(attackTypes))]
		srcCity := realCits[rand.Intn(len(realCits))]
		tgtCity := realCits[rand.Intn(len(realCits))]
		for tgtCity.City == srcCity.City {
			tgtCity = realCits[rand.Intn(len(realCits))]
		}

		srcIP := getRealisticIP(srcCity.Code)
		tgtIP := getRealisticIP(tgtCity.Code)
		port := atk.Ports[rand.Intn(len(atk.Ports))]
		if port == 0 {
			port = 1024 + rand.Intn(64511)
		}

		a := Attack{
			ID:            fmt.Sprintf("ATK-%s-%05d", strings.ToUpper(atk.Type[:3]), i),
			SrcLat:        srcCity.Lat + (rand.Float64()-0.5)*3,
			SrcLng:        srcCity.Lng + (rand.Float64()-0.5)*3,
			TgtLat:        tgtCity.Lat + (rand.Float64()-0.5)*2,
			TgtLng:        tgtCity.Lng + (rand.Float64()-0.5)*2,
			SrcCity:       srcCity.City,
			SrcCountry:    srcCity.Country,
			TgtCity:       tgtCity.City,
			TgtCountry:    tgtCity.Country,
			City:          tgtCity.City,
			Country:       tgtCity.Country,
			Type:          atk.Type,
			Severity:      atk.Severity,
			SrcIP:         srcIP,
			TgtIP:         tgtIP,
			Port:          port,
			Protocol:      atk.Protocol,
			TargetService: atk.Service,
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
		}
		broadcast <- a
		time.Sleep(time.Duration(80+rand.Intn(300)) * time.Millisecond)
	}
}

func getRealisticIP(countryCode string) string {
	pool, ok := sourceIPPools[countryCode]
	if !ok || len(pool) == 0 {
		// Generate generic IP
		return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(223)+1, rand.Intn(256), rand.Intn(256), rand.Intn(256))
	}
	prefix := pool[rand.Intn(len(pool))]
	suffix := rand.Intn(256)
	return prefix + fmt.Sprintf("%d.%d", rand.Intn(256), suffix)
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS error:", err)
		return
	}
	mu.Lock()
	clients[conn] = true
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close()
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
