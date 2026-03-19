package service

import "context"

type DedupService interface {
	IsDuplicateEvent(ctx context.Context, matchID, minute, eventID string) (bool, error)
	IsDuplicateMessage(ctx context.Context, matchID, msgHash string) (bool, error)
}
