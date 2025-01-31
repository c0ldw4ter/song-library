-- Создание таблицы songs (Songs table)
CREATE TABLE IF NOT EXISTS songs (
    id SERIAL PRIMARY KEY,
    group_name TEXT NOT NULL,
    song_name TEXT NOT NULL,
    release_date TEXT,
    text TEXT,
    link TEXT,
    album_cover_url TEXT,
    UNIQUE (group_name, song_name)
);

-- Создание индексов для улучшения производительности (Create indexes for high performance)
CREATE INDEX IF NOT EXISTS idx_songs_group_name ON songs(group_name);
CREATE INDEX IF NOT EXISTS idx_songs_song_name ON songs(song_name);