package camera

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Camera struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	City    string `json:"city"`
	URL     string `json:"url"`
	Type    string `json:"type"` // hls, jpeg, mjpeg, rtsp
}

var cameras = []Camera{
	{"cam1", "Times Square", "US", "New York", "https://videos3.earthcam.com/fecnetwork/9974.flv/playlist.m3u8", "hls"},
	{"cam2", "Eiffel Tower", "FR", "Paris", "https://videos3.earthcam.com/fecnetwork/eiffelcam1.flv/playlist.m3u8", "hls"},
	{"cam3", "Shibuya Crossing", "JP", "Tokyo", "https://videos3.earthcam.com/fecnetwork/shibuyacrossing.flv/playlist.m3u8", "hls"},
	{"cam4", "Big Ben", "GB", "London", "https://videos3.earthcam.com/fecnetwork/bigben.flv/playlist.m3u8", "hls"},
	{"cam5", "Bondi Beach", "AU", "Sydney", "https://videos3.earthcam.com/fecnetwork/bondi2.flv/playlist.m3u8", "hls"},
	{"cam6", "Copacabana", "BR", "Rio", "https://videos3.earthcam.com/fecnetwork/copacabana.flv/playlist.m3u8", "hls"},
	{"cam7", "Dubai Marina", "AE", "Dubai", "https://videos3.earthcam.com/fecnetwork/dubaimarina.flv/playlist.m3u8", "hls"},
	{"cam8", "Victoria Harbour", "CN", "Hong Kong", "https://videos3.earthcam.com/fecnetwork/victoriaharbour.flv/playlist.m3u8", "hls"},
	{"cam9", "Marina Bay", "SG", "Singapore", "https://videos3.earthcam.com/fecnetwork/marinabaysands.flv/playlist.m3u8", "hls"},
	{"cam10", "CN Tower", "CA", "Toronto", "https://videos3.earthcam.com/fecnetwork/cntower.flv/playlist.m3u8", "hls"},
	{"cam11", "Brandenburg Gate", "DE", "Berlin", "https://videos3.earthcam.com/fecnetwork/brandenburggate.flv/playlist.m3u8", "hls"},
	{"cam12", "Table Mountain", "ZA", "Cape Town", "https://videos3.earthcam.com/fecnetwork/tablemountain.flv/playlist.m3u8", "hls"},
}

var mu sync.RWMutex

func Init() {}

func HandleList(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	mu.RLock()
	defer mu.RUnlock()
	var filtered []Camera
	for _, c := range cameras {
		if country == "" || strings.EqualFold(c.Country, country) {
			filtered = append(filtered, c)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

// HandleProxy proxies camera streams to fix CORS/MIME issues
func HandleProxy(w http.ResponseWriter, r *http.Request) {
	camID := r.URL.Query().Get("id")
	if camID == "" {
		http.Error(w, "missing camera id", 400)
		return
	}
	
	var cam *Camera
	mu.RLock()
	for i := range cameras {
		if cameras[i].ID == camID {
			cam = &cameras[i]
			break
		}
	}
	mu.RUnlock()
	
	if cam == nil {
		http.Error(w, "camera not found", 404)
		return
	}

	// Build upstream URL
	upstreamURL := cam.URL
	
	// For HLS streams, return the manifest with proper MIME
	if cam.Type == "hls" || strings.HasSuffix(upstreamURL, ".m3u8") {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		resp, err := http.Get(upstreamURL)
		if err != nil {
			http.Error(w, "upstream error", 502)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
		return
	}

	// For other types, redirect
	http.Redirect(w, r, upstreamURL, http.StatusFound)
}

func HandleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "url": ""})
}

func IsValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}
