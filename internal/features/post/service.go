package post

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("Artikel tidak ditemukan")
	ErrForbidden    = errors.New("Anda tidak memiliki izin untuk mengakses artikel ini")
	ErrInvalidInput = errors.New("Judul dan konten artikel wajib diisi")
)

// Categories adalah daftar kategori artikel yang valid.
var Categories = []string{"Teknologi", "Pendidikan", "Gaya Hidup", "Bisnis", "Opini", "Umum"}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify mengubah judul menjadi slug URL.
func Slugify(s string) string {
	s = nonAlnum.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "artikel"
	}
	return s
}

// ReadingTime menghitung estimasi menit baca: jumlah kata / 200 wpm (minimal 1).
func ReadingTime(content string) int {
	m := (len(strings.Fields(content)) + 199) / 200
	if m < 1 {
		return 1
	}
	return m
}

// NormalizeCategory mengembalikan kategori valid, selain itu "Umum".
func NormalizeCategory(c string) string {
	for _, v := range Categories {
		if strings.EqualFold(v, strings.TrimSpace(c)) {
			return v
		}
	}
	return "Umum"
}

// MakeExcerpt membuat ringkasan otomatis (maks 160 karakter) dari isi artikel.
func MakeExcerpt(content string) string {
	r := []rune(strings.Join(strings.Fields(content), " "))
	if len(r) > 160 {
		return string(r[:160]) + "..."
	}
	return string(r)
}

func trimmed(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func toResponse(p *Post) PostResponse {
	return PostResponse{
		ID:             p.ID,
		UserID:         p.UserID,
		Title:          p.Title,
		Slug:           p.Slug,
		Category:       p.Category,
		Excerpt:        p.Excerpt,
		Content:        p.Content,
		CoverImagePath: p.CoverImagePath,
		IsPublished:    p.IsPublished,
		ReadingTime:    ReadingTime(p.Content),
		AuthorName:     p.User.Username,
		PublishedAt:    p.PublishedAt,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func toPublic(p *Post) PublicPostResponse {
	author := "Anonim"
	if p.User.Username != "" {
		author = p.User.Username
	}
	return PublicPostResponse{
		ID:             p.ID,
		Title:          p.Title,
		Slug:           p.Slug,
		Category:       p.Category,
		Excerpt:        p.Excerpt,
		Content:        p.Content,
		CoverImagePath: p.CoverImagePath,
		IsPublished:    p.IsPublished,
		ReadingTime:    ReadingTime(p.Content),
		AuthorName:     author,
		PublishedAt:    p.PublishedAt,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func toResponses(posts []Post) []PostResponse {
	out := make([]PostResponse, 0, len(posts))
	for i := range posts {
		out = append(out, toResponse(&posts[i]))
	}
	return out
}

func applyPublish(p *Post, publish bool) {
	p.IsPublished = publish
	if !publish {
		p.PublishedAt = nil
		return
	}
	if p.PublishedAt == nil {
		now := time.Now()
		p.PublishedAt = &now
	}
}

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) GetCategories() []string { return Categories }

func (s *Service) uniqueSlug(base string) (string, error) {
	slug := base
	for i := 2; ; i++ {
		exists, err := s.repo.SlugExists(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// owned memuat artikel dan memastikan milik user (admin boleh semuanya).
func (s *Service) owned(id, uid uint, role string) (*Post, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if p.UserID != uid && !strings.EqualFold(role, "admin") {
		return nil, ErrForbidden
	}
	return p, nil
}

func (s *Service) GetPublicPosts(q, category string) ([]PublicPostResponse, error) {
	posts, err := s.repo.ListPublic(q, category)
	if err != nil {
		return nil, err
	}
	out := make([]PublicPostResponse, 0, len(posts))
	for i := range posts {
		out = append(out, toPublic(&posts[i]))
	}
	return out, nil
}

func (s *Service) GetPublicPost(slug string) (*PublicPostResponse, error) {
	p, err := s.repo.FindPublicBySlug(slug)
	if err != nil {
		return nil, ErrNotFound
	}
	res := toPublic(p)
	return &res, nil
}

func (s *Service) Create(uid uint, req CreatePostRequest) (*PostResponse, error) {
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" {
		return nil, ErrInvalidInput
	}
	slug, err := s.uniqueSlug(Slugify(title))
	if err != nil {
		return nil, err
	}

	p := &Post{
		UserID:         uid,
		Title:          title,
		Slug:           slug,
		Category:       NormalizeCategory(req.Category),
		Excerpt:        strings.TrimSpace(req.Excerpt),
		Content:        content,
		CoverImagePath: strings.TrimSpace(req.CoverImagePath),
	}
	if p.Excerpt == "" {
		p.Excerpt = MakeExcerpt(content)
	}
	if req.IsPublished != nil && *req.IsPublished {
		applyPublish(p, true)
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	res := toResponse(p)
	return &res, nil
}

func (s *Service) ListMine(uid uint) ([]PostResponse, error) {
	posts, err := s.repo.ListByUser(uid)
	if err != nil {
		return nil, err
	}
	return toResponses(posts), nil
}

func (s *Service) GetMine(id, uid uint, role string) (*PostResponse, error) {
	p, err := s.owned(id, uid, role)
	if err != nil {
		return nil, err
	}
	res := toResponse(p)
	return &res, nil
}

func (s *Service) Stats(uid uint) (*StatsResponse, error) {
	total, published, err := s.repo.Stats(uid)
	if err != nil {
		return nil, err
	}
	return &StatsResponse{TotalArticles: total, TotalSelesai: published, BelumSelesai: total - published}, nil
}

func (s *Service) Update(id, uid uint, role string, req UpdatePostRequest) (*PostResponse, error) {
	p, err := s.owned(id, uid, role)
	if err != nil {
		return nil, err
	}
	if v := trimmed(req.Title); v != "" {
		p.Title = v
	}
	if v := trimmed(req.Category); v != "" {
		p.Category = NormalizeCategory(v)
	}
	if v := trimmed(req.Excerpt); v != "" {
		p.Excerpt = v
	}
	if v := trimmed(req.Content); v != "" {
		p.Content = v
	}
	if v := trimmed(req.CoverImagePath); v != "" {
		p.CoverImagePath = v
	}
	if req.IsPublished != nil {
		applyPublish(p, *req.IsPublished)
	}
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	res := toResponse(p)
	return &res, nil
}

// TogglePublish membalik status Draf <-> Terbit beserta published_at.
func (s *Service) TogglePublish(id, uid uint, role string) (*PostResponse, error) {
	p, err := s.owned(id, uid, role)
	if err != nil {
		return nil, err
	}
	p.PublishedAt = nil // setiap kali terbit, waktu publikasi diperbarui
	applyPublish(p, !p.IsPublished)
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	res := toResponse(p)
	return &res, nil
}

func (s *Service) Delete(id, uid uint, role string) error {
	if _, err := s.owned(id, uid, role); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *Service) AdminUsers() ([]AdminUserResponse, error) { return s.repo.AdminUsers() }

func (s *Service) AdminPosts() ([]PostResponse, error) {
	posts, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	return toResponses(posts), nil
}

func (s *Service) AdminUserPosts(uid uint) ([]PostResponse, error) {
	posts, err := s.repo.ListByUser(uid)
	if err != nil {
		return nil, err
	}
	return toResponses(posts), nil
}
