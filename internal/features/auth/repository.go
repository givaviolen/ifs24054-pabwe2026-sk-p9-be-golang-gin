package auth

import "gorm.io/gorm"

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(u *User) error { return r.db.Create(u).Error }

func (r *Repository) Save(u *User) error { return r.db.Save(u).Error }

func (r *Repository) first(query string, args ...interface{}) (*User, error) {
	var u User
	if err := r.db.Where(query, args...).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindByUsername(name string) (*User, error) {
	return r.first("username = ?", name)
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	return r.first("email = ?", email)
}

// FindByIdentity mencari berdasarkan username ATAU email.
func (r *Repository) FindByIdentity(identity string) (*User, error) {
	return r.first("username = ? OR email = ?", identity, identity)
}

func (r *Repository) FindByID(id uint) (*User, error) {
	var u User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
