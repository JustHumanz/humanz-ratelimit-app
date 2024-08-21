package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type userAttr struct {
	usrLimit *rate.Limiter
	timeout  time.Time
}

var (
	mu       sync.Mutex
	userIp   = make(map[string]userAttr)
	appLimit = rate.NewLimiter(1, 5)
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
		userIp[ip] = userAttr{
			usrLimit: appLimit,
			timeout:  time.Now(),
		}
	}

	return userIp[ip].usrLimit

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

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			mu.Lock()

			for ip, usrTimeout := range userIp {
				if time.Since(usrTimeout.timeout) > 5*time.Minute {
					fmt.Println(fmt.Sprintf("%s not interact with app for %d, cleanup the struct", ip, 5*time.Minute))
					delete(userIp, ip)
				}
			}
			mu.Unlock()
		}
	}()

	if err := http.ListenAndServe(":2525", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
