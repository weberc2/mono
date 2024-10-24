CREATE TYPE downloadstatus AS ENUM(
    'PENDING',
    'METADATA',
    'PROGRESS',
    'COMPLETE');

CREATE TABLE downloads (
    id VARCHAR(256) NOT NULL PRIMARY KEY,
    status downloadstatus NOT NULL,
    size NUMERIC NOT NULL,
    progress NUMERIC NOT NULL);

CREATE TABLE downloadfiles (
    download VARCHAR(256) NOT NULL REFERENCES downloads(id) ON DELETE CASCADE,
    path VARCHAR(2048) NOT NULL,
    size NUMERIC NOT NULL,
    progress NUMERIC NOT NULL,
    PRIMARY KEY(download, path));

CREATE TABLE imports (
    id VARCHAR(256) NOT NULL PRIMARY KEY,
    spec JSONB NOT NULL,
    status JSONB NOT NULL);