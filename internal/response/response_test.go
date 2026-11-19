package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func run(h gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	h(c)
	return w
}

func TestResponses(t *testing.T) {
	var b Body

	w := run(func(c *gin.Context) { Success(c, 200, "ok", gin.H{"a": 1}) })
	assert.Equal(t, 200, w.Code)
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	assert.Equal(t, "success", b.Status)
	assert.NotNil(t, b.Data)

	w = run(func(c *gin.Context) { Fail(c, 404, "nope") })
	assert.Equal(t, 404, w.Code)
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	assert.Equal(t, "fail", b.Status)

	w = run(func(c *gin.Context) { Error(c, "boom") })
	assert.Equal(t, 500, w.Code)
}
