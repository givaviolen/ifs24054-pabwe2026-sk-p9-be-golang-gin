package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Searcher mencari fakta terkini di Search Engine.
type Searcher interface {
	Search(ctx context.Context, query string) ([]Source, error)
}

type SearchClient struct {
	Endpoint string
	HTTP     *http.Client
}

func NewSearchClient(endpoint string) *SearchClient {
	return &SearchClient{Endpoint: endpoint, HTTP: &http.Client{Timeout: 8 * time.Second}}
}

type rawItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Link        string `json:"link"`
	Snippet     string `json:"snippet"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

func (r rawItem) toSource() Source {
	s := Source{Title: r.Title, URL: r.URL, Snippet: r.Snippet}
	if s.URL == "" {
		s.URL = r.Link
	}
	if s.Snippet == "" {
		s.Snippet = r.Content
	}
	if s.Snippet == "" {
		s.Snippet = r.Description
	}
	return s
}

// Search memanggil GET {endpoint}?q=... dan menerima {"results":[...]}, {"data":[...]}, atau array langsung.
func (s *SearchClient) Search(ctx context.Context, query string) ([]Source, error) {
	u, err := url.Parse(s.Endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("search engine mengembalikan status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []rawItem
	var wrapped struct {
		Results []rawItem `json:"results"`
		Data    []rawItem `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		items = wrapped.Results
		if len(items) == 0 {
			items = wrapped.Data
		}
	}
	if len(items) == 0 {
		var direct []rawItem
		if err := json.Unmarshal(body, &direct); err == nil {
			items = direct
		}
	}

	sources := make([]Source, 0, 5)
	for _, it := range items {
		if len(sources) == 5 {
			break
		}
		sources = append(sources, it.toSource())
	}
	return sources, nil
}
