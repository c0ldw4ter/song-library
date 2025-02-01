-- Songs table(Создание таблицы songs)
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

-- Create indexes for high performance(Создание индексов для улучшения производительности)
CREATE INDEX IF NOT EXISTS idx_songs_group_name ON songs(group_name);
CREATE INDEX IF NOT EXISTS idx_songs_song_name ON songs(song_name);