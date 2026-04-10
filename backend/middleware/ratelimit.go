package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type visitorStore struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

func newVisitorStore() *visitorStore {
	store := &visitorStore{
		visitors: make(map[string]*visitor),
	}

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			store.mu.Lock()
			for ip, v := range store.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(store.visitors, ip)
				}
			}
			store.mu.Unlock()
		}
	}()

	return store
}
func (s *visitorStore) getVisitor(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.visitors[ip]
	if !exists {
		// rate: 10 requests per second
		limiter := rate.NewLimiter(rate.Limit(10), 20)
		s.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

var store = newVisitorStore()

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {

			slog.Warn("could not parse remote addr", "addr", r.RemoteAddr)
			next.ServeHTTP(w, r)
			return
		}

		limiter := store.getVisitor(ip)

		if !limiter.Allow() {
			slog.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
			http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}