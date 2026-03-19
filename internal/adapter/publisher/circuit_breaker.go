package publisher

import (
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
	"uniscore-seeding-bot/internal/pkg/circuitbreaker"
)

// Circuit breaker wrapper (stub).
type CircuitBreakerPublisher struct {
	basePublisher service.PublisherService
	cb            *circuitbreaker.Breaker
}

func NewCircuitBreakerPublisher(basePublisher service.PublisherService, cb *circuitbreaker.Breaker) *CircuitBreakerPublisher {
	return &CircuitBreakerPublisher{
		basePublisher: basePublisher,
		cb:            cb,
	}
}

func (c *CircuitBreakerPublisher) Publish(msg model.ChatMessage) error {

	err := c.cb.Execute(func() error {
		return c.basePublisher.Publish(msg)
	})
	return err
}
