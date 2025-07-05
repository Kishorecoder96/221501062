package logging

import (
    "log"
    "net/http"
    "time"
)

func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request - Method: %s, Path: %s, Time: %s\n",
            r.Method, r.URL.Path, time.Now().Format(time.RFC3339))
        next.ServeHTTP(w, r)
    })
}
