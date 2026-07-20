-- 1. テーブルの作成
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL
);

-- 2. インデックスの作成（チーム名で検索する可能性を考慮）
CREATE INDEX idx_teams_name ON teams(name);