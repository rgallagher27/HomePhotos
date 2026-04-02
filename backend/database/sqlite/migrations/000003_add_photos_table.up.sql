CREATE TABLE photos (
    id              INTEGER PRIMARY KEY,
    file_path       TEXT NOT NULL UNIQUE,
    file_name       TEXT NOT NULL,
    file_size       INTEGER NOT NULL,
    file_mtime      DATETIME NOT NULL,
    format          TEXT NOT NULL,
    width           INTEGER,
    height          INTEGER,
    captured_at     DATETIME,
    camera_make     TEXT,
    camera_model    TEXT,
    lens_model      TEXT,
    focal_length    REAL,
    aperture        REAL,
    shutter_speed   TEXT,
    iso             INTEGER,
    orientation     INTEGER,
    gps_latitude    REAL,
    gps_longitude   REAL,
    fingerprint     TEXT NOT NULL,
    scanned_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cache_status    TEXT NOT NULL DEFAULT 'pending' CHECK (cache_status IN ('pending', 'cached', 'error'))
);

CREATE INDEX idx_photos_captured_at ON photos (captured_at);
CREATE INDEX idx_photos_file_path ON photos (file_path);
CREATE INDEX idx_photos_cache_status ON photos (cache_status);
CREATE INDEX idx_photos_format ON photos (format);
