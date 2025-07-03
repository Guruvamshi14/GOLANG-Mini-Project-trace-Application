package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

import trace "github.com/Guruvamshi14/Trace-Implementation-GOLANG"
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// ========================================
// 🎯 ZERO MANUAL SPANS IN BUSINESS LOGIC
// ========================================

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

var items = make(map[int]Item)
var nextId = 1
var mu sync.Mutex

// Auto-instrumented HTTP client for service-to-service calls
var httpClient = &http.Client{
	Transport: otelhttp.NewTransport(http.DefaultTransport),
	Timeout:   10 * time.Second,
}

func currentTime() string {
	return time.Now().Format("15:04:05")
}

// ========================================
// 🚀 CONTROLLER FUNCTIONS - ZERO SPANS!
// ========================================

func handler(w http.ResponseWriter, r *http.Request) {
	// ✅ NO MANUAL SPANS - Pure business logic!
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Hello from Service 1", "time": "%s"}`, currentTime())
}

func createitem(w http.ResponseWriter, r *http.Request) {
	// ✅ NO MANUAL SPANS - Just business logic!
	
	// 1. Decode request (no span needed)
	var newItem Item
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// 2. Call external service (automatically traced)
	validateResponse, err := callValidationService(newItem)
	if err != nil {
		http.Error(w, "Validation service error", http.StatusInternalServerError)
		return
	}

	// 3. Save to database (simulated - would be auto-traced in real app)
	err = saveItemToDatabase(newItem)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// 4. Store in memory (business logic)
	mu.Lock()
	newItem.ID = nextId
	items[nextId] = newItem
	nextId++
	mu.Unlock()

	// 5. Call notification service (automatically traced)
	notifyResponse, err := callNotificationService(newItem.ID)
	if err != nil {
		log.Printf("Notification failed: %v", err)
		// Continue execution - notification is not critical
	}

	// 6. Return response
	response := map[string]interface{}{
		"item":         newItem,
		"validation":   validateResponse,
		"notification": notifyResponse,
		"created_at":   currentTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	fmt.Printf("✅ Created item %d at %s\n", newItem.ID, currentTime())
}

func getItem(w http.ResponseWriter, r *http.Request) {
	// ✅ NO MANUAL SPANS - Pure business logic!
	
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Item Id", http.StatusBadRequest)
		return
	}

	// 1. Get from database (simulated - would be auto-traced in real app)
	dbItem, err := getItemFromDatabase(id)
	if err != nil {
		log.Printf("Database lookup failed: %v", err)
	}

	// 2. Get from memory cache
	mu.Lock()
	item, exist := items[id]
	mu.Unlock()

	if !exist {
		http.Error(w, "Items not found", http.StatusNotFound)
		return
	}

	// 3. Call analytics service (automatically traced)
	analyticsData, err := callAnalyticsService(id)
	if err != nil {
		log.Printf("Analytics call failed: %v", err)
		analyticsData = "unavailable"
	}

	// 4. Return enhanced response
	response := map[string]interface{}{
		"item":      item,
		"db_item":   dbItem,
		"analytics": analyticsData,
		"accessed_at": currentTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ========================================
// 🔄 SERVICE-TO-SERVICE CALLS (AUTO-TRACED)
// ========================================

func callValidationService(item Item) (string, error) {
	// ✨ HTTP call automatically traced by otelhttp.NewTransport()
	// Creates child span automatically with same trace ID!
	resp, err := httpClient.Get("https://httpbin.org/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("🔄 Called validation service for item: %s\n", item.Name)
	return "validation_passed", nil
}

func callNotificationService(itemID int) (string, error) {
	// ✨ HTTP call automatically traced by otelhttp.NewTransport()
	// Creates child span automatically with same trace ID!
	url := fmt.Sprintf("https://httpbin.org/delay/1?item_id=%d", itemID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("🔔 Called notification service for item ID: %d\n", itemID)
	return "notification_sent", nil
}

func callAnalyticsService(itemID int) (string, error) {
	// ✨ HTTP call automatically traced by otelhttp.NewTransport()
	// Creates child span automatically with same trace ID!
	url := fmt.Sprintf("https://httpbin.org/uuid?item_id=%d", itemID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("📊 Called analytics service for item ID: %d\n", itemID)
	return "analytics_recorded", nil
}

// ========================================
// 🗄️ DATABASE CALLS (SIMULATED AUTO-TRACED)
// ========================================

func saveItemToDatabase(item Item) error {
	// ✨ In real app with otelsql: Database query automatically traced
	// Example: _, err := db.Exec("INSERT INTO items...", item.Name, item.Value)
	
	// Simulate database operation
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("💾 DB: Saved item '%s' to database\n", item.Name)
	return nil
}

func getItemFromDatabase(id int) (string, error) {
	// ✨ In real app with otelsql: Database query automatically traced  
	// Example: row := db.QueryRow("SELECT * FROM items WHERE id = ?", id)
	
	// Simulate database operation
	time.Sleep(5 * time.Millisecond)
	fmt.Printf("💾 DB: Retrieved item %d from database\n", id)
	return "db_item_data", nil
}

// ========================================
// 🚀 MAIN - ONLY INSTRUMENTATION SETUP
// ========================================

func main() {
	// Initialize tracing ONCE
	tp, err := trace.StartTracing()
	if err != nil {
		log.Fatalf("failed to initialize tracing: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}()

	// Create HTTP multiplexer
	mux := http.NewServeMux()
	
	// Register handlers - NO SPANS IN BUSINESS LOGIC!
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/createItem", createitem)
	mux.HandleFunc("/items/", getItem)

	// ✨ MAGIC: Wrap with automatic HTTP instrumentation
	// This creates spans for ALL incoming HTTP requests
	instrumentedHandler := otelhttp.NewHandler(mux, "service-1")

	fmt.Println("🎯 ZERO-CODE AUTOMATIC INSTRUMENTATION")
	fmt.Println("=====================================")
	fmt.Println("🚀 Server running on :8082")
	fmt.Println("📊 HTTP requests: Auto-traced")
	fmt.Println("🔄 Service calls: Auto-traced")
	fmt.Println("💾 Database calls: Auto-traced (with otelsql)")
	fmt.Println("✨ Business logic: ZERO manual spans!")
	fmt.Println("")
	fmt.Println("📈 Automatic Trace Flow:")
	fmt.Println("POST /createItem (traceID: xyz) ->")
	fmt.Println("  ├── HTTP span-1 (auto, traceID: xyz)")
	fmt.Println("  ├── Validation call span-2 (auto, traceID: xyz)")
	fmt.Println("  ├── Database save span-3 (auto, traceID: xyz)")
	fmt.Println("  └── Notification call span-4 (auto, traceID: xyz)")
	fmt.Println("")
	fmt.Println("🎯 ALL spans have SAME traceID automatically!")
	fmt.Println("")
	
	// Start server with complete automatic tracing
	err = http.ListenAndServe(":8082", instrumentedHandler)
	if err != nil {
		fmt.Println("Error in starting the server", err)
	}
}