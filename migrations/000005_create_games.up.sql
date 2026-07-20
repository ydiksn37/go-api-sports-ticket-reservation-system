-- 1. テーブルの作成
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    home_team_id UUID NOT NULL REFERENCES teams(id),
    away_team_id UUID NOT NULL REFERENCES teams(id),
    game_date TIMESTAMP NOT NULL,
    start_time TIMESTAMP NOT NULL
);

-- 2. インデックスの作成（チームごとの試合検索や、日時順での一覧取得を高速化）
CREATE INDEX idx_games_home_team_id ON games(home_team_id);
CREATE INDEX idx_games_away_team_id ON games(away_team_id);
CREATE INDEX idx_games_game_date ON games(game_date);