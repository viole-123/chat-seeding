package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
)

type MessageRepoImpl struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepoImpl {
	return &MessageRepoImpl{db: db}
}

func (r *MessageRepoImpl) SaveMessage(msg model.ChatMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
        INSERT INTO published_messages (id, match_id, room_id, content, persona_id, event_type, is_bot, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.MatchID,
		msg.RoomID,
		msg.Content,
		msg.Persona,
		msg.EventType,
		msg.IsBot,
		msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("Them published message failed: %w", err)

	}
	return nil
}

func (r *MessageRepoImpl) GetMessageHistory(matchID string, limit int) ([]model.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT id, match_id, room_id, content, persona_id, event_type, is_bot, created_at
	FROM (
	    SELECT id, match_id, room_id, content, persona_id, event_type, is_bot, created_at
	    FROM published_messages
	    WHERE match_id = $1
	    ORDER BY created_at DESC
	    LIMIT $2
	) m
	ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, matchID, limit)
	if err != nil {
		return nil, fmt.Errorf("query published_messages failed: %w", err)
	}
	defer rows.Close()

	var messages []model.ChatMessage
	for rows.Next() {
		var m model.ChatMessage
		if err := rows.Scan(
			&m.ID, &m.MatchID, &m.RoomID, &m.Content,
			&m.Persona, &m.EventType, &m.IsBot, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row failed: %w", err)
		}
		m.Timestamp = m.CreatedAt.Unix()
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
