// Package aiengine implements the 5% of decisions the sprint brief hands to
// phi4:14b via a local Ollama server — never OpenAI, never any cloud
// provider. Every call is cached, budgeted (<300 token prompt, <150 token
// response), and backed by a circuit breaker so a slow or dead Ollama
// degrades the simulation to 100% Rule Engine rather than blocking it.
package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient is a minimal client for Ollama's /api/generate endpoint.
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewOllamaClient(baseURL, model string, timeout time.Duration) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type generateRequest struct {
	Model     string         `json:"model"`
	Prompt    string         `json:"prompt"`
	Stream    bool           `json:"stream"`
	KeepAlive string         `json:"keep_alive"`
	Options   generateOptions `json:"options"`
}

type generateOptions struct {
	// NumPredict caps generated tokens — enforces the sprint brief's "AI
	// response < 150 token" budget at the Ollama layer, not just by prompt
	// convention.
	NumPredict int     `json:"num_predict"`
	Temperature float64 `json:"temperature"`
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Generate sends prompt to the model and returns its raw text response.
// keep_alive is set to 30m so the model stays resident in memory across the
// many calls one simulation run makes, instead of paying a multi-second
// reload cost per call (observed ~7-50s cold-load on this project's dev
// machine — see the Performance section of the sprint report).
func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	// ~150 tokens worst case for short English/Vietnamese decision words;
	// low temperature — this is a decision classifier, not creative writing.
	return c.generate(ctx, prompt, 40, 0.2)
}

// GenerateReport is Generate's counterpart for the one-shot, free-form
// report-writing calls simulation_summary.md/business_recommendation.md
// make (insights package) — a fundamentally different usage pattern from
// the 4 binary decision types (one call per simulation run, not thousands),
// so it gets its own, much larger token budget and slightly higher
// temperature (still low — this must stay a faithful rewrite of supplied
// facts, not creative writing) rather than reusing Generate's 40-token cap.
func (c *OllamaClient) GenerateReport(ctx context.Context, prompt string) (string, error) {
	return c.generate(ctx, prompt, 900, 0.3)
}

func (c *OllamaClient) generate(ctx context.Context, prompt string, numPredict int, temperature float64) (string, error) {
	reqBody := generateRequest{
		Model:     c.model,
		Prompt:    prompt,
		Stream:    false,
		KeepAlive: "30m",
		Options: generateOptions{
			NumPredict:  numPredict,
			Temperature: temperature,
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(b))
	}

	var out generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}
	return out.Response, nil
}

// Ping checks whether the Ollama server is reachable at all (used at
// simulation startup to decide whether to even attempt AI calls this run).
func (c *OllamaClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/version", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama /api/version returned status %d", resp.StatusCode)
	}
	return nil
}
