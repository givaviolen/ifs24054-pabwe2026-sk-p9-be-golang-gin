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

func paramID(c *gin.Context) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || n == 0 {
		response.Fail(c, http.StatusBadRequest, "ID tidak valid")
		return 0, false
	}
	return uint(n), true
}

func (h *Handler) fail(c *gin.Context, err error, msg string) {
	if errors.Is(err, ErrNotFound) {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.Error(c, msg)
}

func hideEmail(posts []Post) {
	for i := range posts {
		posts[i].User.Email = ""
	}
}

// ---- Publik ----

func (h *Handler) Categories(c *gin.Context) {
	response.Success(c, http.StatusOK, "Kategori berhasil diambil", Categories)
}

func (h *Handler) PublicList(c *gin.Context) {
	posts, err := h.svc.ListPublic(c.Query("q"), c.Query("category"))
	if err != nil {
		response.Error(c, "Gagal mengambil artikel")
		return
	}
	hideEmail(posts)
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", posts)
}

func (h *Handler) PublicDetail(c *gin.Context) {
	p, err := h.svc.GetPublic(c.Param("slug"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil artikel")
		return
	}
	p.User.Email = ""
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", p)
}

// ---- Penulis ----

func (h *Handler) List(c *gin.Context) {
	posts, err := h.svc.ListMine(c.GetUint("user_id"))
	if err != nil {
		response.Error(c, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", posts)
}

func (h *Handler) Stats(c *gin.Context) {
	s, err := h.svc.Stats(c.GetUint("user_id"))
	if err != nil {
		response.Error(c, "Gagal mengambil statistik")
		return
	}
	response.Success(c, http.StatusOK, "Statistik berhasil diambil", s)
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	p, err := h.svc.GetMine(id, c.GetUint("user_id"))
	if err != nil {
		h.fail(c, err, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", p)
}

func (h *Handler) Create(c *gin.Context) {
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Judul dan konten wajib diisi")
		return
	}
	p, err := h.svc.Create(c.GetUint("user_id"), req)
	if err != nil {
		response.Error(c, "Gagal membuat artikel")
		return
	}
	response.Success(c, http.StatusCreated, "Artikel berhasil dibuat", p)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Judul dan konten wajib diisi")
		return
	}
	p, err := h.svc.Update(id, c.GetUint("user_id"), req)
	if err != nil {
		h.fail(c, err, "Gagal memperbarui artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diperbarui", p)
}

func (h *Handler) TogglePublish(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	p, err := h.svc.TogglePublish(id, c.GetUint("user_id"))
	if err != nil {
		h.fail(c, err, "Gagal mengubah status publikasi")
		return
	}
	response.Success(c, http.StatusOK, "Status publikasi berhasil diubah", p)
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(id, c.GetUint("user_id")); err != nil {
		h.fail(c, err, "Gagal menghapus artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil dihapus", nil)
}

// ---- Admin ----

func (h *Handler) AdminUsers(c *gin.Context) {
	users, err := h.svc.AdminUsers()
	if err != nil {
		response.Error(c, "Gagal mengambil pengguna")
		return
	}
	response.Success(c, http.StatusOK, "Pengguna berhasil diambil", users)
}

func (h *Handler) AdminPosts(c *gin.Context) {
	posts, err := h.svc.AdminPosts()
	if err != nil {
		response.Error(c, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", posts)
}

func (h *Handler) AdminUserPosts(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	posts, err := h.svc.AdminUserPosts(id)
	if err != nil {
		response.Error(c, "Gagal mengambil artikel")
		return
	}
	response.Success(c, http.StatusOK, "Artikel berhasil diambil", posts)
}
