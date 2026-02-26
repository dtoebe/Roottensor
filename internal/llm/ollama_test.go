package llm

import (
	"reflect"
	"testing"
)

func TestNewOllamaProvider(t *testing.T) {

	t.Run("NewOllamaProvider success", func(t *testing.T) {
		wantBaseUrl := "http://example.com"
		wantModel := "fooModel"
		olm := NewOllamaProvider(wantBaseUrl, wantModel)

		if olm == nil {
			t.Fatal("OllamaProvider is nil")
		}

		if olm.baseUrl != wantBaseUrl {
			t.Errorf("baseUrl: got: %s; want: %s", olm.baseUrl, wantBaseUrl)
		}

		if olm.model != wantModel {
			t.Errorf("model: got: %s; want: %s", olm.model, wantModel)
		}

		if reflect.TypeOf(olm.client).String() != "*http.Client" {
			t.Errorf("type of client: got: %s; want: %s",
				reflect.TypeOf(olm.client).String(), "*http.Client")
		}
	})

	t.Run("NewOllamaProvider default baseUrl", func(t *testing.T) {
		wantBaseUrl := "http://localhost:11434"
		olm := NewOllamaProvider("", "fooModel")

		if olm == nil {
			t.Fatal("OllamaProvider is nil")
		}

		if olm.baseUrl != wantBaseUrl {
			t.Fatalf("baseUrl: got: %s; want: %s", olm.baseUrl, wantBaseUrl)
		}
	})

	t.Run("NewOllamaProvider default model", func(t *testing.T) {
		wantModel := "deepseek-r1:8b"
		olm := NewOllamaProvider("", "")

		if olm == nil {
			t.Fatal("OllamaProvider is nil")
		}

		if olm.model != wantModel {
			t.Fatalf("model: got: %s; want: %s", olm.model, wantModel)
		}
	})
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
