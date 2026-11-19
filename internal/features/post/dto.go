package post

import (
	"time"

	"app/internal/features/auth"

	"gorm.io/gorm"
)

// Post adalah model tabel posts.
type Post struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	User           auth.User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"author"`
	Title          string     `gorm:"size:255;not null" json:"title"`
	Slug           string     `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Category       string     `gorm:"size:100" json:"category"`
	Excerpt        string     `gorm:"type:text" json:"excerpt"`
	Content        string     `gorm:"type:text" json:"content"`
	CoverImagePath string     `gorm:"type:text" json:"cover_image_path"`
	IsPublished    bool       `json:"is_published"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ReadingTime    int        `gorm:"-" json:"reading_time"`
}

func (Post) TableName() string { return "posts" }

// AfterFind mengisi estimasi waktu baca setiap artikel yang dibaca dari DB.
func (p *Post) AfterFind(tx *gorm.DB) error {
	p.ReadingTime = ReadingTime(p.Content)
	return nil
}

type PostRequest struct {
	Title          string `json:"title" binding:"required,max=255"`
	Category       string `json:"category"`
	Excerpt        string `json:"excerpt"`
	Content        string `json:"content" binding:"required"`
	CoverImagePath string `json:"cover_image_path"`
	IsPublished    *bool  `json:"is_published"`
}

type Stats struct {
	Total     int64 `json:"total"`
	Published int64 `json:"published"`
	Draft     int64 `json:"draft"`
}

type AdminUser struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	TotalPosts     int64  `json:"total_posts"`
	TotalPublished int64  `json:"total_published"`
}
