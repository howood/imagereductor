package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howood/imagereductor/application/usecase"
	"github.com/howood/imagereductor/di/uccluster"
	"github.com/howood/imagereductor/interfaces/handler"
	"github.com/labstack/echo/v5"
)

func Test_TokenHandler_Request(t *testing.T) {
	t.Parallel()

	cluster := &uccluster.UsecaseCluster{
		TokenUC: usecase.NewTokenUsecase(),
	}
	th := handler.NewTokenHandler(handler.BaseHandler{UcCluster: cluster})

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := th.Request(c); err != nil {
		t.Fatalf("Request returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	tok, ok := body["token"].(string)
	if !ok || tok == "" {
		t.Fatalf("expected non-empty token in body, got %#v", body)
	}
}
