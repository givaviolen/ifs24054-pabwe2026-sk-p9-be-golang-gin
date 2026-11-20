package post

import (
	"strings"

	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) ListPublic(q, category string) ([]Post, error) {
	posts := []Post{}
	tx := r.db.Preload("User").Where("is_published = ?", true)
	if q = strings.ToLower(strings.TrimSpace(q)); q != "" {
		like := "%" + q + "%"
		tx = tx.Where("(LOWER(title) LIKE ? OR LOWER(excerpt) LIKE ? OR LOWER(content) LIKE ?)", like, like, like)
	}
	if category = strings.TrimSpace(category); category != "" && !strings.EqualFold(category, "Semua") {
		tx = tx.Where("category = ?", category)
	}
	err := tx.Order("published_at DESC, id DESC").Find(&posts).Error
	return posts, err
}

func (r *Repository) FindPublicBySlug(slug string) (*Post, error) {
	var p Post
	err := r.db.Preload("User").Where("slug = ? AND is_published = ?", slug, true).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ListByUser(uid uint) ([]Post, error) {
	posts := []Post{}
	err := r.db.Preload("User").Where("user_id = ?", uid).Order("id DESC").Find(&posts).Error
	return posts, err
}

func (r *Repository) FindByID(id uint) (*Post, error) {
	var p Post
	if err := r.db.Preload("User").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Create(p *Post) error { return r.db.Omit("User").Create(p).Error }

func (r *Repository) Save(p *Post) error { return r.db.Omit("User").Save(p).Error }

func (r *Repository) Delete(id uint) error { return r.db.Delete(&Post{}, id).Error }

func (r *Repository) SlugExists(slug string) (bool, error) {
	var n int64
	err := r.db.Model(&Post{}).Where("slug = ?", slug).Count(&n).Error
	return n > 0, err
}

func (r *Repository) Stats(uid uint) (total int64, published int64, err error) {
	if err = r.db.Model(&Post{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		return
	}
	err = r.db.Model(&Post{}).Where("user_id = ? AND is_published = ?", uid, true).Count(&published).Error
	return
}

func (r *Repository) ListAll() ([]Post, error) {
	posts := []Post{}
	err := r.db.Preload("User").Order("id DESC").Find(&posts).Error
	return posts, err
}

func (r *Repository) AdminUsers() ([]AdminUserResponse, error) {
	out := []AdminUserResponse{}
	err := r.db.Table("users").
		Select("users.id, users.username, users.email, users.role, users.created_at, " +
			"COUNT(posts.id) AS total_articles, " +
			"COALESCE(SUM(CASE WHEN posts.is_published THEN 1 ELSE 0 END), 0) AS total_published").
		Joins("LEFT JOIN posts ON posts.user_id = users.id").
		Group("users.id").
		Order("users.id").
		Scan(&out).Error
	return out, err
}
