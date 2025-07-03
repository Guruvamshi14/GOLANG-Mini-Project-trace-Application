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
import "go.opentelemetry.io/otel"

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

var items = make(map[int]Item)
var nextId = 1
var mu sync.Mutex

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "Hello World")
}

func currentTime() string {
	return time.Now().Format("15:04:05")
}

func createitem(w http.ResponseWriter, r *http.Request) {
	// Get the automatically created span from HTTP instrumentation
	tracer := otel.Tracer("Create-Item-Service")
	
	// Only create custom spans for business logic, HTTP span is automatic
	ctx, businessSpan := tracer.Start(r.Context(), "business-logic.create-item")
	defer businessSpan.End()

	_, decodeSpan := tracer.Start(ctx, "decode-request")
	var newItem Item
	err := json.NewDecoder(r.Body).Decode(&newItem)
	decodeSpan.End()
	
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Custom span for critical section
	_, criticalSpan := tracer.Start(ctx, "critical-section")
	mu.Lock()
	newItem.ID = nextId
	items[nextId] = newItem
	nextId++
	mu.Unlock()
	criticalSpan.End()

	// Custom span for response encoding
	_, encodeSpan := tracer.Start(ctx, "encode-response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
	encodeSpan.End()

	fmt.Printf("Created item with ID: %d at %s\n", newItem.ID, currentTime())
}

func getItem(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("Get-Item-Service")
	ctx, businessSpan := tracer.Start(r.Context(), "business-logic.get-item")
	defer businessSpan.End()

	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid Item Id", http.StatusBadRequest)
		return
	}

	_, dbSpan := tracer.Start(ctx, "database.lookup")
	mu.Lock()
	item, exist := items[id]
	mu.Unlock()
	dbSpan.End()

	if !exist {
		http.Error(w, "Items not found", http.StatusNotFound)
		return
	}

	_, encodeSpan := tracer.Start(ctx, "encode-response")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
	encodeSpan.End()
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("Update-Item-Service")
	ctx, businessSpan := tracer.Start(r.Context(), "business-logic.update-item")
	defer businessSpan.End()

	idStr := r.URL.Path[len("/updatedItem/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	_, decodeSpan := tracer.Start(ctx, "decode-request")
	var updatedItem Item
	err = json.NewDecoder(r.Body).Decode(&updatedItem)
	decodeSpan.End()
	
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	_, dbSpan := tracer.Start(ctx, "database.update")
	mu.Lock()
	_, exists := items[id]
	if !exists {
		mu.Unlock()
		http.Error(w, "Id is not present", http.StatusBadRequest)
		return
	}
	items[id] = updatedItem
	mu.Unlock()
	dbSpan.End()

	_, encodeSpan := tracer.Start(ctx, "encode-response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(updatedItem)
	encodeSpan.End()
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("Delete-Item-Service")
	ctx, businessSpan := tracer.Start(r.Context(), "business-logic.delete-item")
	defer businessSpan.End()

	idstr := r.URL.Path[len("/deleteItem/"):]
	id, err := strconv.Atoi(idstr)

	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	_, checkSpan := tracer.Start(ctx, "database.check-exists")
	_, exist := items[id]
	checkSpan.End()

	if !exist {
		http.Error(w, "Data is not present", http.StatusBadRequest)
		return
	}

	_, deleteSpan := tracer.Start(ctx, "database.delete")
	mu.Lock()
	delete(items, id)
	mu.Unlock()
	deleteSpan.End()

	_, encodeSpan := tracer.Start(ctx, "encode-response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode("Given ID is deleted")
	encodeSpan.End()
}

func main() {
	// Initialize tracing
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
	
	// Register handlers - NO manual HTTP span creation needed!
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/createItem", createitem)
	mux.HandleFunc("/items/", getItem)
	mux.HandleFunc("/updatedItem/", updateItem)
	mux.HandleFunc("/deleteItem/", deleteItem)

	// Wrap with automatic HTTP instrumentation
	// This automatically creates spans for ALL HTTP requests
	instrumentedHandler := otelhttp.NewHandler(mux, "item-management-api")

	fmt.Println("🚀 Starting AUTO-INSTRUMENTED server on :8081...")
	fmt.Println("📊 HTTP requests will be automatically traced!")
	fmt.Println("🔍 Business logic uses focused custom spans")
	
	// Start server with automatic HTTP tracing
	err = http.ListenAndServe(":8081", instrumentedHandler)
	if err != nil {
		fmt.Println("Error in starting the server", err)
	}
}