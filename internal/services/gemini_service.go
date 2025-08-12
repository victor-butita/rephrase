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

type GeminiService struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewGeminiService(apiKey string) *GeminiService {
	return &GeminiService{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 90 * time.Second},
	}
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GenerationConfig struct {
	Temperature     float32 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiRequest struct {
	Contents         []GeminiContent   `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content      GeminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
	} `json:"candidates"`
}

type AIDetectionResult struct {
	OverallScore int      `json:"overall_score"`
	Analysis     string   `json:"analysis"`
	RedFlags     []string `json:"red_flags"`
}

type PlagiarismMatch struct {
	MatchingText    string  `json:"matching_text"`
	PotentialSource string  `json:"potential_source"`
	Confidence      float32 `json:"confidence"`
}

type PlagiarismResult struct {
	IsSimilarityFound bool              `json:"is_similarity_found"`
	OverallConfidence float32           `json:"overall_confidence"`
	Matches           []PlagiarismMatch `json:"matches"`
}

type ResearchResult struct {
	Topic                     string   `json:"topic"`
	ExecutiveSummary          string   `json:"executive_summary"`
	HistoricalContext         string   `json:"historical_context"`
	CoreConcepts              []string `json:"core_concepts"`
	ControversiesAndCritiques []string `json:"controversies_and_critiques"`
	PracticalApplications     []string `json:"practical_applications"`
}

func (s *GeminiService) RephraseText(text, tone, complexity, dialect, freezeKeywords string) (string, error) {
	promptBuilder := strings.Builder{}
	promptBuilder.WriteString("You are a world-class senior editor and copywriter. Your task is to perform a deep rewrite of the following text based on a strict set of directives. Your goal is not a simple rephrasing, but a professional transformation of the content.\n\n# DIRECTIVES:\n")
	promptBuilder.WriteString(fmt.Sprintf("1.  **Tone & Voice:** The final text must embody a '%s' tone. It should be consistent and professionally executed.\n", tone))
	promptBuilder.WriteString(fmt.Sprintf("2.  **Audience Complexity:** The vocabulary, sentence structure, and concepts must be precisely calibrated for a '%s' audience.\n", complexity))
	promptBuilder.WriteString("3.  **Clarity and Flow:** Rewrite for maximum clarity. Eliminate jargon, passive voice, and redundant phrases. Ensure sentences and paragraphs transition logically.\n")

	if dialect != "" && dialect != "American English (Default)" {
		promptBuilder.WriteString(fmt.Sprintf("4.  **Dialect:** The output must strictly adhere to %s spelling, grammar, and idioms.\n", dialect))
	}

	if strings.TrimSpace(freezeKeywords) != "" {
		promptBuilder.WriteString(fmt.Sprintf(
			"5.  **Keyword Integrity (Non-negotiable):** The following keywords/phrases are mission-critical and MUST appear in the final text exactly as written, without any modification: [%s].\n",
			freezeKeywords,
		))
	}

	promptBuilder.WriteString("\n# OUTPUT FORMAT:\n- Your response MUST be ONLY the rewritten text.\n- DO NOT include any preamble, headers, notes, or explanations (e.g., 'Here is the rewritten text:'). Your entire output will be the final, polished text and nothing else.\n\n")
	promptBuilder.WriteString(fmt.Sprintf("# ORIGINAL TEXT TO REWRITE:\n---\n%s\n---", text))

	return s.generateContent(promptBuilder.String(), 4096, 0.7) // Higher temp for creative rewrite
}

func (s *GeminiService) DetectAI(text string) (*AIDetectionResult, error) {
	prompt := fmt.Sprintf(`
You are a forensic linguistic analysis tool. Your sole function is to analyze text for statistical markers and patterns indicative of generative AI authorship.

Your analysis must be based on the following criteria:
- **Lexical Diversity:** Is the vocabulary unusually complex or simplistic?
- **Syntactic Patterns:** Are sentence structures repetitive? Is there an over-reliance on certain transitional phrases?
- **Content Vacuity:** Does the text contain generic, non-committal statements or lack specific, verifiable details?
- **Unnatural Phrasing:** Are there any awkward word choices or idioms that a native speaker would find odd?
- **Uniformity:** Is the tone and quality perfectly consistent, lacking the typical variance of human writing?

You MUST respond with ONLY a valid, minified JSON object. Do not include any explanation or markdown code fences. The JSON schema is non-negotiable.

JSON Schema:
{
  "overall_score": <int, 0-100, your confidence score that the text is AI-generated>,
  "analysis": "<string, a brief 1-2 sentence summary of your reasoning for the score>",
  "red_flags": [<string, a list of specific phrases or sentences from the text that most strongly support your analysis>]
}

If no strong red flags are found, return an empty array for "red_flags".

Text for Forensic Analysis:
---
%s
---
`, text)

	var result AIDetectionResult
	err := s.generateStructuredContent(prompt, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get or parse AI detection result: %w", err)
	}
	return &result, nil
}

func (s *GeminiService) CheckPlagiarism(text string) (*PlagiarismResult, error) {
	prompt := fmt.Sprintf(`
You are an internal text auditing service. Your purpose is to perform a semantic similarity check on the provided text against your internal knowledge base (your training data). This is NOT a live web search. Your goal is to identify passages with high semantic overlap to known sources, suggesting potential unattributed content.

You MUST respond with ONLY a valid, minified JSON object. Do not use markdown or any explanatory text outside the JSON.

The required JSON schema is as follows:
{
  "is_similarity_found": <bool, true if any significant overlap is detected, else false>,
  "overall_confidence": <float, 0.0-1.0, your confidence in the overall assessment>,
  "matches": [
    {
      "matching_text": "<string, the exact snippet from the input text that shows similarity>",
      "potential_source": "<string, a description of the likely source document or topic from your knowledge base>",
      "confidence": <float, 0.0-1.0, your confidence that this specific snippet is a match>
    }
  ]
}

If no similarities are found, "is_similarity_found" must be false, "overall_confidence" must be low, and "matches" must be an empty array.

Text to Audit:
---
%s
---
`, text)

	var result PlagiarismResult
	err := s.generateStructuredContent(prompt, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get or parse plagiarism result: %w", err)
	}
	return &result, nil
}

func (s *GeminiService) ResearchTopic(topic string) (*ResearchResult, error) {
	prompt := fmt.Sprintf(`
You are a professional research analyst tasked with generating a comprehensive and balanced executive briefing on a given topic. The briefing must be structured, objective, and multi-faceted.

You MUST respond with ONLY a valid, minified JSON object. Do not include markdown or any text outside the JSON structure.

The JSON schema for the executive briefing is as follows:
{
  "topic": "<string, the topic provided>",
  "executive_summary": "<string, a concise 2-3 sentence overview suitable for a busy executive, stating the topic's significance>",
  "historical_context": "<string, a brief explanation of the origin and evolution of the topic>",
  "core_concepts": [<string, a list of the fundamental principles, technologies, or ideas that define the topic>],
  "controversies_and_critiques": [<string, a list of the primary debates, opposing viewpoints, or criticisms related to the topic>],
  "practical_applications": [<string, a list of real-world examples, case studies, or uses of the topic>]
}

Generate this executive briefing for the following topic.

Topic:
---
%s
---
`, topic)

	var result ResearchResult
	err := s.generateStructuredContent(prompt, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get or parse research result: %w", err)
	}
	return &result, nil
}

func (s *GeminiService) generateStructuredContent(prompt string, target interface{}) error {
	// For structured data, we use a lower temperature for more predictable, deterministic output.
	responseText, err := s.generateContent(prompt, 8192, 0.2)
	if err != nil {
		return fmt.Errorf("gemini API call failed: %w", err)
	}

	cleanJSON := strings.TrimSpace(responseText)
	if strings.HasPrefix(cleanJSON, "```json") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(cleanJSON, "```")
		cleanJSON = strings.TrimSpace(cleanJSON)
	} else if strings.HasPrefix(cleanJSON, "```") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```")
		cleanJSON = strings.TrimSuffix(cleanJSON, "```")
		cleanJSON = strings.TrimSpace(cleanJSON)
	}

	if err := json.Unmarshal([]byte(cleanJSON), target); err != nil {
		log.Printf("Failed to unmarshal JSON. Raw response from AI was: %s", responseText)
		return fmt.Errorf("could not parse structured response from AI: %w", err)
	}

	return nil
}

