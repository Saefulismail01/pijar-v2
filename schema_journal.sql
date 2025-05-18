-- Tabel journals
CREATE TABLE IF NOT EXISTS journals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title TEXT,
    content TEXT,
    feeling VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel journal_analyses
CREATE TABLE IF NOT EXISTS journal_analyses (
    id SERIAL PRIMARY KEY,
    journal_id INTEGER NOT NULL REFERENCES journals(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sentiment_score REAL,
    sentiment_label VARCHAR(50),
    analyzed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel trend_analyses
CREATE TABLE IF NOT EXISTS trend_analyses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    period_type VARCHAR(50), -- e.g., 'weekly', 'monthly'
    average_sentiment REAL,
    entry_count INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Optional: Indeks untuk optimasi query
CREATE INDEX IF NOT EXISTS idx_journal_analyses_user_id ON journal_analyses(user_id);
CREATE INDEX IF NOT EXISTS idx_trend_analyses_user_id ON trend_analyses(user_id);
CREATE INDEX IF NOT EXISTS idx_journal_analyses_analyzed_at ON journal_analyses(analyzed_at);

-- Insert sample journals
INSERT INTO journals (user_id, judul, isi, perasaan) VALUES
(1, 'Hari Pertama Kerja', 'Hari ini saya mulai kerja di tempat baru.', 'senang'),
(1, 'Proyek Baru', 'Dapat tugas baru dari atasan, sedikit menegangkan.', 'cemas'),
(2, 'Liburan ke Pantai', 'Akhirnya bisa liburan ke pantai setelah sekian lama.', 'bahagia');
