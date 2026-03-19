-- published_messages: lưu lại mọi message bot đã publish thành công
CREATE TABLE IF NOT EXISTS published_messages (
    id            VARCHAR(200) PRIMARY KEY,
    match_id      VARCHAR(100) NOT NULL,
    room_id       VARCHAR(100) NOT NULL,
    content       TEXT         NOT NULL,
    persona_id    VARCHAR(100),
    event_type    VARCHAR(50),
    is_bot        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_published_messages_match_id   ON published_messages(match_id);
CREATE INDEX IF NOT EXISTS idx_published_messages_created_at ON published_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_published_messages_event_type ON published_messages(event_type);