package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"sync"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
)

type handlers struct {
	sf singleflight.Group
}

var (
	cache = make(map[string][]byte)
	mu    sync.RWMutex
)

func init() {
	// Seed data with bcrypt hashed passwords
	addUser("user1", "password1")
	addUser("user2", "password2")
	addUser("user3", "password3")
}

// AB command to test the login endpoints:
// Without singleflight:
// ab -n 1000 -c 100 -A user1:password1 http://localhost:8080/login
// With singleflight:
// ab -n 1000 -c 100 -A user1:password1 http://localhost:8080/login-singleflight

func main() {
	fmt.Println("Starting server on port 8080")

	h := &handlers{}

	http.HandleFunc("/login", h.loginHandler)
	http.HandleFunc("/login-singleflight", h.loginHandlerWithSingleflight)

	fmt.Println("Endpoints:")
	fmt.Println("http://localhost:8080/login")
	fmt.Println("http://localhost:8080/login-singleflight")
	fmt.Println("\nUse a tool like Apache Benchmark (ab) to test the performance difference")
	fmt.Println("Example: ab -n 1000 -c 100 http://localhost:8080/login")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (h *handlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if validateCredentials(username, password) {
		w.Write([]byte("Login successful"))
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func (h *handlers) loginHandlerWithSingleflight(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, _, _ := h.sf.Do(username, func() (interface{}, error) {
		return validateCredentials(username, password), nil
	})

	if result.(bool) {
		w.Write([]byte("Login successful"))
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func addUser(username, password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error hashing password for %s: %v", username, err)
	}
	mu.Lock()
	defer mu.Unlock()
	cache[username] = hashedPassword
}

func validateCredentials(username, password string) bool {
	mu.RLock()
	defer mu.RUnlock()

	hashedPassword, exists := cache[username]
	time.Sleep(100 * time.Millisecond) // Simulate slow lookup

	if !exists {
		return false
	}
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil
}
