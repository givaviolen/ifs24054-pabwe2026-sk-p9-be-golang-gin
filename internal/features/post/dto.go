package post

import (
	"time"

	"app/internal/features/auth"
)

// Post adalah model tabel posts.
type Post struct {
	ID             uint       `gorm:"primaryKey"`
	UserID         uint       `gorm:"not null;index"`
	User           auth.User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Title          string     `gorm:"size:255;not null"`
	Slug           string     `gorm:"size:255;uniqueIndex;not null"`
	Category       string     `gorm:"size:100"`
	Excerpt        string     `gorm:"type:text"`
	Content        string     `gorm:"type:text"`
	CoverImagePath string     `gorm:"type:text"`
	IsPublished    bool
	PublishedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Post) TableName() string { return "posts" }

type CreatePostRequest struct {
	Title          string `json:"title"`
	Category       string `json:"category"`
	Excerpt        string `json:"excerpt"`
	Content        string `json:"content"`
	CoverImagePath string `json:"cover_image_path"`
	IsPublished    *bool  `json:"is_published"`
}

// UpdatePostRequest: semua field opsional; field kosong tidak menimpa data lama.
type UpdatePostRequest struct {
	Title          *string `json:"title"`
	Category       *string `json:"category"`
	Excerpt        *string `json:"excerpt"`
	Content        *string `json:"content"`
	CoverImagePath *string `json:"cover_image_path"`
	IsPublished    *bool   `json:"is_published"`
}

// PublicPostResponse tidak memuat user_id maupun data sensitif.
type PublicPostResponse struct {
	ID             uint       `json:"id"`
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Category       string     `json:"category"`
	Excerpt        string     `json:"excerpt"`
	Content        string     `json:"content"`
	CoverImagePath string     `json:"cover_image_path"`
	IsPublished    bool       `json:"is_published"`
	ReadingTime    int        `json:"reading_time"`
	AuthorName     string     `json:"author_name"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type PostResponse struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"user_id"`
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Category       string     `json:"category"`
	Excerpt        string     `json:"excerpt"`
	Content        string     `json:"content"`
	CoverImagePath string     `json:"cover_image_path"`
	IsPublished    bool       `json:"is_published"`
	ReadingTime    int        `json:"reading_time"`
	AuthorName     string     `json:"author_name,omitempty"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type StatsResponse struct {
	TotalArticles int64 `json:"total_articles"`
	TotalSelesai  int64 `json:"total_selesai"`
	BelumSelesai  int64 `json:"belum_selesai"`
}

type AdminUserResponse struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	TotalArticles  int64     `json:"total_articles"`
	TotalPublished int64     `json:"total_published"`
	CreatedAt      time.Time `json:"created_at"`
}
