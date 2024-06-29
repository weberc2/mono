CREATE TABLE showvideofiles (
    filepath TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    -- TODO: year TEXT NOT NULL,
    season TEXT NOT NULL,
    episode TEXT NOT NULL,
    mediahash TEXT NOT NULL
);

CREATE TABLE showsubtitles (
    filepath TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    -- TODO: year TEXT NOT NULL,
    season TEXT NOT NULL,
    episode TEXT NOT NULL,
    language TEXT NOT NULL
);

CREATE TABLE filmvideofiles (
    filepath TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    year TEXT NOT NULL,
    mediahash TEXT NOT NULL
);

CREATE TABLE filmsubtitles (
    filepath TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    year TEXT NOT NULL,
    language TEXT NOT NULL
);