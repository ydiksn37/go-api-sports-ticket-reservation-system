-- 1. テーブルの作成
CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seat_grade VARCHAR(50) NOT NULL,
    price INTEGER NOT NULL,
    total_seats INTEGER NOT NULL
);

-- 2. インデックスの作成
CREATE INDEX idx_seats_seat_grade ON seats(seat_grade);