package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

type LLMDataGenerator struct {
	apiKey     string
	apiURL     string
	model      string
	httpClient *http.Client
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func NewLLMDataGenerator(apiKey string) *LLMDataGenerator {
	return &LLMDataGenerator{
		apiKey: apiKey,
		apiURL: "https://api.openai.com/v1/chat/completions",
		model:  "gpt-4",
		httpClient: &http.Client{
			Timeout: time.Minute * 2,
		},
	}
}

func (ldg *LLMDataGenerator) GenerateAttackPatterns(attackType string, count int) ([]AttackPattern, error) {
	prompt := fmt.Sprintf(`
    Generate %d realistic %s attack patterns for cybersecurity training.
    
    For each pattern, provide:
    - signature: regex pattern to detect the attack
    - payload_examples: 3 example payloads
    - severity: 1-10 severity score
    - description: brief description of the attack
    - mitigation: suggested mitigation strategy
    
    Make patterns realistic and varied.
    Return as JSON array.
    `, count, attackType)

	response, err := ldg.callOpenAI(prompt)
	if err != nil {
		return nil, err
	}

	var patterns []AttackPattern
	if err := json.Unmarshal([]byte(response), &patterns); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return patterns, nil
}

func (ldg *LLMDataGenerator) callOpenAI(prompt string) (string, error) {
	request := OpenAIRequest{
		Model: ldg.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a cybersecurity expert generating synthetic training data.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   2000,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", ldg.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ldg.apiKey))

	resp, err := ldg.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// Enhanced threat signature loading using LLM-generated patterns
func (ips *IntrusionPreventionSystem) loadThreatSignaturesFromLLM() error {
	generator := NewLLMDataGenerator(os.Getenv("OPENAI_API_KEY"))

	attackTypes := []string{"sql_injection", "xss", "command_injection", "directory_traversal"}

	for _, attackType := range attackTypes {
		patterns, err := generator.GenerateAttackPatterns(attackType, 10)
		if err != nil {
			log.Printf("Failed to generate patterns for %s: %v", attackType, err)
			continue
		}

		for _, pattern := range patterns {
			signature := ThreatSignature{
				ID:          fmt.Sprintf("LLM_%s_%d", attackType, len(ips.threatDetector.signatures)),
				Name:        pattern.Name,
				Pattern:     pattern.Signature,
				Protocol:    "HTTP",
				Severity:    pattern.Severity,
				Action:      "BLOCK",
				Description: pattern.Description,
				Regex:       regexp.MustCompile(pattern.Signature),
			}

			ips.threatDetector.signatures = append(ips.threatDetector.signatures, signature)
		}
	}

	return nil
}
