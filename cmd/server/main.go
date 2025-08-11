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
	err := godotenv.Load()
	if err != nil { log.Println("No .env file found, reading from environment") }
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" { log.Fatal("GEMINI_API_KEY environment variable is not set") }

	geminiService := services.NewGeminiService(geminiAPIKey)
	
	// **UPDATED:** Use the new ProcessHandler
	processHandler := handlers.NewProcessHandler(geminiService)

	mux := http.NewServeMux()
	// **UPDATED:** The single endpoint for all actions
	mux.Handle("/api/process", processHandler)
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	port := "8080"
	fmt.Printf("Starting Rephrase server on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}