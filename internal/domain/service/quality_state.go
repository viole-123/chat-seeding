package service

type QualityStateService interface {
	IsMessageDuplicated(matchID string, msgHash string) (bool, error)
	SaveMessageHash(matchID string, msgHash string, ttlSeconds int) error
}
