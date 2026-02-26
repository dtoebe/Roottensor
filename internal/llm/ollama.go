package llm

import (
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
