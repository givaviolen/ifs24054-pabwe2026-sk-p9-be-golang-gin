package auth

import (
	"errors"
	"net/http"
	"strings"

	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) fail(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrDuplicateUser):
		response.Fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		response.Fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrUserNotFound):
		response.Fail(c, http.StatusNotFound, err.Error())
	default:
		response.Error(c, fallback)
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Payload request tidak valid")
		return
	}
	res, err := h.svc.Register(req)
	if err != nil {
		h.fail(c, err, "Gagal memproses registrasi")
		return
	}
	response.Success(c, http.StatusCreated, "Registrasi berhasil", res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Payload request tidak valid")
		return
	}
	if strings.TrimSpace(req.Username) == "" || req.Password == "" {
		response.Fail(c, http.StatusBadRequest, "Username dan password wajib diisi")
		return
	}
	res, err := h.svc.Login(req)
	if err != nil {
		h.fail(c, err, "Gagal memproses login")
		return
	}
	response.Success(c, http.StatusOK, "Login berhasil", res)
}

func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, "Logout berhasil", nil)
}

func (h *Handler) Profile(c *gin.Context) {
	u, err := h.svc.Profile(c.GetUint("user_id"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil profil")
		return
	}
	response.Success(c, http.StatusOK, "Profil berhasil diambil", ToUserResponse(u))
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Payload request tidak valid")
		return
	}
	u, err := h.svc.UpdateProfile(c.GetUint("user_id"), req)
	if err != nil {
		h.fail(c, err, "Gagal memperbarui profil")
		return
	}
	response.Success(c, http.StatusOK, "Profil berhasil diperbarui", ToUserResponse(u))
}
