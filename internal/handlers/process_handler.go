package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/victor-butita/rephrase/internal/services" // Use your module path
)

// ProcessHandler now handles all text processing requests
type ProcessHandler struct {
	GeminiService *services.GeminiService
}

func NewProcessHandler(gs *services.GeminiService) *ProcessHandler {
	return &ProcessHandler{
		GeminiService: gs,
	}
}

// APIRequest now includes the action the user wants to perform
type APIRequest struct {
	Text       string `json:"text"`
	Action     string `json:"action"` // "humanize", "detect", or "plagiarize"
	Tone       string `json:"tone,omitempty"`
	Complexity string `json:"complexity,omitempty"`
}

// APIResponse is now more flexible to handle different results
type APIResponse struct {
	ResultType      string `json:"result_type"`
	Text            string `json:"text,omitempty"`
	AiScore         int    `json:"ai_score,omitempty"`
	PlagiarismReport string `json:"plagiarism_report,omitempty"`
	Error           string `json:"error,omitempty"`
}

func (h *ProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var reqData APIRequest
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		h.writeError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// **NEW:** Backend word count validation
	if len(strings.Fields(reqData.Text)) > 200 {
		h.writeError(w, "Input text exceeds the 200-word limit.", http.StatusBadRequest)
		return
	}

	// **NEW:** Switch based on the requested action
	switch reqData.Action {
	case "humanize":
		h.handleHumanize(w, reqData)
	case "detect":
		h.handleDetect(w, reqData)
	case "plagiarize":
		h.handlePlagiarize(w, reqData)
	default:
		h.writeError(w, "Invalid action specified", http.StatusBadRequest)
	}
}

func (h *ProcessHandler) handleHumanize(w http.ResponseWriter, reqData APIRequest) {
	rewrittenText, err := h.GeminiService.RephraseText(reqData.Text, reqData.Tone, reqData.Complexity)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, APIResponse{ResultType: "humanize", Text: rewrittenText}, http.StatusOK)
}

func (h *ProcessHandler) handleDetect(w http.ResponseWriter, reqData APIRequest) {
	score, err := h.GeminiService.DetectAIPercentage(reqData.Text)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, APIResponse{ResultType: "detect", AiScore: score}, http.StatusOK)
}

func (h *ProcessHandler) handlePlagiarize(w http.ResponseWriter, reqData APIRequest) {
	report, err := h.GeminiService.CheckPlagiarism(reqData.Text)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, APIResponse{ResultType: "plagiarize", PlagiarismReport: report}, http.StatusOK)
}

// Helper functions (unchanged, but now use the more flexible APIResponse)
func (h *ProcessHandler) writeJSON(w http.ResponseWriter, data APIResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
func (h *ProcessHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	h.writeJSON(w, APIResponse{Error: message}, statusCode)
}