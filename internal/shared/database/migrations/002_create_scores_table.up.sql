CREATE TABLE IF NOT EXISTS scores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id VARCHAR(100) NOT NULL,
    score BIGINT NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scores_user_id ON scores(user_id);
CREATE INDEX idx_scores_game_id ON scores(game_id);
CREATE INDEX idx_scores_submitted_at ON scores(submitted_at);
CREATE INDEX idx_scores_user_game ON scores(user_id, game_id);
CREATE INDEX idx_scores_score_desc ON scores(score DESC);

