package ai

import (
	"errors"
	"net/http"

	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Topik wajib diisi (minimal 3 karakter)")
		return
	}
	g, err := h.svc.Generate(c.Request.Context(), req)
	if errors.Is(err, ErrBadModelOutput) {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	if err != nil {
		response.Fail(c, http.StatusBadGateway, "Gagal menghasilkan artikel dari layanan AI")
		return
	}
	response.Success(c, http.StatusOK, "Draf artikel berhasil dibuat", g)
}
