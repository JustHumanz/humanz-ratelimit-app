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

func getUserIP(raddr string) (string, error) {
	ip, _, err := net.SplitHostPort(raddr)
	if err != nil {
		return "", err
	}

	return ip, nil

}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		srcAddr, err := getUserIP(r.RemoteAddr)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		}

		usrReq := getUserLimit(srcAddr)

		if !usrReq.Allow() {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		fmt.Println(srcAddr)
		fmt.Fprint(w, "Hello, Humanz!")
	})

	http.HandleFunc("/kano", func(w http.ResponseWriter, r *http.Request) {
		srcAddr, err := getUserIP(r.RemoteAddr)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		}

		fmt.Fprint(w, "Hello", srcAddr)
	})

	if err := http.ListenAndServe(":2525", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
