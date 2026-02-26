package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type OllamaProvider struct {
	baseUrl string
	model   string
	client  *http.Client
}

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

func NewOllamaProvider(baseUrl, model string) *OllamaProvider {
	if !IsURL(baseUrl) {
		baseUrl = "http://localhost:11434"
	}

	if model == "" {
		model = "deepseek-r1:8b"
	}

	return &OllamaProvider{
		baseUrl: baseUrl,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			},
		},
	}
}

// TODO: Move to better place
func IsURL(webUrl string) bool {
	u, err := url.ParseRequestURI(webUrl)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}

	return true
}

func (p *OllamaProvider) Model() string {
	return p.model
}

func (p *OllamaProvider) BaseURL() string {
	return p.baseUrl
}

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type CallOptions struct {
	Temperature float32
	MaxTokens   int
	Stream      bool
	Model       string
}

type ollamaChatRequest struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Message Message `json:"message"`
	Error   string  `json:"error,omitempty"`
	Done    bool    `json:"done"`
	Content string  `json:"content,omitempty"`
}

func (p *OllamaProvider) Chat(
	ctx context.Context,
	msgs []Message,
	opts *CallOptions) (string, error) {
	if opts != nil && opts.Stream {
		var buf bytes.Buffer
		err := p.chatStream(ctx, msgs, opts, func(chunk string) {
			buf.WriteString(chunk)
		})
		if err != nil {
			return "", err
		}

		return buf.String(), nil
	}

	req := p.buildRequest(msgs, opts)
	resp, err := p.doRequest(ctx, req)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (p *OllamaProvider) chatStream(
	ctx context.Context,
	msgs []Message,
	opts *CallOptions,
	onChunk func(string)) error {
	if onChunk == nil {
		return errors.New("onChunk callback cannot be nil")
	}

	req := p.buildRequest(msgs, opts)
	_, err := p.doRequest(ctx, req)
	if err != nil {
		return err
	}

	_, err = p.doRequestStream(ctx, req, onChunk)
	return err
}

func (p *OllamaProvider) buildRequest(msgs []Message, opts *CallOptions) *ollamaChatRequest {
	model := p.model
	temp := 0.0
	maxTokens := 0
	isStream := false

	if opts != nil {
		if opts.Model != "" {
			model = opts.Model
		}
		if opts.Temperature != 0 {
			temp = float64(opts.Temperature)
		}
		if opts.MaxTokens != 0 {
			maxTokens = opts.MaxTokens
		}
		if opts.Stream {
			isStream = opts.Stream
		}
	}

	ollamaMsgs := make([]Message, 0, len(msgs))
	for _, m := range msgs {
		ollamaMsgs = append(ollamaMsgs, Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	options := map[string]any{}
	if temp != 0 {
		options["temperature"] = temp
	}
	if maxTokens > 0 {
		options["num_predict"] = maxTokens
	}

	if len(options) == 0 {
		options = nil
	}

	return &ollamaChatRequest{
		Model:    model,
		Messages: ollamaMsgs,
		Stream:   isStream,
		Options:  options,
	}
}

func (p *OllamaProvider) doRequest(ctx context.Context, reqBody *ollamaChatRequest) (string, error) {
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal error: %v", err)
	}

	url, err := url.JoinPath(p.baseUrl, "/api/chat")
	if err != nil {
		return "", fmt.Errorf("ollama stream build url error: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("newRequest error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("olloama request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama response returned status code: %d", resp.StatusCode)
	}

	var parsed ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("ollama response decode error: %v", err)
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("ollama returned error: %v", parsed.Error)
	}

	return parsed.Message.Content, nil
}

func (p *OllamaProvider) doRequestStream(
	ctx context.Context,
	reqBody *ollamaChatRequest,
	onChunk func(string)) (string, error) {
	reqBody.Stream = true

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama stream marshal error: %v", err)
	}

	url, err := url.JoinPath(p.baseUrl, "/api/chat")
	if err != nil {
		return "", fmt.Errorf("ollama stream build url error: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("ollama stream create request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama streaming request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama response returned status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var full bytes.Buffer

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var chunk ollamaChatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			return "", fmt.Errorf("decode stream chunk error: %v", err)
		}
		if chunk.Error != "" {
			return "", fmt.Errorf("ollama stream error: %s", chunk.Error)
		}
		if chunk.Content != "" {
			onChunk(chunk.Content)
			full.WriteString(chunk.Content)
		}
		if chunk.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read stream error: %v", err)
	}

	return full.String(), nil
}
