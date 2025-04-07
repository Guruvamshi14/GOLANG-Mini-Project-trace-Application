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

import "go.opentelemetry.io/otel"

// func startTracing() (*trace.TracerProvider, error) {
// 	headers := map[string]string{
// 		"content-type": "application/json",
// 	}

// 	exporter, err := otlptrace.New(
// 		context.Background(),
// 		otlptracehttp.NewClient(
// 			otlptracehttp.WithEndpoint("localhost:4318"),
// 			otlptracehttp.WithHeaders(headers),
// 			otlptracehttp.WithInsecure(),
// 		),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating new exporter: %w", err)
// 	}

// 	tracerprovider := trace.NewTracerProvider(
// 		trace.WithBatcher(
// 			exporter,
// 			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
// 			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
// 			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
// 		),
// 		trace.WithResource(
// 			resource.NewWithAttributes(
// 				semconv.SchemaURL,
// 				semconv.ServiceNameKey.String("product-app"),
// 			),
// 		),
// 	)

// 	otel.SetTracerProvider(tracerprovider)

// 	return tracerprovider, nil
// }

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

	tracer := otel.Tracer("Crete-Item-Tracer")

	ctx, parentSpan := tracer.Start(r.Context(), "Create-Item-Request")
	defer parentSpan.End()

	_, spanEncoding := tracer.Start(ctx, "Encoding-Item")

	spanId := spanEncoding.SpanContext().SpanID().String()
	traceID := spanEncoding.SpanContext().TraceID().String()
	fmt.Println("Trace ID : ", traceID, "Span ID : ", spanId)
	var newItem Item
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		// http.Error(w, "error message", statusCode)
		// 404
		// http.StatusNotFound
		// http.StatusBadRequest

		defer spanEncoding.End()
		return
	}

	defer spanEncoding.End()

	// fmt.Println("Current Time Before entering MUTEX :", currentTime())

	_, spanLock := tracer.Start(ctx, "Entering in to Critical Section")
	mu.Lock()
	spanId = spanLock.SpanContext().SpanID().String()
	traceID = spanLock.SpanContext().TraceID().String()
	// fmt.Println("Trace ID : ", traceID, "Span ID : ", spanId)
	defer spanLock.End()

	// time.Sleep(20 * time.Second)

	_, spanCreateItem := tracer.Start(ctx, "Creating the Item")
	newItem.ID = nextId
	items[nextId] = newItem
	nextId++
	spanId = spanCreateItem.SpanContext().SpanID().String()
	traceID = spanCreateItem.SpanContext().TraceID().String()
	// fmt.Println("Trace ID : ", traceID, "Span ID : ", spanId)
	defer spanCreateItem.End()

	_, spanUnlock := tracer.Start(ctx, "Exit to Critical Section")
	mu.Unlock()
	spanId = spanUnlock.SpanContext().SpanID().String()
	traceID = spanUnlock.SpanContext().TraceID().String()
	fmt.Println("Trace ID : ", traceID, "Span ID : ", spanId)
	defer spanUnlock.End()
	// fmt.Println("Current Time After Exiting to from Critical Section :", currentTime())

	// time.Sleep(10 * time.Second)

	// fmt.Println("Time: ", currentTime())

	_, spanResponse := tracer.Start(ctx, "Encoding-Data")
	spanId = spanResponse.SpanContext().SpanID().String()
	traceID = spanResponse.SpanContext().TraceID().String()
	fmt.Println("Trace ID : ", traceID, "Span ID : ", spanId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Status 200
	json.NewEncoder(w).Encode(newItem)
	defer spanResponse.End()
}

func getItem(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Path[len("/items/"):]

	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid Item Id", http.StatusBadRequest)
		return
	}

	fmt.Println(idStr, " ", id)

	mu.Lock()
	fmt.Println("Time Entered in to Critical Section :", currentTime())
	item, exist := items[id]
	mu.Unlock()

	if !exist {
		http.Error(w, "Items not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func updateItem(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Path[len("/updatedItem/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, " Invaid Request ", http.StatusBadRequest)
		return
	}
	var updatedItem Item
	err = json.NewDecoder(r.Body).Decode(&updatedItem)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	_, exists := items[id]

	if !exists {
		http.Error(w, "Id is not present", http.StatusBadRequest)
		mu.Unlock()
		return
	}
	items[id] = updatedItem
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(updatedItem)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {

	idstr := r.URL.Path[len("/deleteItem/"):]

	id, err := strconv.Atoi(idstr)

	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	_, exist := items[id]

	if !exist {
		http.Error(w, " Data is not present ", http.StatusBadRequest)
		return
	}

	mu.Lock()
	delete(items, id)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode("Given ID is deleted")

}

func main() {

	tp, err := trace.StartTracing()
	if err != nil {
		log.Fatalf("failed to initialize tracing: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}()

	http.HandleFunc("/", handler)
	http.HandleFunc("/createItem", createitem)
	http.HandleFunc("/items/", getItem)
	http.HandleFunc("/updatedItem/", updateItem)
	http.HandleFunc("/deleteItem/", deleteItem)

	for key, value := range items {
		fmt.Println(key, " ", value.Value)
	}

	fmt.Println("Starting server on :8080..")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error in starting the server", err)
	}
}
