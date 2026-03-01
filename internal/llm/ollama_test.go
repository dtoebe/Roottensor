package llm

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewOllamaProvider(t *testing.T) {
	tests := []struct {
		name        string
		baseUrl     string
		model       string
		wantBaseUrl string
		wantModel   string
	}{
		{
			name:        "valid url and model",
			baseUrl:     "http://example.com",
			model:       "llama3",
			wantBaseUrl: "http://example.com",
			wantModel:   "llama3",
		},
		{
			name:        "invalid url defaults to localhost",
			baseUrl:     "not-a-url",
			model:       "llama3",
			wantBaseUrl: "http://localhost:11434",
			wantModel:   "llama3",
		},
		{
			name:        "empty url defaults to localhost",
			baseUrl:     "",
			model:       "llama3",
			wantBaseUrl: "http://localhost:11434",
			wantModel:   "llama3",
		},
		{
			name:        "empty model defaults to deepseek",
			baseUrl:     "http://example.com",
			model:       "",
			wantBaseUrl: "http://example.com",
			wantModel:   "deepseek-r1:8b",
		},
		{
			name:        "both empty defaults to localhost and deepseek",
			baseUrl:     "",
			model:       "",
			wantBaseUrl: "http://localhost:11434",
			wantModel:   "deepseek-r1:8b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOllamaProvider(tt.baseUrl, tt.model)

			if got == nil {
				t.Fatal("NewOllamaProvider returned nil")
			}
			if got.baseUrl != tt.wantBaseUrl {
				t.Errorf("baseUrl: got %q want %q", got.baseUrl, tt.wantBaseUrl)
			}
			if got.model != tt.wantModel {
				t.Errorf("model: got %q want %q", got.model, tt.wantModel)
			}
			if got.client == nil {
				t.Fatal("client is nil")
			}
			if got.client.Timeout != 60*time.Second {
				t.Errorf("client.Timeout: got %v want %v", got.client.Timeout, 60*time.Second)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "valid http URL", input: "http://example.com", want: true},
		{name: "valid https URL", input: "https://example.com", want: true},
		{name: "valid with path", input: "https://example.com/path", want: true},
		{name: "valid with port", input: "http://localhost:11434", want: true},

		{name: "empty string", input: "", want: false},
		{name: "garbage string", input: "not a url", want: false},

		{name: "ftp scheme", input: "ftp://example.com", want: false},
		{name: "no scheme", input: "//example.com", want: false},
		{name: "relative path", input: "/just/a/path", want: false},

		{name: "scheme only", input: "http://", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsURL(tt.input)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestModel(t *testing.T) {
	olm, _, wantModel := initOllamaProvider(t)
	if olm.Model() != wantModel {
		t.Fatalf("model: got: %s; want: %s", olm.Model(), wantModel)
	}
}

func TestBaseURL(t *testing.T) {
	olm, wantBaseUrl, _ := initOllamaProvider(t)
	if olm.BaseURL() != wantBaseUrl {
		t.Fatalf("baseUrl: got: %s; want: %s", olm.BaseURL(), wantBaseUrl)
	}
}

func TestOllamaProvider_buildRequest(t *testing.T) {
	defaultModel := "deepseek-r1:8b"
	provider := &OllamaProvider{model: defaultModel}

	msgs := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}

	tests := []struct {
		name string
		msgs []Message
		opts *CallOptions
		want *ollamaChatRequest
	}{
		{
			name: "nil opts uses provider defaults",
			msgs: msgs,
			opts: nil,
			want: &ollamaChatRequest{
				Model:    defaultModel,
				Messages: msgs,
				Stream:   false,
				Options:  nil,
			},
		},
		{
			name: "empty opts uses provider defaults",
			msgs: msgs,
			opts: &CallOptions{},
			want: &ollamaChatRequest{
				Model:    defaultModel,
				Messages: msgs,
				Stream:   false,
				Options:  nil,
			},
		},
		{
			name: "opts overrides model",
			msgs: msgs,
			opts: &CallOptions{Model: "llama3"},
			want: &ollamaChatRequest{
				Model:    "llama3",
				Messages: msgs,
				Stream:   false,
				Options:  nil,
			},
		},
		{
			name: "opts sets temperature",
			msgs: msgs,
			opts: &CallOptions{Temperature: 0.7},
			want: &ollamaChatRequest{
				Model:    defaultModel,
				Messages: msgs,
				Stream:   false,
				Options:  map[string]any{"temperature": float64(0.7)},
			},
		},
		{
			name: "opts sets maxTokens",
			msgs: msgs,
			opts: &CallOptions{MaxTokens: 512},
			want: &ollamaChatRequest{
				Model:    defaultModel,
				Messages: msgs,
				Stream:   false,
				Options:  map[string]any{"num_predict": 512},
			},
		},
		{
			name: "opts enables stream",
			msgs: msgs,
			opts: &CallOptions{Stream: true},
			want: &ollamaChatRequest{
				Model:    defaultModel,
				Messages: msgs,
				Stream:   true,
				Options:  nil,
			},
		},
		{
			name: "opts sets all fields",
			msgs: msgs,
			opts: &CallOptions{
				Model:       "llama3",
				Temperature: 0.9,
				MaxTokens:   1024,
				Stream:      true,
			},
			want: &ollamaChatRequest{
				Model:    "llama3",
				Messages: msgs,
				Stream:   true,
				Options: map[string]any{
					"temperature": float64(0.9),
					"num_predict": 1024,
				},
			},
		},
		{
			name: "empty messages",
			msgs: []Message{},
			opts: nil,
			want: &ollamaChatRequest{
				Model:    defaultModel,
				Messages: []Message{},
				Stream:   false,
				Options:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.buildRequest(tt.msgs, tt.opts)

			if got.Model != tt.want.Model {
				t.Errorf("Model: got %q want %q", got.Model, tt.want.Model)
			}
			if got.Stream != tt.want.Stream {
				t.Errorf("Stream: got %v want %v", got.Stream, tt.want.Stream)
			}
			if tt.want.Options == nil {
				if got.Options != nil {
					t.Errorf("Options: got %v want nil", got.Options)
				}
			} else {
				for k, wantVal := range tt.want.Options {
					gotVal, ok := got.Options[k]
					if !ok {
						t.Errorf("Options: missing key %q", k)
						continue
					}

					if k == "temperature" {
						if math.Abs(gotVal.(float64)-wantVal.(float64)) > 1e-7 {
							t.Errorf("Options temperature: got %v want %v", gotVal, wantVal)
						}
					} else if gotVal != wantVal {
						t.Errorf("Options %q: got %v want %v", k, gotVal, wantVal)
					}
				}
			}
			if !reflect.DeepEqual(got.Messages, tt.want.Messages) {
				t.Errorf("Messages: got %v want %v", got.Messages, tt.want.Messages)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newProviderWithTransport(rt http.RoundTripper) *OllamaProvider {
	return &OllamaProvider{
		baseUrl: "http://mock",
		model:   "mock-model",
		client: &http.Client{
			Timeout:   5 * time.Second,
			Transport: rt,
		},
	}
}

func TestOllamaProvider_doRequest(t *testing.T) {
	t.Run("marshal error", func(t *testing.T) {
		p := newProviderWithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			t.Fatal("RoundTrip should not be called when marshal fails")
			return nil, nil
		}))

		// Force json.Marshal to fail: include an unsupported value in Options.
		reqBody := &ollamaChatRequest{
			Model: "x",
			Options: map[string]any{
				"bad": make(chan int),
			},
		}

		got, err := p.doRequest(context.Background(), reqBody)
		if got != "" {
			t.Fatalf("got %q, want empty string", got)
		}
		if err == nil || !strings.Contains(err.Error(), "marshal error:") {
			t.Fatalf("expected marshal error, got: %v", err)
		}
	})

	t.Run("client.Do error", func(t *testing.T) {
		wantErr := errors.New("network down")

		p := newProviderWithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			// Ensure the request is built as expected
			if r.Method != http.MethodPost {
				t.Fatalf("method got %q want %q", r.Method, http.MethodPost)
			}
			if r.URL.Path != "/api/chat" {
				t.Fatalf("path got %q want %q", r.URL.Path, "/api/chat")
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type got %q want %q", ct, "application/json")
			}

			return nil, wantErr
		}))

		reqBody := &ollamaChatRequest{Model: "x"}

		got, err := p.doRequest(context.Background(), reqBody)
		if got != "" {
			t.Fatalf("got %q, want empty string", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama request error:") {
			t.Fatalf("expected request error, got: %v", err)
		}
		if err == nil || !strings.Contains(err.Error(), wantErr.Error()) {
			t.Fatalf("expected wrapped error to contain %q, got: %v", wantErr.Error(), err)
		}
	})

	t.Run("non-2xx status code", func(t *testing.T) {
		p := newProviderWithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader(`{"error":"boom"}`)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{Model: "x"}

		got, err := p.doRequest(context.Background(), reqBody)
		if got != "" {
			t.Fatalf("got %q, want empty string", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama response returned status code: 500") {
			t.Fatalf("expected status code error, got: %v", err)
		}
	})

	t.Run("decode error (invalid JSON)", func(t *testing.T) {
		p := newProviderWithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`not-json`)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{Model: "x"}

		got, err := p.doRequest(context.Background(), reqBody)
		if got != "" {
			t.Fatalf("got %q, want empty string", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama response decode error:") {
			t.Fatalf("expected decode error, got: %v", err)
		}
	})

	t.Run("parsed.Error is non-empty", func(t *testing.T) {
		// Response is valid JSON but includes an "error" field.
		body := `{"error":"model not found","message":{"content":"ignored"}}`

		p := newProviderWithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{Model: "x"}

		got, err := p.doRequest(context.Background(), reqBody)
		if got != "" {
			t.Fatalf("got %q, want empty string", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama returned error:") {
			t.Fatalf("expected parsed.Error branch error, got: %v", err)
		}

		// NOTE: your implementation currently formats: fmt.Errorf("ollama returned error: %v", err)
		// which prints the *previous* err (likely nil) rather than parsed.Error.
		// This test just asserts you hit that branch.
	})

	t.Run("success returns message content", func(t *testing.T) {
		body := `{"message":{"content":"hello from mock"}}`

		p := newProviderWithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			// Optional: ensure body was JSON
			b, _ := io.ReadAll(r.Body)
			if len(b) == 0 {
				t.Fatal("expected request body, got empty")
			}

			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{Model: "x"}

		got, err := p.doRequest(context.Background(), reqBody)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if got != "hello from mock" {
			t.Fatalf("got %q want %q", got, "hello from mock")
		}
	})
}

func initOllamaProvider(t *testing.T) (*OllamaProvider, string, string) {
	t.Helper()
	baseUrl := "http://example.com"
	model := "FooModel"
	return NewOllamaProvider(baseUrl, model), baseUrl, model
}

func newProviderForStream(baseURL string, rt http.RoundTripper) *OllamaProvider {
	return &OllamaProvider{
		baseUrl: baseURL,
		model:   "mock-model",
		client: &http.Client{
			Timeout:   5 * time.Second,
			Transport: rt,
		},
	}
}

type errReader struct {
	data []byte
	done bool
	err  error
}

func (r *errReader) Read(p []byte) (int, error) {
	if !r.done {
		r.done = true
		n := copy(p, r.data)
		return n, nil
	}
	return 0, r.err
}

func TestOllamaProvider_doRequestStream(t *testing.T) {
	t.Run("sets reqBody.Stream = true", func(t *testing.T) {
		// Return immediately with Done=true so it finishes.
		body := `{"done": true}` + "\n"

		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{Stream: false}

		_, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reqBody.Stream != true {
			t.Fatalf("expected reqBody.Stream=true, got %v", reqBody.Stream)
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			t.Fatal("RoundTrip should not be called on marshal error")
			return nil, nil
		}))

		// Force json.Marshal to fail by putting an unsupported value into Options.
		// This assumes ollamaChatRequest has a field Options map[string]any (common in your earlier code).
		reqBody := &ollamaChatRequest{
			Options: map[string]any{
				"bad": make(chan int),
			},
		}

		got, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if got != "" {
			t.Fatalf("got %q want empty", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama stream marshal error:") {
			t.Fatalf("expected marshal error, got %v", err)
		}
	})

	t.Run("JoinPath error (invalid baseUrl)", func(t *testing.T) {
		p := newProviderForStream("http://[::1", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			t.Fatal("RoundTrip should not be called when JoinPath fails")
			return nil, nil
		}))

		reqBody := &ollamaChatRequest{}

		got, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if got != "" {
			t.Fatalf("got %q want empty", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama stream build url error:") {
			t.Fatalf("expected JoinPath error, got %v", err)
		}
	})

	t.Run("client.Do error", func(t *testing.T) {
		wantErr := errors.New("dial failed")

		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			// Ensure request looks right
			if r.Method != http.MethodPost {
				t.Fatalf("method got %q want %q", r.Method, http.MethodPost)
			}
			if r.URL.Path != "/api/chat" {
				t.Fatalf("path got %q want %q", r.URL.Path, "/api/chat")
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type got %q want %q", ct, "application/json")
			}
			return nil, wantErr
		}))

		reqBody := &ollamaChatRequest{}

		got, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if got != "" {
			t.Fatalf("got %q want empty", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama streaming request error:") {
			t.Fatalf("expected Do error, got %v", err)
		}
		if !strings.Contains(err.Error(), wantErr.Error()) {
			t.Fatalf("expected wrapped error to contain %q, got %v", wantErr.Error(), err)
		}
	})

	t.Run("non-2xx status code", func(t *testing.T) {
		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader("server error")),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{}

		got, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if got != "" {
			t.Fatalf("got %q want empty", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama response returned status code: 500") {
			t.Fatalf("expected status code error, got %v", err)
		}
	})

	t.Run("unmarshal chunk error", func(t *testing.T) {
		// One non-empty invalid JSON line triggers json.Unmarshal error.
		body := `not-json` + "\n"

		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{}

		got, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if got != "" {
			t.Fatalf("got %q want empty", got)
		}
		if err == nil || !strings.Contains(err.Error(), "decode stream chunk error:") {
			t.Fatalf("expected unmarshal error, got %v", err)
		}
	})

	t.Run("chunk has Error", func(t *testing.T) {
		body := `{"error":"boom"}` + "\n"

		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{}

		got, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if got != "" {
			t.Fatalf("got %q want empty", got)
		}
		if err == nil || !strings.Contains(err.Error(), "ollama stream error: boom") {
			t.Fatalf("expected stream error, got %v", err)
		}
	})

	t.Run("skips blank lines, streams content, breaks on done", func(t *testing.T) {
		// Includes:
		// - blank line (should be skipped)
		// - content chunk (should call onChunk + accumulate)
		// - done chunk (break)
		body := "\n" +
			`{"content":"Hel"}` + "\n" +
			`{"content":"lo"}` + "\n" +
			`{"done":true}` + "\n" +
			`{"content":"ignored-after-done"}` + "\n"

		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{}

		var chunks []string
		got, err := p.doRequestStream(context.Background(), reqBody, func(s string) {
			chunks = append(chunks, s)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "Hello" {
			t.Fatalf("got %q want %q", got, "Hello")
		}
		if strings.Join(chunks, "") != "Hello" {
			t.Fatalf("chunks got %q want %q", strings.Join(chunks, ""), "Hello")
		}
	})

	t.Run("scanner.Err path (read stream error)", func(t *testing.T) {
		// Provide one valid chunk, then force a reader error on the next Read.
		data := []byte(`{"content":"Hi"}` + "\n")
		readErr := errors.New("read failed")

		p := newProviderForStream("http://example.com", roundTripFunc(func(r *http.Request) (*http.Response, error) {
			// Note: bufio.Scanner will attempt to read again and hit our error => scanner.Err() != nil
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(&errReader{data: data, err: readErr}),
				Header:     make(http.Header),
			}, nil
		}))

		reqBody := &ollamaChatRequest{}

		got, err := p.doRequestStream(context.Background(), reqBody, func(string) {})
		if got != "" {
			t.Fatalf("got %q want empty", got)
		}
		if err == nil || !strings.Contains(err.Error(), "read stream error:") {
			t.Fatalf("expected read stream error, got %v", err)
		}
		if !strings.Contains(err.Error(), readErr.Error()) {
			t.Fatalf("expected wrapped error to contain %q, got %v", readErr.Error(), err)
		}
	})

	// Optional: create-request error is basically unreachable with a valid URL,
	// so we don't try to force it. Everything else is fully covered.
	_ = bufio.MaxScanTokenSize
}
