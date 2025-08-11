package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/victor-butita/rephrase/internal/handlers" // Use your module path
	"github.com/victor-butita/rephrase/internal/services" // Use your module path

)

func main() {
	// --- Initialization ---
	err := godotenv.Load()
	if err != nil { log.Println("No .env file found, reading from environment") }
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" { log.Fatal("GEMINI_API_KEY environment variable is not set") }

	// --- Dependency Injection ---
	geminiService := services.NewGeminiService(geminiAPIKey)
	
	// **NEW:** Create the Hub and StatsTracker
	hub := handlers.NewHub()
	statsTracker := handlers.NewStatsTracker(hub)
	
	// **NEW:** Start the Hub's main loop in a goroutine
	go hub.Run(statsTracker)

	// **UPDATED:** Inject the StatsTracker into the ProcessHandler
	processHandler := handlers.NewProcessHandler(geminiService, statsTracker)

	// --- Routing ---
	mux := http.NewServeMux()
	mux.Handle("/api/process", processHandler)
	mux.HandleFunc("/ws", hub.ServeWs) // **NEW:** WebSocket endpoint
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// --- Start Server ---
	port := "8080"
	fmt.Printf("Starting Rephrase AI server on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}