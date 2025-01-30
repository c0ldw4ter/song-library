CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    group_name TEXT NOT NULL,
    song_name TEXT NOT NULL,
    release_date TEXT,
    text TEXT,
    link TEXT,
    album_cover_url TEXT,
    UNIQUE (group_name, song_name)
);