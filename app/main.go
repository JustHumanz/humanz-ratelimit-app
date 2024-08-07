package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

var (
	mu       sync.Mutex
	userIp   = make(map[string]*rate.Limiter)
	reqLimit = rate.NewLimiter(1, 5)
)

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

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		usrReq := getUserLimit(ip)

		if !usrReq.Allow() {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		fmt.Println(ip)
		fmt.Fprint(w, "Hello, Humanz!")
	})

	if err := http.ListenAndServe(":2525", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
