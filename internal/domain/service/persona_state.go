package service

import "context"

type PersonaStateService interface {
	IsOnCoolDown(ctx context.Context, personaId string) (bool, error)
	SetCoolDown(ctx context.Context, personaId string, durationSeconds int) error

	IsAntiRepeat(ctx context.Context, matchId, personaId, msgHash string) (bool, error)
	SetLastMessageHash(ctx context.Context, matchId, personaId, msgHash string, ttlDurationSeconds int) error

	SaveLastPersona(ctx context.Context, matchId, personaId string) error
	GetLastPersona(ctx context.Context, matchId string) (string, error)
}
