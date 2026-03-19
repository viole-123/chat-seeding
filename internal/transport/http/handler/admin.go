package handler

import (
	"fmt"
	"uniscore-seeding-bot/internal/domain/service"
	"uniscore-seeding-bot/internal/usecase/safety"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	killSwitchService service.KillSwitchService
	shadowBanService  *safety.ShadowBanService
}

func NewAdminHandler(killSwitchService service.KillSwitchService, shadowBanService *safety.ShadowBanService) *AdminHandler {
	return &AdminHandler{
		killSwitchService: killSwitchService,
		shadowBanService:  shadowBanService,
	}
}

type KillSwitchRequest struct {
	Scope    string `json:"scope"`        // global, league, match
	ID       string `json:"id,omitempty"` // leagueID hoặc matchID, nếu scope là global thì bỏ qua
	IsKilled bool   `json:"is_killed"`
}

func (h *AdminHandler) SetKillSwitch(c *gin.Context) {
	var req KillSwitchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	//validate scop
	if err := validateScopeAndID(req.Scope, req.ID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.killSwitchService.SetKillSwitch(c.Request.Context(), req.Scope, req.ID, req.IsKilled); err != nil {
		c.JSON(500, gin.H{"error": "failed to set kill switch"})
		return
	}

	status := "enabled"
	if req.IsKilled {
		status = "killed" // FIX: is_killed=true → status="killed", không phải "disabled"
	}
	c.JSON(200, gin.H{"message": "kill switch updated", "scope": req.Scope, "id": req.ID, "status": status})

}

func (h *AdminHandler) GetKillSwitch(c *gin.Context) {
	scope := c.Query("scope")
	id := c.Query("id")

	if scope == "" {
		c.JSON(400, gin.H{"error": "scope is required"})
		return
	}

	if err := validateScopeAndID(scope, id); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	isKilled, err := h.killSwitchService.IsKilled(c.Request.Context(), scope, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get kill switch status"})
		return
	}

	c.JSON(200, gin.H{"scope": scope, "id": id, "is_killed": isKilled})
}

type ShadowBanRequest struct {
	Scope   string `json:"scope"`
	ID      string `json:"id,omitempty"`
	Enabled bool   `json:"enabled"`
}

func (h *AdminHandler) SetShadowBan(c *gin.Context) {
	if h.shadowBanService == nil {
		c.JSON(500, gin.H{"error": "shadow ban service unavailable"})
		return
	}

	var req ShadowBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := validateScopeAndID(req.Scope, req.ID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.shadowBanService.Set(c.Request.Context(), req.Scope, req.ID, req.Enabled); err != nil {
		c.JSON(500, gin.H{"error": "failed to set shadow ban"})
		return
	}

	c.JSON(200, gin.H{"scope": req.Scope, "id": req.ID, "enabled": req.Enabled})
}

func (h *AdminHandler) GetShadowBan(c *gin.Context) {
	if h.shadowBanService == nil {
		c.JSON(500, gin.H{"error": "shadow ban service unavailable"})
		return
	}

	scope := c.Query("scope")
	id := c.Query("id")
	if scope == "" {
		c.JSON(400, gin.H{"error": "scope is required"})
		return
	}
	if err := validateScopeAndID(scope, id); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	enabled, err := h.shadowBanService.IsEnabled(c.Request.Context(), scope, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get shadow ban status"})
		return
	}

	c.JSON(200, gin.H{"scope": scope, "id": id, "enabled": enabled})
}
func validateScopeAndID(scope, id string) error {
	switch scope {
	case "global", "league", "match":
		// ok
	default:
		return fmt.Errorf("scope must be global, league, or match")
	}
	if (scope == "league" || scope == "match") && id == "" {
		return fmt.Errorf("id is required for %s scope", scope)
	}
	return nil
}
