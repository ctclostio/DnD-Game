package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMProvider defines the interface for Large Language Model providers
type LLMProvider interface {
	GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error)
	GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error)
}

// AIConfig holds configuration for AI services
type AIConfig struct {
	Provider string
	APIKey   string
	Model    string
	Enabled  bool
}

// LLMRequest represents a request to the LLM
type LLMRequest struct {
	Prompt       string
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
}

// LLMResponse represents a response from the LLM
type LLMResponse struct {
	Content string
	Error   error
}

// OpenAIProvider implements LLMProvider using OpenAI's API
type OpenAIProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI LLM provider
func NewOpenAIProvider(apiKey string, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateCompletion sends a request to OpenAI and returns the completion
func (p *OpenAIProvider) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature":     0.7,
		"max_tokens":      2000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateContent is an alias for GenerateCompletion
func (p *OpenAIProvider) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return p.GenerateCompletion(ctx, prompt, systemPrompt)
}

// AnthropicProvider implements LLMProvider using Anthropic's Claude API
type AnthropicProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicProvider creates a new Anthropic LLM provider
func NewAnthropicProvider(apiKey string, model string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateCompletion sends a request to Anthropic and returns the completion
func (p *AnthropicProvider) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 2000,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"system":      systemPrompt,
		"temperature": 0.7,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no response from Anthropic")
	}

	return response.Content[0].Text, nil
}

// GenerateContent is an alias for GenerateCompletion
func (p *AnthropicProvider) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return p.GenerateCompletion(ctx, prompt, systemPrompt)
}

// OpenRouterProvider implements LLMProvider using OpenRouter's API
type OpenRouterProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenRouterProvider creates a new OpenRouter LLM provider
func NewOpenRouterProvider(apiKey string, model string) *OpenRouterProvider {
	return &OpenRouterProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateCompletion sends a request to OpenRouter and returns the completion
func (p *OpenRouterProvider) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/ctclostio/DnD-Game")
	req.Header.Set("X-Title", "D&D Custom Race Generator")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenRouter")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateContent is an alias for GenerateCompletion
func (p *OpenRouterProvider) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return p.GenerateCompletion(ctx, prompt, systemPrompt)
}

// MockLLMProvider for testing
type MockLLMProvider struct {
	Response string
	Error    error
}

// GenerateCompletion returns a mock response
func (m *MockLLMProvider) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}

	// Return a sample balanced race for testing
	if m.Response == "" {
		return `{
			"name": "Crystalborn",
			"description": "Beings of living crystal that emerged from deep underground caverns. Their bodies shimmer with inner light and resonate with magical energy.",
			"abilityScoreIncreases": {"constitution": 2, "intelligence": 1},
			"size": "Medium",
			"speed": 30,
			"traits": [
				{"name": "Crystal Body", "description": "You have resistance to poison damage and advantage on saving throws against being poisoned."},
				{"name": "Resonant Mind", "description": "You can communicate telepathically with any creature within 30 feet that shares a language with you."},
				{"name": "Living Construct", "description": "You don't need to eat, drink, or breathe. You are immune to disease."}
			],
			"languages": ["Common", "Terran"],
			"darkvision": 60,
			"resistances": ["poison"],
			"immunities": [],
			"skillProficiencies": ["Arcana"],
			"toolProficiencies": [],
			"weaponProficiencies": [],
			"armorProficiencies": [],
			"balanceScore": 6,
			"balanceExplanation": "This race is well-balanced with defensive abilities offset by no offensive bonuses. The telepathy is limited in range and the construct traits are mainly flavor."
		}`, nil
	}

	return m.Response, nil
}

// GenerateContent is an alias for GenerateCompletion
func (m *MockLLMProvider) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return m.GenerateCompletion(ctx, prompt, systemPrompt)
}

// NewLLMProvider creates a new LLM provider based on configuration
func NewLLMProvider(config AIConfig) LLMProvider {
	if !config.Enabled {
		return &MockLLMProvider{}
	}

	switch config.Provider {
	case "openai":
		return NewOpenAIProvider(config.APIKey, config.Model)
	case "anthropic":
		return NewAnthropicProvider(config.APIKey, config.Model)
	default:
		// Default to mock provider if unknown
		return &MockLLMProvider{}
	}
}
