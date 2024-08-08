package main

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

var (
	mu       sync.Mutex
	userIp   = make(map[string]*rate.Limiter)
	reqLimit = rate.NewLimiter(1, 5)
)

func getUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func getUserLimit(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	_, exists := userIp[ip]
	if !exists {
		userIp[ip] = reqLimit
	}

	return userIp[ip]

}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		srcAddr := getUserIP(r)

		usrReq := getUserLimit(srcAddr)

		if !usrReq.Allow() {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		fmt.Println(srcAddr)
		fmt.Fprint(w, "Hello, Humanz!")
	})

	http.HandleFunc("/kano", func(w http.ResponseWriter, r *http.Request) {
		srcAddr := getUserIP(r)

		fmt.Fprint(w, "Hello ", srcAddr)
	})

	if err := http.ListenAndServe(":2525", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
