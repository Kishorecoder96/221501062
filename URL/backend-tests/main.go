package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "strings"
    "sync"
    "time"

    "URL/logging-middleware"
)

// Request struct for creating short URL
type CreateShortURLRequest struct {
    URL       string `json:"url"`
    Validity  int    `json:"validity,omitempty"`
    Shortcode string `json:"shortcode,omitempty"`
}

// Response struct
type CreateShortURLResponse struct {
    ShortLink string `json:"shortLink"`
    Expiry    string `json:"expiry"`
}

// Click data
type ClickData struct {
    Time     time.Time `json:"time"`
    Referrer string    `json:"referrer"`
    Geo      string    `json:"geo"`
}

// URL data
type URLData struct {
    OriginalURL string
    CreatedAt   time.Time
    Expiry      time.Time
    Clicks      []ClickData
}

var (
    urlStore = make(map[string]*URLData)
    storeMu  sync.RWMutex
)

const (
    host                  = "http://localhost:8080/"
    defaultValidityMinutes = 30
)

// Generate shortcode
func generateShortCode(n int) string {
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

// Handler to create short URL
func createShortURLHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req CreateShortURLRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil || req.URL == "" {
        http.Error(w, `{"error":"Invalid request body or missing URL"}`, http.StatusBadRequest)
        return
    }

    code := req.Shortcode
    if code == "" {
        for {
            code = generateShortCode(6)
            storeMu.RLock()
            _, exists := urlStore[code]
            storeMu.RUnlock()
            if !exists {
                break
            }
        }
    } else {
        storeMu.RLock()
        _, exists := urlStore[code]
        storeMu.RUnlock()
        if exists {
            http.Error(w, `{"error":"Shortcode already exists"}`, http.StatusBadRequest)
            return
        }
    }

    validity := req.Validity
    if validity == 0 {
        validity = defaultValidityMinutes
    }

    createdAt := time.Now()
    expiry := createdAt.Add(time.Duration(validity) * time.Minute)

    data := &URLData{
        OriginalURL: req.URL,
        CreatedAt:   createdAt,
        Expiry:      expiry,
        Clicks:      []ClickData{},
    }

    storeMu.Lock()
    urlStore[code] = data
    storeMu.Unlock()

    resp := CreateShortURLResponse{
        ShortLink: host + code,
        Expiry:    expiry.Format(time.RFC3339),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}

// Redirect handler
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    code := strings.TrimPrefix(r.URL.Path, "/")
    storeMu.RLock()
    data, exists := urlStore[code]
    storeMu.RUnlock()
    if !exists {
        http.Error(w, `{"error":"Shortcode not found"}`, http.StatusNotFound)
        return
    }

    if time.Now().After(data.Expiry) {
        http.Error(w, `{"error":"Short URL has expired"}`, http.StatusGone)
        return
    }

    click := ClickData{
        Time:     time.Now(),
        Referrer: r.Referer(),
        Geo:      "IN",
    }

    storeMu.Lock()
    data.Clicks = append(data.Clicks, click)
    storeMu.Unlock()

    http.Redirect(w, r, data.OriginalURL, http.StatusFound)
}

// Stats handler
func statsHandler(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) != 3 {
        http.Error(w, `{"error":"Invalid path"}`, http.StatusBadRequest)
        return
    }
    code := parts[2]

    storeMu.RLock()
    data, exists := urlStore[code]
    storeMu.RUnlock()
    if !exists {
        http.Error(w, `{"error":"Shortcode not found"}`, http.StatusNotFound)
        return
    }

    resp := map[string]interface{}{
        "totalClicks": len(data.Clicks),
        "originalURL": data.OriginalURL,
        "createdAt":   data.CreatedAt.Format(time.RFC3339),
        "expiry":      data.Expiry.Format(time.RFC3339),
        "clicks":      data.Clicks,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func main() {
    rand.Seed(time.Now().UnixNano())

    mux := http.NewServeMux()
    mux.HandleFunc("/shorturls", createShortURLHandler)
    mux.HandleFunc("/shorturls/", statsHandler)
    mux.HandleFunc("/", redirectHandler)

    loggedMux := logging.Middleware(mux)

    fmt.Println("Server running at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
