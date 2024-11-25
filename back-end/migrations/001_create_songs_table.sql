CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    group_name VARCHAR(255) NOT NULL,
    song_name VARCHAR(255) NOT NULL,
    release_date VARCHAR(255),
    text TEXT,
    link VARCHAR(255),
     album_cover_url VARCHAR(255)
);
