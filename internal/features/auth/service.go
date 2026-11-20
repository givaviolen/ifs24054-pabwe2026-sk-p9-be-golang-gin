package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput       = errors.New("Input tidak valid")
	ErrDuplicateUser      = errors.New("Username atau email sudah terdaftar")
	ErrInvalidCredentials = errors.New("Username atau password salah")
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

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	parts := strings.Split(email, "@")
	return len(parts) == 2 && strings.Contains(parts[1], ".")
}

type Service struct {
	repo       *Repository
	secret     string
	allowAdmin bool
}

func NewService(repo *Repository, secret string, allowAdmin bool) *Service {
	return &Service{repo: repo, secret: secret, allowAdmin: allowAdmin}
}

func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if len(username) < 3 {
		return nil, fmt.Errorf("%w: username minimal 3 karakter", ErrInvalidInput)
	}
	if !validEmail(email) {
		return nil, fmt.Errorf("%w: format email salah", ErrInvalidInput)
	}
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("%w: password minimal 6 karakter", ErrInvalidInput)
	}

	// Role hanya "admin" / "member"; selain itu dipaksa "member".
	role := "member"
	if s.allowAdmin && strings.EqualFold(strings.TrimSpace(req.Role), "admin") {
		role = "admin"
	}

	if u, _ := s.repo.FindByUsername(username); u != nil {
		return nil, ErrDuplicateUser
	}
	if u, _ := s.repo.FindByEmail(email); u != nil {
		return nil, ErrDuplicateUser
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user := &User{Username: username, Email: email, Password: hash, Role: role}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return s.authResponse(user)
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	identity := strings.TrimSpace(req.Username)
	if identity == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}
	user, err := s.repo.FindByIdentity(identity)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return s.authResponse(user)
}

func (s *Service) authResponse(u *User) (*AuthResponse, error) {
	tok, err := GenerateToken(s.secret, u)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: tok, User: ToUserResponse(u)}, nil
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

	if email := strings.ToLower(strings.TrimSpace(req.Email)); email != "" {
		if !validEmail(email) {
			return nil, fmt.Errorf("%w: format email salah", ErrInvalidInput)
		}
		if other, _ := s.repo.FindByEmail(email); other != nil && other.ID != id {
			return nil, ErrDuplicateUser
		}
		u.Email = email
	}

	if req.Password != "" {
		if len(req.Password) < 6 {
			return nil, fmt.Errorf("%w: password minimal 6 karakter", ErrInvalidInput)
		}
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
