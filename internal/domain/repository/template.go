package repository

import "uniscore-seeding-bot/internal/domain/model"

type TemplateRepository interface {
	GetAllTemplates() ([]model.Template, error)
	FindMatchingTemplates(eventType, lang, personaID string) ([]model.Template, error)
	GetTemplateByID(id string) (*model.Template, error)
}

// Compile-time check
// var _ TemplateRepository = (*nil)(nil)
