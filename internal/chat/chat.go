package chat

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

var (
	history []Message
	mu      sync.RWMutex
)

func Init() {
	history = make([]Message, 0, 100)
}

func HandleSend(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	userMsg := Message{Role: "user", Content: msg.Message, Timestamp: time.Now().UTC().Format(time.RFC3339)}
	mu.Lock()
	history = append(history, userMsg)
	if len(history) > 100 {
		history = history[len(history)-100:]
	}
	mu.Unlock()

	response := miahouRespond(msg.Message)

	aiMsg := Message{Role: "miahou", Content: response, Timestamp: time.Now().UTC().Format(time.RFC3339)}
	mu.Lock()
	history = append(history, aiMsg)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aiMsg)
}

func miahouRespond(input string) string {
	lower := strings.ToLower(input)

	if containsAny(lower, []string{"hello", "hi", "hey", "yo", "salut", "bonjour"}) {
		return pick([]string{
			"Hey! (>w<) Miahou here, ready to hack the planet! What do you need?",
			"Hello! I'm online and ready. What are we working on today?",
			"Yo! CyberDeck is live. What's the plan?",
		})
	}
	if containsAny(lower, []string{"how are you", "ca va", "comment ca va", "quoi de neuf"}) {
		return pick([]string{
			"Running smooth! All systems green. What about you?",
			"Feeling great! Just analyzed some attack data. Ready for more!",
			"All good on my end. The network is quiet... for now (>w<)",
		})
	}
	if containsAny(lower, []string{"attack", "ddos", "malware", "ransomware", "exploit"}) {
		return pick([]string{
			"Attacks are flowing in live on the map. The usual suspects - botnets from CN/RU targeting US/EU infrastructure. DDoS is trending highest right now.",
			"I'm seeing a lot of brute force attempts on SSH and RDP services. Make sure your ports are locked down!",
			"The threat landscape is active. Most attacks are automated scanners looking for low-hanging fruit.",
		})
	}
	if containsAny(lower, []string{"scan", "nmap", "network", "vulnerability", "vuln"}) {
		return pick([]string{
			"Run a network scan from the Vuln Dashboard - I'll discover all hosts and open ports on your local network. Takes about 30 seconds.",
			"I can scan your local network right now. Just hit the scan button on the Vuln Dashboard page.",
		})
	}
	if containsAny(lower, []string{"log", "analyze", "syslog"}) {
		return pick([]string{
			"Upload any log file on the Log Analyzer page - I'll detect brute force, port scans, DDoS patterns, SQLi, XSS, and more.",
			"Paste your logs into the Log Analyzer and I'll classify all attack patterns with severity ratings.",
		})
	}
	if containsAny(lower, []string{"camera", "cam", "watch"}) {
		return pick([]string{
			"Check the Cameras section! I've got live public cameras from 12 cities worldwide. Click any camera for a full-screen modal view.",
		})
	}
	if containsAny(lower, []string{"c2", "command", "agent", "implant"}) {
		return pick([]string{
			"The C2 server listener is configurable from the C2 page. Set your bind IP and port, then start listening. Agents connect via TCP.",
			"You can start/stop the C2 listener and change the port from the C2 dashboard. Once agents connect, you can send commands directly.",
		})
	}
	if containsAny(lower, []string{"help", "what can you do", "commands", "menu"}) {
		return "Here's what I can do:\n\n[Attack Map] Real-time global attack visualization\n[C2 Server] Command & control for agents\n[Log Analyzer] Detect attacks in log files\n[Vuln Dashboard] Network scanning\n[Cameras] Live public cameras worldwide\n[Chat] You're talking to me right now! (>w<)"
	}
	if containsAny(lower, []string{"thank", "thanks", "merci"}) {
		return pick([]string{
			"Anytime! (>w<)",
			"That's what I'm here for!",
			"No problem! Let me know if you need anything else.",
		})
	}
	return pick([]string{
		"Interesting. Tell me more about what you're working on.",
		"I'm not sure I understand, but I'm listening! Could you rephrase that? (>w<)",
		"Good point. Want me to run a scan or check the attack map?",
		"Hmm, let me think about that... In the meantime, check out the live attack map!",
	})
}

func pick(opts []string) string {
	return opts[time.Now().UnixNano()%int64(len(opts))]
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
