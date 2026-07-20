-- 1. テーブルの作成
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id),
    seat_id UUID NOT NULL REFERENCES seats(id),
    reservation_id UUID REFERENCES reservations(id),
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. インデックスの作成（試合ごとの空席検索、予約に紐づくチケット取得を高速化）
CREATE INDEX idx_tickets_game_id ON tickets(game_id);
CREATE INDEX idx_tickets_seat_id ON tickets(seat_id);
CREATE INDEX idx_tickets_reservation_id ON tickets(reservation_id);
CREATE INDEX idx_tickets_status ON tickets(status);

-- 3. 自動更新トリガーの作成
CREATE TRIGGER update_tickets_updated_at BEFORE UPDATE
    ON tickets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();