package post

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var ErrNotFound = errors.New("Artikel tidak ditemukan")

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

func makeExcerpt(content string) string {
	r := []rune(strings.Join(strings.Fields(content), " "))
	if len(r) > 160 {
		return string(r[:160]) + "..."
	}
	return string(r)
}

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) uniqueSlug(base string, excludeID uint) (string, error) {
	slug := base
	for i := 2; ; i++ {
		exists, err := s.repo.SlugExists(slug, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

func (s *Service) ListPublic(q, category string) ([]Post, error) { return s.repo.ListPublic(q, category) }

func (s *Service) GetPublic(slug string) (*Post, error) {
	p, err := s.repo.FindPublicBySlug(slug)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *Service) ListMine(uid uint) ([]Post, error) { return s.repo.ListByUser(uid) }

func (s *Service) GetMine(id, uid uint) (*Post, error) {
	p, err := s.repo.FindByIDAndUser(id, uid)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *Service) Create(uid uint, req PostRequest) (*Post, error) {
	slug, err := s.uniqueSlug(Slugify(req.Title), 0)
	if err != nil {
		return nil, err
	}
	p := &Post{
		UserID:         uid,
		Title:          strings.TrimSpace(req.Title),
		Slug:           slug,
		Category:       NormalizeCategory(req.Category),
		Excerpt:        strings.TrimSpace(req.Excerpt),
		Content:        req.Content,
		CoverImagePath: strings.TrimSpace(req.CoverImagePath),
	}
	if p.Excerpt == "" {
		p.Excerpt = makeExcerpt(p.Content)
	}
	if req.IsPublished != nil && *req.IsPublished {
		now := time.Now()
		p.IsPublished, p.PublishedAt = true, &now
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	p.ReadingTime = ReadingTime(p.Content)
	return p, nil
}

func (s *Service) Update(id, uid uint, req PostRequest) (*Post, error) {
	p, err := s.GetMine(id, uid)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	if title != p.Title {
		slug, err := s.uniqueSlug(Slugify(title), p.ID)
		if err != nil {
			return nil, err
		}
		p.Slug = slug
	}
	p.Title = title
	p.Category = NormalizeCategory(req.Category)
	p.Excerpt = strings.TrimSpace(req.Excerpt)
	if p.Excerpt == "" {
		p.Excerpt = makeExcerpt(req.Content)
	}
	p.Content = req.Content
	p.CoverImagePath = strings.TrimSpace(req.CoverImagePath)
	if req.IsPublished != nil {
		applyPublish(p, *req.IsPublished)
	}
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	p.ReadingTime = ReadingTime(p.Content)
	return p, nil
}

func applyPublish(p *Post, publish bool) {
	if publish && !p.IsPublished {
		now := time.Now()
		p.IsPublished, p.PublishedAt = true, &now
	}
	if !publish {
		p.IsPublished, p.PublishedAt = false, nil
	}
}

// TogglePublish membalik status Draf <-> Terbit beserta published_at.
func (s *Service) TogglePublish(id, uid uint) (*Post, error) {
	p, err := s.GetMine(id, uid)
	if err != nil {
		return nil, err
	}
	applyPublish(p, !p.IsPublished)
	if err := s.repo.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(id, uid uint) error {
	if _, err := s.GetMine(id, uid); err != nil {
		return err
	}
	return s.repo.Delete(id, uid)
}

func (s *Service) Stats(uid uint) (Stats, error)         { return s.repo.Stats(uid) }
func (s *Service) AdminUsers() ([]AdminUser, error)      { return s.repo.AdminUsers() }
func (s *Service) AdminPosts() ([]Post, error)           { return s.repo.ListAll() }
func (s *Service) AdminUserPosts(uid uint) ([]Post, error) { return s.repo.ListByUser(uid) }
