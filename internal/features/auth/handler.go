package auth

import (
	"errors"
	"net/http"

	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data registrasi tidak valid: "+err.Error())
		return
	}
	res, err := h.svc.Register(req)
	if errors.Is(err, ErrUserExists) {
		response.Fail(c, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		response.Error(c, "Gagal mendaftarkan akun")
		return
	}
	response.Success(c, http.StatusCreated, "Registrasi berhasil", res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Username dan password wajib diisi")
		return
	}
	res, err := h.svc.Login(req)
	if errors.Is(err, ErrInvalidCredentials) {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		response.Error(c, "Gagal login")
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
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Profil berhasil diambil", u)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data profil tidak valid: "+err.Error())
		return
	}
	u, err := h.svc.UpdateProfile(c.GetUint("user_id"), req)
	if errors.Is(err, ErrUserExists) {
		response.Fail(c, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, ErrUserNotFound) {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		response.Error(c, "Gagal memperbarui profil")
		return
	}
	response.Success(c, http.StatusOK, "Profil berhasil diperbarui", u)
}
