package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("Username atau password salah")
	ErrUserExists         = errors.New("Username atau email sudah digunakan")
	ErrUserNotFound       = errors.New("Pengguna tidak ditemukan")
)

type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func GenerateToken(secret string, u *User) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseToken(secret, token string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode tanda tangan tidak valid")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}

type Service struct {
	repo   *Repository
	secret string
}

func NewService(repo *Repository, secret string) *Service {
	return &Service{repo: repo, secret: secret}
}

func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	exists, err := s.repo.Exists(req.Username, req.Email, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	u := &User{Username: req.Username, Email: req.Email, Password: hash, Role: "member"}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	tok, err := GenerateToken(s.secret, u)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: tok, User: u}, nil
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	u, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	tok, err := GenerateToken(s.secret, u)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: tok, User: u}, nil
}

func (s *Service) Profile(id uint) (*User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Service) UpdateProfile(id uint, req UpdateProfileRequest) (*User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	exists, err := s.repo.Exists(req.Username, req.Email, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	u.Username = req.Username
	u.Email = req.Email
	if req.Password != "" {
		hash, err := HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		u.Password = hash
	}
	if err := s.repo.Save(u); err != nil {
		return nil, err
	}
	return u, nil
}
