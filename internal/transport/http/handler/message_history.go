package handler

import (
	"net/http"
	"strconv"
	"uniscore-seeding-bot/internal/domain/repository"

	"github.com/gin-gonic/gin"
)

type MessageHistoryHandler struct {
	messageRepo repository.MessageRepository
}

func NewMessageHistoryHandler(messageRepo repository.MessageRepository) *MessageHistoryHandler {
	return &MessageHistoryHandler{messageRepo: messageRepo}
}

func (h *MessageHistoryHandler) List(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	if c.Request.Method == http.MethodOptions {
		c.Status(http.StatusNoContent)
		return
	}

	matchID := c.Query("match_id")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match_id is required"})
		return
	}

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			if parsed > 500 {
				parsed = 500
			}
			limit = parsed
		}
	}

	messages, err := h.messageRepo.GetMessageHistory(matchID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch message history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match_id": matchID,
		"count":    len(messages),
		"messages": messages,
	})
}
