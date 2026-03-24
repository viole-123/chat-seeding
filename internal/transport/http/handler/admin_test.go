package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeKillSwitchService struct{}

func (f *fakeKillSwitchService) SetKillSwitch(ctx context.Context, scope, id string, isKilled bool) error {
	return nil
}

func (f *fakeKillSwitchService) IsKilled(ctx context.Context, matchID, leagueID string) (bool, error) {
	if matchID == "m123" || leagueID == "l123" {
		return true, nil
	}
	return false, nil
}

func TestGetKillSwitch_MatchScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewAdminHandler(&fakeKillSwitchService{}, nil)
	router.GET("/admin/kill-switch/status", h.GetKillSwitch)

	req := httptest.NewRequest(http.MethodGet, "/admin/kill-switch/status?scope=match&id=m123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if body := w.Body.String(); body == "" || body[0] != '{' {
		t.Fatalf("expected json response body")
	}
}

func TestGetKillSwitch_InvalidScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewAdminHandler(&fakeKillSwitchService{}, nil)
	router.GET("/admin/kill-switch/status", h.GetKillSwitch)

	req := httptest.NewRequest(http.MethodGet, "/admin/kill-switch/status?scope=bad", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
