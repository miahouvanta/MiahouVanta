package main

import (
	"cyberdeck/internal/attackmap"
	"cyberdeck/internal/c2"
	"cyberdeck/internal/camera"
	"cyberdeck/internal/chat"
	"cyberdeck/internal/logs"
	"cyberdeck/internal/scanner"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("CYBERDECK UI - Miahou v1.0")

	attackmap.Init()
	scanner.Init()
	logs.Init()
	camera.Init()
	chat.Init()
	c2.Init()

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/api/attack-map/stream", attackmap.HandleWS)
	http.HandleFunc("/api/scanner/scan", scanner.HandleScan)
	http.HandleFunc("/api/scanner/results", scanner.HandleResults)
	http.HandleFunc("/api/logs/analyze", logs.HandleAnalyze)
	http.HandleFunc("/api/logs/upload", logs.HandleUpload)
	http.HandleFunc("/api/cameras/list", camera.HandleList)
	http.HandleFunc("/api/cameras/proxy", camera.HandleProxy)
	http.HandleFunc("/api/cameras/stream", camera.HandleStream)
	http.HandleFunc("/api/c2/clients", c2.HandleClients)
	http.HandleFunc("/api/c2/send", c2.HandleSend)
	http.HandleFunc("/api/c2/ws", c2.HandleWS)
	http.HandleFunc("/api/c2/config", c2.HandleConfig)
	http.HandleFunc("/api/c2/start", c2.HandleStart)
	http.HandleFunc("/api/c2/stop", c2.HandleStop)
	http.HandleFunc("/api/chat/send", chat.HandleSend)

	tmplDir := "web/templates"
	pages := map[string]struct {
		Title string
		Page  string
	}{
		"/":               {"Attack Map", "attack-map"},
		"/attack-map":     {"Attack Map", "attack-map"},
		"/c2":             {"C2 Server", "c2"},
		"/vuln-dashboard": {"Vuln Dashboard", "vuln-dashboard"},
		"/log-analyzer":   {"Log Analyzer", "log-analyzer"},
		"/cameras":        {"Cameras", "cameras"},
		"/chat":           {"Chat", "chat"},
	}

	for path, info := range pages {
		p := info
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path && path != "/" {
				http.NotFound(w, r)
				return
			}
			t, err := template.ParseFiles(
				tmplDir+"/layout.html",
				tmplDir+"/"+p.Page+".html",
			)
			if err != nil {
				log.Println("Template error:", err)
				http.Error(w, "Template error", 500)
				return
			}
			data := map[string]interface{}{
				"Title":  p.Title,
				"Active": p.Page,
			}
			if err := t.ExecuteTemplate(w, "layout", data); err != nil {
				log.Println("Render error:", err)
			}
		})
	}

	port := os.Getenv("CYBERDECK_PORT")
	if port == "" { port = "8080" }
	log.Printf("CyberDeck running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
