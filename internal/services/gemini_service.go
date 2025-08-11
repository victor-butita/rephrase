package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// GeminiService and NewGeminiService are unchanged
type GeminiService struct { APIKey string; HTTPClient *http.Client }
func NewGeminiService(apiKey string) *GeminiService { return &GeminiService{APIKey: apiKey, HTTPClient: &http.Client{}} }

// All API and AIDetectionResult structs are unchanged
type GeminiPart struct { Text string `json:"text"` }
type GeminiContent struct { Parts []GeminiPart `json:"parts"`; Role string `json:"role"` }
type GeminiCandidate struct { Content GeminiContent `json:"content"`; FinishReason string `json:"finishReason"` }
type GenerationConfig struct { Temperature float32 `json:"temperature"`; MaxOutputTokens int `json:"maxOutputTokens"` }
type SafetySetting struct { Category string `json:"category"`; Threshold string `json:"threshold"` }
type GeminiRequest struct {
	Contents         []GeminiContent   `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
}
type GeminiResponse struct { Candidates []GeminiCandidate `json:"candidates"` }
type AIDetectionResult struct { OverallScore int `json:"overall_score"`; Sentences []string `json:"sentences"` }

// --- Core Service Functions ---

// **UPDATED:** RephraseText now accepts and uses advanced options.
func (s *GeminiService) RephraseText(text, tone, complexity, dialect, freezeKeywords string) (string, error) {
	// Start with the base prompt
	promptBuilder := strings.Builder{}
	promptBuilder.WriteString(fmt.Sprintf(
		"You are an expert editor. Rewrite the following text to make it sound more natural, engaging, and human-written. "+
			"Adopt a %s tone and ensure the reading complexity is appropriate for a %s audience. ",
		tone, complexity,
	))

	// **NEW:** Dynamically add instructions based on advanced options.
	if dialect != "" && dialect != "American English (Default)" {
		promptBuilder.WriteString(fmt.Sprintf("The rewritten text must be in %s. ", dialect))
	}

	if strings.TrimSpace(freezeKeywords) != "" {
		promptBuilder.WriteString(fmt.Sprintf(
			"Crucially, the following keywords MUST remain unchanged in the final text, exactly as they are written: [%s]. ",
			freezeKeywords,
		))
	}

	promptBuilder.WriteString("Do not add any preamble, introduction, or concluding remarks. Respond ONLY with the rewritten text. ")
	promptBuilder.WriteString(fmt.Sprintf("Original text:\n\n\"%s\"", text))

	return s.generateContent(promptBuilder.String())
}

// DetectAI, CheckPlagiarism, and generateContent are all unchanged from the previous final version.
func (s *GeminiService) DetectAI(text string) (AIDetectionResult, error) {
	prompt := fmt.Sprintf("Analyze the following text... Text to analyze:\n\n\"%s\"", text)
	responseText, err := s.generateContent(prompt)
	if err != nil { return AIDetectionResult{}, err }
	cleanJSONString := strings.TrimSpace(responseText)
	if strings.HasPrefix(cleanJSONString, "```json") {
		cleanJSONString = strings.TrimPrefix(cleanJSONString, "```json")
		cleanJSONString = strings.TrimSuffix(cleanJSONString, "```")
	}
	var result AIDetectionResult
	if err := json.Unmarshal([]byte(cleanJSONString), &result); err != nil {
		return AIDetectionResult{}, fmt.Errorf("could not parse structured AI detection response: %w", err)
	}
	return result, nil
}

func (s *GeminiService) CheckPlagiarism(text string) (string, error) {
	prompt := fmt.Sprintf("You are a plagiarism detection tool... Text to analyze:\n\n\"%s\"", text)
	return s.generateContent(prompt)
}

func (s *GeminiService) ResearchTopic(topic string) (string, error) {
	prompt := fmt.Sprintf("You are a helpful research assistant... Topic:\n\n\"%s\"", topic)
	return s.generateContent(prompt)
}

func (s *GeminiService) generateContent(prompt string) (string, error) {
	config := &GenerationConfig{Temperature: 0.7, MaxOutputTokens: 1024}
	safetySettings := []SafetySetting{
		{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
	}
	reqBody := GeminiRequest{ Contents: []GeminiContent{{Parts: []GeminiPart{{Text: prompt}}}}, GenerationConfig: config, SafetySettings: safetySettings }
	jsonData, err := json.Marshal(reqBody)
	if err != nil { return "", fmt.Errorf("error creating request body: %w", err) }
	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + s.APIKey
	var resp *http.Response
	maxRetries := 4
	backoffDuration := 1 * time.Second
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil { return "", fmt.Errorf("error creating http request: %w", err) }
		req.Header.Set("Content-Type", "application/json")
		resp, err = s.HTTPClient.Do(req)
		if err != nil {
			log.Printf("Attempt %d: Network error... Retrying...", i+1, err, backoffDuration)
			time.Sleep(backoffDuration)
			backoffDuration *= 2
			continue
		}
		if resp.StatusCode == http.StatusServiceUnavailable {
			log.Printf("Attempt %d: Gemini API is overloaded... Retrying...", i+1, backoffDuration)
			resp.Body.Close()
			time.Sleep(backoffDuration)
			backoffDuration *= 2
			continue
		}
		break
	}
	if resp == nil { return "", fmt.Errorf("gemini API did not respond after %d retries", maxRetries) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(respBody))
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil { return "", fmt.Errorf("error reading successful response body: %w", err) }
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("error parsing successful Gemini response: %w", err)
	}
	if len(geminiResp.Candidates) > 0 {
		if len(geminiResp.Candidates[0].Content.Parts) > 0 { return geminiResp.Candidates[0].Content.Parts[0].Text, nil }
		if geminiResp.Candidates[0].FinishReason != "STOP" { return "", fmt.Errorf("text generation was stopped early. Reason: %s", geminiResp.Candidates[0].FinishReason) }
	}
	return "", fmt.Errorf("no content found in Gemini response")
}