# URL Shortener Microservice

## ✨ Features

- Create short URLs with optional custom code and expiry time
- Redirect to original URL
- Track and display click stats (number of clicks, time, referrer, geo)
- Simple in-memory storage
- Logging middleware included

---

## ⚙️ How to run

```bash
go mod tidy
go run backend-tests/main.go

```

curl -X POST -H "Content-Type: application/json" -d "{\"url\":\"https://golang.org\",\"validity\":10,\"shortcode\":\"goLang\"}" http://localhost:8080/shorturls

"http://localhost:8080/goLang"

