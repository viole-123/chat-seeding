package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type EventSenderHandler struct {
	producer sarama.SyncProducer
	topic    string
}

func NewEventSenderHandler(producer sarama.SyncProducer, topic string) *EventSenderHandler {
	return &EventSenderHandler{producer: producer, topic: topic}
}

func (h *EventSenderHandler) SendEvent(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	if !json.Valid(body) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: h.topic,
		Value: sarama.ByteEncoder(body),
	}

	partition, offset, err := h.producer.SendMessage(msg)
	if err != nil {
		log.Printf("❌ [SEND-EVENT] Failed to send to Kafka: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send to kafka: " + err.Error()})
		return
	}

	log.Printf("✅ [SEND-EVENT] Sent %d bytes → partition=%d offset=%d", len(body), partition, offset)
	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"partition": partition,
		"offset":    offset,
		"bytes":     len(body),
	})
}
