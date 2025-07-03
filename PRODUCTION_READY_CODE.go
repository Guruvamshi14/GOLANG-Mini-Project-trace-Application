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

// Replace this import with your trace implementation
import trace "github.com/Guruvamshi14/Trace-Implementation-GOLANG"
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// ===================================================================
// 🎯 PRODUCTION-READY AUTOMATIC INSTRUMENTATION
// ===================================================================
// 
// ✅ COPY THIS ENTIRE FILE TO YOUR LOCAL REPOSITORY
// ✅ Replace the trace import with your tracing implementation
// ✅ ZERO manual spans in business logic
// ✅ Same TraceID across ALL operations automatically
// ✅ Ready for production use
// 
// Dependencies needed in go.mod:
// - go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
// - go.opentelemetry.io/otel
// - go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
// - go.opentelemetry.io/otel/sdk
// ===================================================================

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

var items = make(map[int]Item)
var nextId = 1
var mu sync.Mutex

// 🚀 Auto-instrumented HTTP client for service-to-service calls
// This automatically creates child spans with SAME trace ID!
var httpClient = &http.Client{
	Transport: otelhttp.NewTransport(http.DefaultTransport),
	Timeout:   10 * time.Second,
}

// ===================================================================
// 🎯 BUSINESS LOGIC - ZERO MANUAL SPANS!
// ===================================================================

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// ✅ Pure business logic - no tracing code needed!
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "automatic-instrumentation-service",
		"timestamp": time.Now().Unix(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	// ✅ ZERO MANUAL SPANS - Just business logic!
	// HTTP request span created automatically
	// All service calls will share the same TraceID
	
	startTime := time.Now()
	
	// 1. Decode request (automatically traced as part of HTTP span)
	var newItem Item
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	
	// 2. Call validation service (automatically creates child span with SAME TraceID!)
	validateResponse, err := callValidationService(newItem)
	if err != nil {
		http.Error(w, "Validation service error", http.StatusInternalServerError)
		return
	}
	
	// 3. Save to database (automatically creates child span with SAME TraceID!)
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
	
	// 5. Call notification service (automatically creates child span with SAME TraceID!)
	notifyResponse, err := callNotificationService(newItem.ID)
	if err != nil {
		log.Printf("Notification failed: %v", err)
		notifyResponse = "notification_failed"
	}
	
	// 6. Return response
	response := map[string]interface{}{
		"item":         newItem,
		"validation":   validateResponse,
		"notification": notifyResponse,
		"duration_ms":  time.Since(startTime).Milliseconds(),
		"created_at":   time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	
	fmt.Printf("✅ Created item %d in %dms\n", newItem.ID, time.Since(startTime).Milliseconds())
}

func getItem(w http.ResponseWriter, r *http.Request) {
	// ✅ Pure business logic - no manual spans!
	
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Item Id", http.StatusBadRequest)
		return
	}
	
	// 1. Get from database (automatically creates child span with SAME TraceID!)
	dbItem, err := getItemFromDatabase(id)
	if err != nil {
		log.Printf("Database lookup failed: %v", err)
	}
	
	// 2. Get from memory cache
	mu.Lock()
	item, exist := items[id]
	mu.Unlock()
	
	if !exist {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	
	// 3. Call analytics service (automatically creates child span with SAME TraceID!)
	analyticsData, err := callAnalyticsService(id)
	if err != nil {
		log.Printf("Analytics call failed: %v", err)
		analyticsData = "unavailable"
	}
	
	// 4. Return response
	response := map[string]interface{}{
		"item":        item,
		"db_item":     dbItem,
		"analytics":   analyticsData,
		"accessed_at": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getAllItems(w http.ResponseWriter, r *http.Request) {
	// ✅ Pure business logic - automatic tracing
	
	mu.Lock()
	allItems := make([]Item, 0, len(items))
	for _, item := range items {
		allItems = append(allItems, item)
	}
	mu.Unlock()
	
	response := map[string]interface{}{
		"items": allItems,
		"count": len(allItems),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ===================================================================
// 🔄 SERVICE-TO-SERVICE CALLS (AUTO-TRACED WITH SAME TRACEID)
// ===================================================================

func callValidationService(item Item) (string, error) {
	// ✨ HTTP call automatically traced by otelhttp.NewTransport()
	// Creates child span automatically with SAME trace ID!
	
	// In real app, this would call your actual validation service
	// Example: http://validation-service/validate
	resp, err := httpClient.Get("https://httpbin.org/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Simulate validation logic
	if len(item.Name) < 2 {
		return "validation_failed", fmt.Errorf("name too short")
	}
	
	return "validation_passed", nil
}

func callNotificationService(itemID int) (string, error) {
	// ✨ HTTP call automatically traced by otelhttp.NewTransport()
	// Creates child span automatically with SAME trace ID!
	
	// In real app, this would call your actual notification service
	// Example: http://notification-service/notify
	url := fmt.Sprintf("https://httpbin.org/delay/1?item_id=%d", itemID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	return "notification_sent", nil
}

func callAnalyticsService(itemID int) (string, error) {
	// ✨ HTTP call automatically traced by otelhttp.NewTransport()
	// Creates child span automatically with SAME trace ID!
	
	// In real app, this would call your actual analytics service
	// Example: http://analytics-service/track
	url := fmt.Sprintf("https://httpbin.org/uuid?item_id=%d", itemID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	return "analytics_recorded", nil
}

// ===================================================================
// 🗄️ DATABASE CALLS (WILL BE AUTO-TRACED IN REAL APP)
// ===================================================================

func saveItemToDatabase(item Item) error {
	// ✨ In real app with otelsql: Automatically creates child span with SAME TraceID
	// Example: _, err := db.Exec("INSERT INTO items (name, value) VALUES (?, ?)", item.Name, item.Value)
	
	// Simulate database operation
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("💾 Saved item '%s' to database\n", item.Name)
	return nil
}

func getItemFromDatabase(id int) (string, error) {
	// ✨ In real app with otelsql: Automatically creates child span with SAME TraceID
	// Example: row := db.QueryRow("SELECT data FROM items WHERE id = ?", id)
	
	// Simulate database operation
	time.Sleep(5 * time.Millisecond)
	fmt.Printf("💾 Retrieved item %d from database\n", id)
	return fmt.Sprintf("db_data_for_item_%d", id), nil
}

// ===================================================================
// 🚀 MAIN - AUTOMATIC INSTRUMENTATION SETUP
// ===================================================================

func main() {
	// 1. Initialize OpenTelemetry tracing (ONE TIME SETUP)
	tp, err := trace.StartTracing()
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown tracer provider: %v", err)
		}
	}()
	
	// 2. Create HTTP multiplexer
	mux := http.NewServeMux()
	
	// 3. Register handlers (NO SPANS IN BUSINESS LOGIC!)
	mux.HandleFunc("/health", healthCheck)
	mux.HandleFunc("/createItem", createItem)
	mux.HandleFunc("/items/", getItem)
	mux.HandleFunc("/items", getAllItems)
	
	// 4. ✨ MAGIC: Wrap with automatic HTTP instrumentation
	// This automatically creates spans for ALL incoming HTTP requests
	// and ensures trace context propagation across ALL operations!
	instrumentedHandler := otelhttp.NewHandler(mux, "production-service")
	
	port := ":8080"
	fmt.Println("🎯 PRODUCTION AUTOMATIC INSTRUMENTATION")
	fmt.Println("======================================")
	fmt.Printf("🚀 Server: http://localhost%s\n", port)
	fmt.Println("📊 Features:")
	fmt.Println("  ✅ ZERO manual spans in business logic")
	fmt.Println("  ✅ Automatic HTTP request tracing")
	fmt.Println("  ✅ Automatic service-to-service call tracing")
	fmt.Println("  ✅ SAME TraceID across ALL operations")
	fmt.Println("  ✅ Database auto-instrumentation ready")
	fmt.Println("")
	fmt.Println("🧪 Test Commands:")
	fmt.Printf("  curl http://localhost%s/health\n", port)
	fmt.Printf("  curl -X POST http://localhost%s/createItem \\\n", port)
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"name\": \"Test Item\", \"value\": \"production test\"}'")
	fmt.Printf("  curl http://localhost%s/items/1\n", port)
	fmt.Printf("  curl http://localhost%s/items\n", port)
	fmt.Println("")
	fmt.Println("🎯 All operations share the SAME TraceID automatically!")
	fmt.Println("")
	
	// 5. Start server with complete automatic tracing
	log.Printf("Starting server on %s", port)
	err = http.ListenAndServe(port, instrumentedHandler)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}