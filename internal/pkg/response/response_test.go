package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := newTestContext()

	data := map[string]interface{}{"key": "value"}
	Success(c, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, data, resp.Data)
}

func TestFail(t *testing.T) {
	c, w := newTestContext()

	Fail(c, "error message")

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Code)
	assert.Equal(t, "error message", resp.Message)
}

func TestFailWithCode(t *testing.T) {
	c, w := newTestContext()

	FailWithCode(c, http.StatusBadRequest, "bad request")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Code)
	assert.Equal(t, "bad request", resp.Message)
}

func TestUnauthorized(t *testing.T) {
	c, w := newTestContext()

	Unauthorized(c, "unauthorized")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "unauthorized", resp.Message)
}

func TestForbidden(t *testing.T) {
	c, w := newTestContext()

	Forbidden(c, "forbidden")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, "forbidden", resp.Message)
}

func TestNotFound(t *testing.T) {
	c, w := newTestContext()

	NotFound(c, "not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.Code)
	assert.Equal(t, "not found", resp.Message)
}

func TestServerError(t *testing.T) {
	c, w := newTestContext()

	ServerError(c, "server error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.Code)
	assert.Equal(t, "server error", resp.Message)
}

func TestPaginated(t *testing.T) {
	c, w := newTestContext()

	items := []string{"item1", "item2"}
	Paginated(c, items, 100, 1, 10)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(100), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["page_size"])
}

func TestErrorResp(t *testing.T) {
	resp := ErrorResp("error message")
	assert.Equal(t, 1, resp.Code)
	assert.Equal(t, "error message", resp.Message)
}