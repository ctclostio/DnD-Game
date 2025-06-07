package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestOpenAIProvider_GenerateCompletion(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/v1/chat/completions", r.URL.Path)
			require.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			
			// Verify request body
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			require.Equal(t, "gpt-4", reqBody["model"])
			require.Equal(t, float64(0.7), reqBody["temperature"])
			require.Equal(t, float64(2000), reqBody["max_tokens"])
			
			messages := reqBody["messages"].([]interface{})
			require.Len(t, messages, 2)
			
			systemMsg := messages[0].(map[string]interface{})
			require.Equal(t, "system", systemMsg["role"])
			require.Equal(t, "You are a helpful assistant", systemMsg["content"])
			
			userMsg := messages[1].(map[string]interface{})
			require.Equal(t, "user", userMsg["role"])
			require.Equal(t, "Generate a custom race", userMsg["content"])
			
			// Send response
			response := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]string{
							"content": `{"name": "Test Race", "description": "A test race"}`,
						},
					},
				},
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		// Create provider with test server
		provider := NewOpenAIProvider("test-api-key", "gpt-4")
		provider.httpClient = &http.Client{Timeout: 5 * time.Second}
		
		// Override API URL for testing
		originalURL := "https://api.openai.com/v1/chat/completions"
		defer func() { _ = originalURL }() // Avoid unused variable warning
		
		// Monkey patch the URL (in real code, make this configurable)
		ctx := testutil.TestContext()
		prompt := "Generate a custom race"
		systemPrompt := "You are a helpful assistant"
		
		// For this test, we'll need to modify the provider to accept a custom URL
		// Since we can't modify the hardcoded URL, we'll test the actual API behavior
		// by mocking the HTTP client instead
		
		// Create a custom HTTP client that redirects to our test server
		provider.httpClient = &http.Client{
			Transport: &testTransport{
				testServer: server,
			},
			Timeout: 5 * time.Second,
		}
		
		result, err := provider.GenerateCompletion(ctx, prompt, systemPrompt)
		
		require.NoError(t, err)
		require.Equal(t, `{"name": "Test Race", "description": "A test race"}`, result)
	})

	t.Run("API error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
		}))
		defer server.Close()
		
		provider := NewOpenAIProvider("invalid-key", "gpt-4")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "OpenAI API error")
		require.Contains(t, err.Error(), "status 400")
		require.Contains(t, err.Error(), "Invalid API key")
	})

	t.Run("network error", func(t *testing.T) {
		provider := NewOpenAIProvider("test-key", "gpt-4")
		provider.httpClient = &http.Client{
			Transport: &failingTransport{},
			Timeout:   1 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "failed to send request")
	})

	t.Run("invalid response format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"invalid": "response"}`))
		}))
		defer server.Close()
		
		provider := NewOpenAIProvider("test-key", "gpt-4")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "no response from OpenAI")
	})

	t.Run("empty choices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"choices": []interface{}{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		provider := NewOpenAIProvider("test-key", "gpt-4")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "no response from OpenAI")
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		
		provider := NewOpenAIProvider("test-key", "gpt-4")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
	})
}

func TestAnthropicProvider_GenerateCompletion(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/v1/messages", r.URL.Path)
			require.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
			require.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			
			// Verify request body
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			require.Equal(t, "claude-3-opus", reqBody["model"])
			require.Equal(t, "System prompt", reqBody["system"])
			require.Equal(t, float64(0.7), reqBody["temperature"])
			require.Equal(t, float64(2000), reqBody["max_tokens"])
			
			messages := reqBody["messages"].([]interface{})
			require.Len(t, messages, 1)
			
			userMsg := messages[0].(map[string]interface{})
			require.Equal(t, "user", userMsg["role"])
			require.Equal(t, "User prompt", userMsg["content"])
			
			// Send response
			response := map[string]interface{}{
				"content": []map[string]string{
					{"text": "Generated content from Claude"},
				},
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		provider := NewAnthropicProvider("test-api-key", "claude-3-opus")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "User prompt", "System prompt")
		
		require.NoError(t, err)
		require.Equal(t, "Generated content from Claude", result)
	})

	t.Run("API error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"type": "authentication_error", "message": "Invalid API key"}}`))
		}))
		defer server.Close()
		
		provider := NewAnthropicProvider("invalid-key", "claude-3-opus")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "Anthropic API error")
		require.Contains(t, err.Error(), "status 401")
	})

	t.Run("empty content response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"content": []interface{}{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		provider := NewAnthropicProvider("test-key", "claude-3-opus")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "no response from Anthropic")
	})
}

