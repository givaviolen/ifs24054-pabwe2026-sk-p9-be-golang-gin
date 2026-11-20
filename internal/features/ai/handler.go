package ai

import (
	"net/http"
	"strings"

	"app/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Field topic wajib diisi dengan format JSON yang valid")
		return
	}
	if strings.TrimSpace(req.Topic) == "" {
		response.Fail(c, http.StatusBadRequest, "Field topic wajib diisi")
		return
	}
	g, err := h.svc.Generate(c.Request.Context(), req)
	if err != nil {
		response.Error(c, "Gagal memproses draf OpenAI: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Draf artikel AI berhasil dibuat", g)
}