func (s *GeminiService) generateContent(prompt string, maxTokens int, temperature float32) (string, error) {
	config := &GenerationConfig{
		Temperature:     temperature,
		MaxOutputTokens: maxTokens,
	}
	safetySettings := []SafetySetting{
		{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
	}

	reqBody := GeminiRequest{
		Contents:         []GeminiContent{{Parts: []GeminiPart{{Text: prompt}}}},
		GenerationConfig: config,
		SafetySettings:   safetySettings,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling request body: %w", err)
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + s.APIKey

	var resp *http.Response
	maxRetries := 4
	backoffDuration := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("error creating http request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = s.HTTPClient.Do(req)
		if err != nil {
			log.Printf("Attempt %d/%d: Network error during Gemini request: %v. Retrying in %v...", i+1, maxRetries, err, backoffDuration)
			time.Sleep(backoffDuration)
			backoffDuration *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			log.Printf("Attempt %d/%d: Gemini API is busy (status %d). Retrying in %v...", i+1, maxRetries, resp.StatusCode, backoffDuration)
			resp.Body.Close()
			time.Sleep(backoffDuration)
			backoffDuration *= 2
			continue
		}

		break
	}

	if resp == nil {
		return "", fmt.Errorf("gemini API did not respond after %d retries", maxRetries)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading successful response body: %w", err)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		log.Printf("Failed to unmarshal Gemini's main response object. Raw response: %s", string(respBody))
		return "", fmt.Errorf("error parsing Gemini response wrapper: %w", err)
	}

	if len(geminiResp.Candidates) > 0 {
		candidate := geminiResp.Candidates[0]
		if candidate.FinishReason != "STOP" && candidate.FinishReason != "MAX_TOKENS" {
			return "", fmt.Errorf("text generation stopped for an unexpected reason: %s", candidate.FinishReason)
		}
		if len(candidate.Content.Parts) > 0 {
			return candidate.Content.Parts[0].Text, nil
		}
	}

	return "", fmt.Errorf("no content found in Gemini response")
}
