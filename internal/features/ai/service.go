package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"app/internal/features/post"
)

const systemPrompt = `Anda adalah penulis blog profesional berbahasa Indonesia.
Tulis artikel utuh, menarik, dan terstruktur dalam format Markdown (gunakan heading ##, daftar, dan paragraf).
Jika diberikan "KONTEKS PENELUSURAN", gunakan sebagai rujukan fakta terkini dan sebutkan sitasi dengan format [1], [2], dst.
Balas HANYA dengan satu objek JSON tanpa teks lain dan tanpa blok kode, dengan kunci:
"title" (string), "category" (salah satu: Teknologi, Pendidikan, Gaya Hidup, Bisnis, Opini, Umum),
"excerpt" (ringkasan pemikat 1-2 kalimat), "content" (isi artikel Markdown).`

var ErrEmptyOutput = errors.New("respons OpenAI tidak berisi artikel")

type Service struct {
	search Searcher
	llm    Completer
}

func NewService(s Searcher, c Completer) *Service { return &Service{search: s, llm: c} }

func buildPrompt(req GenerateRequest, sources []Source) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topik artikel: %s\n", req.Topic)
	if req.Category != "" {
		fmt.Fprintf(&b, "Kategori yang diinginkan: %s\n", req.Category)
	}
	if len(sources) > 0 {
		b.WriteString("\nKONTEKS PENELUSURAN:\n")
		for i, s := range sources {
			fmt.Fprintf(&b, "[%d] %s (%s)\n%s\n", i+1, s.Title, s.URL, s.Snippet)
		}
	}
	return b.String()
}

// parseArticle mengambil objek JSON dari keluaran model (menoleransi ```json fences).
func parseArticle(raw string) (*GeneratedPost, bool) {
	start, end := strings.Index(raw, "{"), strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil, false
	}
	var g GeneratedPost
	if err := json.Unmarshal([]byte(raw[start:end+1]), &g); err != nil {
		return nil, false
	}
	if strings.TrimSpace(g.Content) == "" {
		return nil, false
	}
	return &g, true
}

// Generate menghasilkan draf artikel; jika pencarian gagal, artikel tetap dibuat tanpa sumber.
func (s *Service) Generate(ctx context.Context, req GenerateRequest) (*GeneratedPost, error) {
	topic := strings.TrimSpace(req.Topic)
	req.Topic = topic

	useSearch := req.IncludeWebSearch == nil || *req.IncludeWebSearch
	sources := []Source{}
	if useSearch {
		if found, err := s.search.Search(ctx, topic); err == nil {
			sources = found
		}
	}

	raw, err := s.llm.Complete(ctx, systemPrompt, buildPrompt(req, sources))
	if err != nil {
		return nil, err
	}

	g, ok := parseArticle(raw)
	if !ok {
		// Model membalas teks Markdown biasa: pakai apa adanya sebagai isi artikel.
		g = &GeneratedPost{Title: topic, Content: strings.TrimSpace(raw)}
	}
	if strings.TrimSpace(g.Content) == "" {
		return nil, ErrEmptyOutput
	}
	if strings.TrimSpace(g.Title) == "" {
		g.Title = topic
	}
	if strings.TrimSpace(g.Excerpt) == "" {
		g.Excerpt = post.MakeExcerpt(g.Content)
	}

	if req.Category != "" {
		g.Category = post.NormalizeCategory(req.Category)
	} else {
		g.Category = post.NormalizeCategory(g.Category)
	}

	g.Sources = sources
	if len(sources) > 0 {
		var b strings.Builder
		b.WriteString(strings.TrimSpace(g.Content))
		b.WriteString("\n\n## Sumber Rujukan\n")
		for i, src := range sources {
			fmt.Fprintf(&b, "%d. [%s](%s)\n", i+1, src.Title, src.URL)
		}
		g.Content = b.String()
	}
	return g, nil
}
