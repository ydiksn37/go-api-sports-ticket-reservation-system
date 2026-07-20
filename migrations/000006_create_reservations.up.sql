-- 1. テーブルの作成
CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    game_id UUID NOT NULL REFERENCES games(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. インデックスの作成（自分の予約一覧取得や、試合ごとの集計、ステータスによる期限切れチェックを高速化）
CREATE INDEX idx_reservations_user_id ON reservations(user_id);
CREATE INDEX idx_reservations_game_id ON reservations(game_id);
CREATE INDEX idx_reservations_status ON reservations(status);

-- 3. 自動更新トリガーの作成
CREATE TRIGGER update_reservations_updated_at BEFORE UPDATE
    ON reservations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();