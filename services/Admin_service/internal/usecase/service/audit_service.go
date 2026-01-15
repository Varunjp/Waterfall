package service

import (
	"admin_service/internal/domain/entities"
	repo "admin_service/internal/repository/interfaces"
)

type AuditService struct {
	repo repo.AuditRepository
}

func NewAuditService(r repo.AuditRepository) *AuditService {
	return &AuditService{r}
}

func (s *AuditService) Log(actorType, actorID, action, resource string, meta map[string]interface{}) {
	_ = s.repo.Log(&entities.AuditLog{
		ActorType: actorType,
		ActorID:   actorID,
		Action:    action,
		Resource:  resource,
		Metadata:  meta,
	})
}
