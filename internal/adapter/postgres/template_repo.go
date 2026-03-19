package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"uniscore-seeding-bot/internal/domain/model"
)

type TemplateRepoImpl struct {
	db *sql.DB
}

func NewTemplateRepo(db *sql.DB) *TemplateRepoImpl { return &TemplateRepoImpl{db: db} }

func (r *TemplateRepoImpl) GetAllTemplates() ([]model.Template, error) {
	query := `
	SELECT id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at
	FROM templates
	WHERE enabled = true
	ORDER BY priority desc`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query templates failed: %w", err)
	}

	defer rows.Close()

	var templates []model.Template
	for rows.Next() {
		var t model.Template
		var conditionJSON []byte
		var personaID sql.NullString
		var phase sql.NullString
		err := rows.Scan(&t.ID, &phase, &t.EventType, &t.Lang, &personaID, &t.Text, &t.Priority, &t.Enabled, &conditionJSON, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan template failed: %w", err)
		}
		if phase.Valid {
			t.Phase = model.MatchPhase(phase.String)
		}
		if personaID.Valid {
			t.PersonaID = personaID.String
		}
		if len(conditionJSON) > 0 {
			if err := json.Unmarshal(conditionJSON, &t.Conditions); err != nil {
				return nil, fmt.Errorf("unmarshal conditions failed: %w", err)
			}
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *TemplateRepoImpl) FindMatchingTemplates(eventType, lang, personaID string) ([]model.Template, error) {
	query := `
        SELECT id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at
        FROM templates
        WHERE enabled = true
          AND event_type = $1
          AND lang = $2
          AND (persona_id = $3 OR persona_id IS NULL)
        ORDER BY priority DESC
    `
	rows, err := r.db.QueryContext(context.Background(), query, eventType, lang, personaID)
	if err != nil {
		return nil, fmt.Errorf("query templates failed: %w", err)
	}
	defer rows.Close()
	var templates []model.Template
	for rows.Next() {
		var t model.Template
		var conditionJSON []byte
		var personaID sql.NullString
		err := rows.Scan(&t.ID, &t.EventType, &t.Lang, &personaID, &t.Text, &t.Priority, &t.Enabled, &conditionJSON, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan template failed: %w", err)
		}
		if personaID.Valid {
			t.PersonaID = personaID.String
		}
		if len(conditionJSON) > 0 {
			if err := json.Unmarshal(conditionJSON, &t.Conditions); err != nil {
				return nil, fmt.Errorf("unmarshal conditions failed: %w", err)
			}

		}

		templates = append(templates, t)
	}
	return templates, nil
}

func (r *TemplateRepoImpl) GetTemplateByID(id string) (*model.Template, error) {
	query := `
        SELECT id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at
        FROM templates
        WHERE id = $1 and enabled = true
    `
	var t model.Template
	var conditionJSON []byte
	var personaID sql.NullString
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(&t.ID, &t.EventType, &t.Lang, &personaID, &t.Text, &t.Priority, &t.Enabled, &conditionJSON, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("query template by ID failed: %w", err)
	}
	if personaID.Valid {
		t.PersonaID = personaID.String
	}
	if len(conditionJSON) > 0 {
		if err := json.Unmarshal(conditionJSON, &t.Conditions); err != nil {
			return nil, fmt.Errorf("unmarshal conditions failed: %w", err)
		}
	}
	return &t, nil
}
