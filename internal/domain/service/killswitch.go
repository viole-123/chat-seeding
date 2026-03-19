package service

import "context"

type KillSwitchService interface {
	SetKillSwitch(ctx context.Context, scope, id string, isKilled bool) error
	IsKilled(ctx context.Context, scope, id string) (bool, error)
}
