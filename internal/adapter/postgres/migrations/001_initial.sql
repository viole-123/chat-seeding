-- Create matches table for storing match events
CREATE TABLE IF NOT EXISTS matches (
    id VARCHAR(100) PRIMARY KEY,
    status INT NOT NULL,
    phase VARCHAR(30) NOT NULL DEFAULT 'not_started',
    home_score INT[] DEFAULT ARRAY[0,0,0,0,0,0,0],
    away_score INT[] DEFAULT ARRAY[0,0,0,0,0,0,0],
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create match_stats table
CREATE TABLE IF NOT EXISTS match_stats (
    id SERIAL PRIMARY KEY,
    match_id VARCHAR(100) REFERENCES matches(id) ON DELETE CASCADE,
    stat_type INT NOT NULL,
    home_value INT NOT NULL DEFAULT 0,
    away_value INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create match_incidents table
CREATE TABLE IF NOT EXISTS match_incidents (
    id SERIAL PRIMARY KEY,
    match_id VARCHAR(100) REFERENCES matches(id) ON DELETE CASCADE,
    incident_type INT NOT NULL,
    position INT NOT NULL,
    time INT NOT NULL,
    add_time INT DEFAULT 0,
    player_id VARCHAR(100),
    player_name TEXT,
    in_player_id VARCHAR(100),
    in_player_name TEXT,
    out_player_id VARCHAR(100),
    out_player_name TEXT,
    home_score INT,
    away_score INT,
    reason_type INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_matches_timestamp ON matches(timestamp);
CREATE INDEX IF NOT EXISTS idx_matches_phase ON matches(phase);
CREATE INDEX IF NOT EXISTS idx_match_stats_match_id ON match_stats(match_id);
CREATE INDEX IF NOT EXISTS idx_match_incidents_match_id ON match_incidents(match_id);
CREATE INDEX IF NOT EXISTS idx_match_incidents_type ON match_incidents(incident_type);