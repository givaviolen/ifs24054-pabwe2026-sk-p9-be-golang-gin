package post

import (
	"errors"
	"net/http"
	"strconv"

	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func paramID(c *gin.Context, label string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || n == 0 {
		response.Fail(c, http.StatusBadRequest, "ID "+label+" tidak valid")
		return 0, false
	}
	return uint(n), true
}

func (h *Handler) fail(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrInvalidInput):
		response.Fail(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, fallback+": "+err.Error())
	}
}

// ---- Publik ----

func (h *Handler) Categories(c *gin.Context) {
	response.Success(c, http.StatusOK, "Daftar kategori berhasil diambil", h.svc.GetCategories())
}

func (h *Handler) PublicList(c *gin.Context) {
	posts, err := h.svc.GetPublicPosts(c.Query("q"), c.Query("category"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil daftar artikel publik")
		return
	}
	response.Success(c, http.StatusOK, "Daftar artikel publik berhasil diambil", posts)
}

func (h *Handler) PublicDetail(c *gin.Context) {
	p, err := h.svc.GetPublicPost(c.Param("slug"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Detail artikel berhasil diambil", p)
}

// ---- Penulis ----

func (h *Handler) List(c *gin.Context) {
	posts, err := h.svc.ListMine(c.GetUint("user_id"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Daftar artikel penulis berhasil diambil", posts)
}

func (h *Handler) Stats(c *gin.Context) {
	s, err := h.svc.Stats(c.GetUint("user_id"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil statistik")
		return
	}
	response.Success(c, http.StatusOK, "Statistik artikel berhasil diambil", s)
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := paramID(c, "artikel")
	if !ok {
		return
	}
	p, err := h.svc.GetMine(id, c.GetUint("user_id"), c.GetString("role"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Detail artikel berhasil diambil", p)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Payload request tidak valid")
		return
	}
	p, err := h.svc.Create(c.GetUint("user_id"), req)
	if err != nil {
		h.fail(c, err, "Gagal membuat artikel")
		return
	}
	response.Success(c, http.StatusCreated, "Artikel berhasil dibuat", p)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := paramID(c, "artikel")
	if !ok {
		return
	}
	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Payload request tidak valid")
		return
	}
	p, err := h.svc.Update(id, c.GetUint("user_id"), c.GetString("role"), req)
	if err != nil {
		h.fail(c, err, "Gagal memperbarui artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diperbarui", p)
}

func (h *Handler) TogglePublish(c *gin.Context) {
	id, ok := paramID(c, "artikel")
	if !ok {
		return
	}
	p, err := h.svc.TogglePublish(id, c.GetUint("user_id"), c.GetString("role"))
	if err != nil {
		h.fail(c, err, "Gagal mengubah status publikasi")
		return
	}
	response.Success(c, http.StatusOK, "Status publikasi artikel berhasil diubah", p)
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := paramID(c, "artikel")
	if !ok {
		return
	}
	if err := h.svc.Delete(id, c.GetUint("user_id"), c.GetString("role")); err != nil {
		h.fail(c, err, "Gagal menghapus artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil dihapus", nil)
}

// ---- Admin ----

func (h *Handler) AdminUsers(c *gin.Context) {
	users, err := h.svc.AdminUsers()
	if err != nil {
		h.fail(c, err, "Gagal mengambil daftar pengguna")
		return
	}
	response.Success(c, http.StatusOK, "Daftar pengguna berhasil diambil", users)
}

func (h *Handler) AdminPosts(c *gin.Context) {
	posts, err := h.svc.AdminPosts()
	if err != nil {
		h.fail(c, err, "Gagal mengambil seluruh artikel")
		return
	}
	response.Success(c, http.StatusOK, "Seluruh artikel berhasil diambil", posts)
}

func (h *Handler) AdminUserPosts(c *gin.Context) {
	id, ok := paramID(c, "pengguna")
	if !ok {
		return
	}
	posts, err := h.svc.AdminUserPosts(id)
	if err != nil {
		h.fail(c, err, "Gagal mengambil artikel pengguna")
		return
	}
	response.Success(c, http.StatusOK, "Artikel pengguna berhasil diambil", posts)
}
