package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

type IncommingLimit struct {
	limit Limit
	brust int
}

type client struct {
	limiter  *Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*client)
)

func init() {
	go CleanupClients()
}

// CleanupClients Every minute check the map for clients that haven't been seen for
// more than 3 minutes and delete the entries.
func CleanupClients() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for id, client := range clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(clients, id)
			}
		}
		mu.Unlock()
	}
}

// PerClientRateLimiter Middleware that checks limiter specific to the ID to rate limit the request.
func PerClientRateLimiter(next func(writer http.ResponseWriter, request *http.Request), incommming IncommingLimit) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path
		mu.Lock()
		if _, found := clients[id]; !found {
			clients[id] = &client{limiter: NewLimiter(incommming.limit, incommming.brust)}
		}
		clients[id].lastSeen = time.Now()
		if !clients[id].limiter.Allow() {
			mu.Unlock()
			log.Println(time.Now(), " Request denied for ", id)
			RecordDenied(id)
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		mu.Unlock()
		next(w, r)
	})
}

// ResponseData Builds and returns a consistent response for client, sets the content type, status code and JSON body
func ResponseData(writer http.ResponseWriter, message Message) {

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

// UserHandler handler that servers user detail to user endpoint
func UserHandler(writer http.ResponseWriter, request *http.Request) {
	pattern := "/user/:id/data"
	path := request.URL.Path
	id, _ := ParseID(path, pattern)

	RecordSuccess(path)

	message := Message{
		Status: "Successful",
		Body:   "Hey I am user " + id,
	}
	ResponseData(writer, message)
}

// AdminHandler handler that servers admin detail to admin endpoint
func AdminHandler(writer http.ResponseWriter, request *http.Request) {
	pattern := "/admin/:id/data"
	path := request.URL.Path
	id, _ := ParseID(path, pattern)

	RecordSuccess(path)

	message := Message{
		Status: "Successful",
		Body:   "Hey I am Admin " + id,
	}
	ResponseData(writer, message)
}

// PublicInfoHandler handler that servers public info to public info handler
func PublicInfoHandler(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	RecordSuccess(path)

	message := Message{
		Status: "Successful",
		Body:   "Hey I a Public Info Handler",
	}
	ResponseData(writer, message)
}

// UpdateRateLimiterHandler updates the rate limiter for all the clients
func UpdateRateLimiterHandler(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query().Get("limit")
	limit, err := strconv.ParseFloat(query, 64)
	if err != nil {
		log.Println("Error converting string to float64:", err)
		return
	}
	for _, client := range clients {
		clientLimiter := *client
		clientLimiter.limiter.SetLimit(Limit(limit))
		clientLimiter.limiter.SetBurst(int(limit))
	}
	message := Message{
		Status: "Successful",
		Body:   "Limiter updated to " + query,
	}
	ResponseData(writer, message)
}

// ViewMetricsHandler handler to server the metrics that records successful and denial counts of endpoints
func ViewMetricsHandler(writer http.ResponseWriter, request *http.Request) {

	var metricsSlice []Metric
	for _, metric := range metrics {
		metricsSlice = append(metricsSlice, metric)
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(metricsSlice); err != nil {
		http.Error(writer, "Failed to encode metrics", http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/user/", PerClientRateLimiter(UserHandler, IncommingLimit{5, 5}))
	mux.Handle("/admin/", PerClientRateLimiter(AdminHandler, IncommingLimit{2, 3}))
	mux.HandleFunc("/public/info", PublicInfoHandler)
	mux.HandleFunc("/update/rates", UpdateRateLimiterHandler)
	mux.HandleFunc("/metrics", ViewMetricsHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
}
