package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// GeminiService holds the configuration for the Gemini API client.
type GeminiService struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewGeminiService creates a new service with the provided API key.
func NewGeminiService(apiKey string) *GeminiService {
	return &GeminiService{
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// --- Structs for Gemini API ---
type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role"`
}

type GeminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type GenerationConfig struct {
	Temperature     float32 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

// **NEW:** This struct defines a single safety setting.
type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// **UPDATED:** The main request struct now includes SafetySettings.
type GeminiRequest struct {
	Contents         []GeminiContent   `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

// --- Core Service Functions ---
// (RephraseText, DetectAIPercentage, and CheckPlagiarism functions are unchanged)
func (s *GeminiService) RephraseText(text, tone, complexity string) (string, error) {
	prompt := fmt.Sprintf(
		"You are an expert editor. Rewrite the following text to make it sound more natural, engaging, and human-written. "+
			"Adopt a %s tone and ensure the reading complexity is appropriate for a %s audience. "+
			"Do not add any preamble, introduction, or concluding remarks. Respond ONLY with the rewritten text. "+
			"Original text:\n\n\"%s\"",
		tone, complexity, text,
	)
	return s.generateContent(prompt)
}

func (s *GeminiService) DetectAIPercentage(text string) (int, error) {
	prompt := fmt.Sprintf(
		"Analyze the following text and determine the likelihood that it was written by an AI. "+
			"Provide a percentage from 0 to 100, where 0 is definitively human-written and 100 is definitively AI-generated. "+
			"Respond with ONLY the numerical percentage, without the '%%' sign or any other text. For example: 85. "+
			"Text to analyze:\n\n\"%s\"",
		text,
	)
	responseText, err := s.generateContent(prompt)
	if err != nil { return 0, err }
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(responseText)
	if match == "" { return 0, fmt.Errorf("could not extract percentage from AI response: %s", responseText) }
	percentage, err := strconv.Atoi(match)
	if err != nil { return 0, fmt.Errorf("could not convert extracted text to a number: %w", err) }
	return percentage, nil
}

func (s *GeminiService) CheckPlagiarism(text string) (string, error) {
	prompt := fmt.Sprintf(
		"You are a plagiarism detection tool. Analyze the following text and search for sentences or phrases that are highly similar to existing content on the public internet. "+
			"If no significant plagiarism is found, respond with ONLY the word 'UNIQUE'. "+
			"If potential plagiarism is found, respond with a list of the potentially plagiarized phrases and the likely source URL. "+
			"Format the response as: 'MATCH: [plagiarized phrase] | SOURCE: [URL]'. List each match on a new line. "+
			"Text to analyze:\n\n\"%s\"",
		text,
	)
	return s.generateContent(prompt)
}

// **UPDATED:** The generateContent function now adds safety settings to the request.
func (s *GeminiService) generateContent(prompt string) (string, error) {
	config := &GenerationConfig{
		Temperature:     0.7,
		MaxOutputTokens: 1024,
	}
	
	// **NEW:** Define safety settings to be less restrictive.
	safetySettings := []SafetySetting{
		{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
	}

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: []GeminiPart{{Text: prompt}}},
		},
		GenerationConfig: config,
		SafetySettings:   safetySettings, // Attach safety settings to the request
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error creating request body: %w", err)
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + s.APIKey

	// (The retry logic below is unchanged)
	var resp *http.Response
	maxRetries := 4
	backoffDuration := 1 * time.Second
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil { return "", fmt.Errorf("error creating http request: %w", err) }
		req.Header.Set("Content-Type", "application/json")
		resp, err = s.HTTPClient.Do(req)
		if err != nil {
			log.Printf("Attempt %d: Network error calling Gemini: %v. Retrying in %v...", i+1, err, backoffDuration)
			time.Sleep(backoffDuration)
			backoffDuration *= 2
			continue
		}
		if resp.StatusCode == http.StatusServiceUnavailable {
			log.Printf("Attempt %d: Gemini API is overloaded (503). Retrying in %v...", i+1, backoffDuration)
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

	// **UPDATED:** Add logging for the finish reason for better debugging
	if len(geminiResp.Candidates) > 0 {
		log.Printf("Gemini Finish Reason: %s", geminiResp.Candidates[0].FinishReason)
		if len(geminiResp.Candidates[0].Content.Parts) > 0 {
			return geminiResp.Candidates[0].Content.Parts[0].Text, nil
		}
		// If the model finished for a reason other than "STOP" (like "SAFETY") and returned no content.
		if geminiResp.Candidates[0].FinishReason != "STOP" {
			return "", fmt.Errorf("text generation was stopped early by the API. Reason: %s", geminiResp.Candidates[0].FinishReason)
		}
	}

	return "", fmt.Errorf("no content found in Gemini response")
}