func TestOpenRouterProvider_GenerateCompletion(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/api/v1/chat/completions", r.URL.Path)
			require.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			require.Equal(t, "https://github.com/your-username/dnd-game", r.Header.Get("HTTP-Referer"))
			require.Equal(t, "D&D Custom Race Generator", r.Header.Get("X-Title"))
			
			// Verify request body
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			require.Equal(t, "openai/gpt-4", reqBody["model"])
			
			// Send response
			response := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]string{
							"content": "Generated content from OpenRouter",
						},
					},
				},
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		provider := NewOpenRouterProvider("test-api-key", "openai/gpt-4")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "User prompt", "System prompt")
		
		require.NoError(t, err)
		require.Equal(t, "Generated content from OpenRouter", result)
	})

	t.Run("rate limit error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limit exceeded"}`))
		}))
		defer server.Close()
		
		provider := NewOpenRouterProvider("test-key", "openai/gpt-4")
		provider.httpClient = &http.Client{
			Transport: &testTransport{testServer: server},
			Timeout:   5 * time.Second,
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "test", "test")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Contains(t, err.Error(), "OpenRouter API error")
		require.Contains(t, err.Error(), "status 429")
	})
}

func TestMockLLMProvider(t *testing.T) {
	t.Run("returns custom response", func(t *testing.T) {
		provider := &MockLLMProvider{
			Response: "Custom test response",
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "prompt", "system")
		
		require.NoError(t, err)
		require.Equal(t, "Custom test response", result)
	})

	t.Run("returns error when set", func(t *testing.T) {
		provider := &MockLLMProvider{
			Error: errors.New("mock error"),
		}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "prompt", "system")
		
		require.Error(t, err)
		require.Empty(t, result)
		require.Equal(t, "mock error", err.Error())
	})

	t.Run("returns default race response", func(t *testing.T) {
		provider := &MockLLMProvider{}
		
		ctx := testutil.TestContext()
		result, err := provider.GenerateCompletion(ctx, "prompt", "system")
		
		require.NoError(t, err)
		require.NotEmpty(t, result)
		
		// Verify it's valid JSON
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)
		
		// Verify it contains expected fields
		require.Equal(t, "Crystalborn", parsed["name"])
		require.Equal(t, "Medium", parsed["size"])
		require.Equal(t, float64(30), parsed["speed"])
		require.Equal(t, float64(6), parsed["balanceScore"])
		
		// Verify ability score increases
		asi := parsed["abilityScoreIncreases"].(map[string]interface{})
		require.Equal(t, float64(2), asi["constitution"])
		require.Equal(t, float64(1), asi["intelligence"])
		
		// Verify traits
		traits := parsed["traits"].([]interface{})
		require.Len(t, traits, 3)
	})
}

func TestLLMProviderIntegration(t *testing.T) {
	t.Run("provider interface compatibility", func(t *testing.T) {
		providers := []LLMProvider{
			NewOpenAIProvider("key", "model"),
			NewAnthropicProvider("key", "model"),
			NewOpenRouterProvider("key", "model"),
			&MockLLMProvider{Response: "test"},
		}
		
		ctx := testutil.TestContext()
		
		for _, provider := range providers {
			// Verify each provider implements the interface correctly
			var _ LLMProvider = provider
			
			// For real providers, we'd get errors due to invalid keys
			// But we're just testing interface compliance
			_, _ = provider.GenerateCompletion(ctx, "test", "test")
		}
	})
}

// Test helpers

// testTransport redirects requests to a test server
type testTransport struct {
	testServer *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server
	testURL := t.testServer.URL + req.URL.Path
	testReq, err := http.NewRequest(req.Method, testURL, req.Body)
	if err != nil {
		return nil, err
	}
	
	// Copy headers
	testReq.Header = req.Header
	
	// Use default transport to actually make the request
	return http.DefaultTransport.RoundTrip(testReq)
}

// failingTransport always returns an error
type failingTransport struct{}

func (f *failingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("network error")
}

// Benchmark tests
func BenchmarkOpenAIProvider_GenerateCompletion(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "Benchmark response",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	provider := NewOpenAIProvider("test-key", "gpt-4")
	provider.httpClient = &http.Client{
		Transport: &testTransport{testServer: server},
		Timeout:   5 * time.Second,
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.GenerateCompletion(ctx, "test prompt", "test system")
	}
}

func BenchmarkMockLLMProvider_GenerateCompletion(b *testing.B) {
	provider := &MockLLMProvider{Response: "test response"}
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.GenerateCompletion(ctx, "test prompt", "test system")
	}
}