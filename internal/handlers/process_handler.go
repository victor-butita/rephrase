package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/victor-butita/rephrase/internal/services" // Use your module path
)

// ProcessHandler and NewProcessHandler are unchanged
type ProcessHandler struct { GeminiService *services.GeminiService; StatsTracker *StatsTracker }
func NewProcessHandler(gs *services.GeminiService, st *StatsTracker) *ProcessHandler { return &ProcessHandler{ GeminiService: gs, StatsTracker: st } }

// **UPDATED:** APIRequest now includes all advanced options
type APIRequest struct {
	Text           string `json:"text"`
	Action         string `json:"action"`
	Tone           string `json:"tone,omitempty"`
	Complexity     string `json:"complexity,omitempty"`
	Dialect        string `json:"dialect,omitempty"`
	FreezeKeywords string `json:"freeze_keywords,omitempty"`
}

// APIResponse is unchanged
type APIResponse struct {
	ResultType       string                           `json:"result_type"`
	Text             string                           `json:"text,omitempty"`
	DetectionResult  *services.AIDetectionResult      `json:"detection_result,omitempty"`
	PlagiarismReport string                           `json:"plagiarism_report,omitempty"`
	ResearchResult   string                           `json:"research_result,omitempty"`
	Error            string                           `json:"error,omitempty"`
}

// ServeHTTP is unchanged
func (h *ProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { h.writeError(w, "Invalid request method", http.StatusMethodNotAllowed); return }
	var reqData APIRequest
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil { h.writeError(w, "Invalid JSON payload", http.StatusBadRequest); return }
	if reqData.Action != "research" && len(strings.Fields(reqData.Text)) > 200 { h.writeError(w, "Input text exceeds the 200-word limit.", http.StatusBadRequest); return }
	h.StatsTracker.Increment(reqData.Action)
	switch reqData.Action {
	case "humanize":
		h.handleHumanize(w, reqData)
	case "detect":
		h.handleDetect(w, reqData)
	case "plagiarize":
		h.handlePlagiarize(w, reqData)
	case "research":
		h.handleResearch(w, reqData)
	default:
		h.writeError(w, "Invalid action specified", http.StatusBadRequest)
	}
}

// **UPDATED:** handleHumanize now passes the advanced options to the service
func (h *ProcessHandler) handleHumanize(w http.ResponseWriter, reqData APIRequest) {
	rewrittenText, err := h.GeminiService.RephraseText(reqData.Text, reqData.Tone, reqData.Complexity, reqData.Dialect, reqData.FreezeKeywords)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, APIResponse{ResultType: "humanize", Text: rewrittenText}, http.StatusOK)
}

// handleDetect, handlePlagiarize, and handleResearch are unchanged
func (h *ProcessHandler) handleDetect(w http.ResponseWriter, reqData APIRequest) {
	result, err := h.GeminiService.DetectAI(reqData.Text)
	if err != nil { h.writeError(w, err.Error(), http.StatusInternalServerError); return }
	h.writeJSON(w, APIResponse{ResultType: "detect", DetectionResult: &result}, http.StatusOK)
}
func (h *ProcessHandler) handlePlagiarize(w http.ResponseWriter, reqData APIRequest) {
	report, err := h.GeminiService.CheckPlagiarism(reqData.Text)
	if err != nil { h.writeError(w, err.Error(), http.StatusInternalServerError); return }
	h.writeJSON(w, APIResponse{ResultType: "plagiarize", PlagiarismReport: report}, http.StatusOK)
}
func (h *ProcessHandler) handleResearch(w http.ResponseWriter, reqData APIRequest) {
	if reqData.Text == "" { h.writeError(w, "Research topic cannot be empty", http.StatusBadRequest); return }
	result, err := h.GeminiService.ResearchTopic(reqData.Text)
	if err != nil { h.writeError(w, err.Error(), http.StatusInternalServerError); return }
	h.writeJSON(w, APIResponse{ResultType: "research", ResearchResult: result}, http.StatusOK)
}

// Helper functions are unchanged
func (h *ProcessHandler) writeJSON(w http.ResponseWriter, data APIResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(statusCode); json.NewEncoder(w).Encode(data)
}
func (h *ProcessHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	h.writeJSON(w, APIResponse{Error: message}, statusCode)
}