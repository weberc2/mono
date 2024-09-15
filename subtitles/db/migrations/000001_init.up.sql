CREATE TABLE showvideofiles (
    filepath TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    year TEXT NOT NULL,
    season TEXT NOT NULL,
    episode TEXT NOT NULL
);

CREATE TABLE showsubtitlefiles (
    filepath TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    year TEXT NOT NULL,
    season TEXT NOT NULL,
    episode TEXT NOT NULL,
    language TEXT NOT NULL
);

CREATE TYPE downloadstatus AS ENUM (
    'PENDING',
    'SEARCHING',
    'FETCHING_URL',
    'DOWNLOADING',
    'COMPLETE'
);

CREATE TABLE showdownloads (
    title TEXT NOT NULL,
    year TEXT NOT NULL,
    season TEXT NOT NULL,
    episode TEXT NOT NULL,
    language TEXT NOT NULL,
    opensubtitlesid TEXT NOT NULL,
    url TEXT NOT NULL,
    filepath TEXT NOT NULL,
    status downloadstatus NOT NULL,
    created TIMESTAMPTZ NOT NULL,
    reservationexpiry TIMESTAMPTZ NOT NULL,
    PRIMARY KEY(title, year, season, episode, language)
);
