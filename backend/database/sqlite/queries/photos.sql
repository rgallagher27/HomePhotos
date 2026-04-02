-- name: CreatePhoto :one
INSERT INTO photos (
    file_path, file_name, file_size, file_mtime, format,
    width, height, captured_at, camera_make, camera_model, lens_model,
    focal_length, aperture, shutter_speed, iso, orientation,
    gps_latitude, gps_longitude, fingerprint, cache_status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, file_path, file_name, file_size, file_mtime, format,
    width, height, captured_at, camera_make, camera_model, lens_model,
    focal_length, aperture, shutter_speed, iso, orientation,
    gps_latitude, gps_longitude, fingerprint, scanned_at, cache_status;

-- name: UpdatePhoto :exec
UPDATE photos SET
    file_size = ?, file_mtime = ?, format = ?,
    width = ?, height = ?, captured_at = ?,
    camera_make = ?, camera_model = ?, lens_model = ?,
    focal_length = ?, aperture = ?, shutter_speed = ?,
    iso = ?, orientation = ?,
    gps_latitude = ?, gps_longitude = ?,
    fingerprint = ?, scanned_at = CURRENT_TIMESTAMP,
    cache_status = ?
WHERE id = ?;

-- name: GetPhotoByID :one
SELECT id, file_path, file_name, file_size, file_mtime, format,
    width, height, captured_at, camera_make, camera_model, lens_model,
    focal_length, aperture, shutter_speed, iso, orientation,
    gps_latitude, gps_longitude, fingerprint, scanned_at, cache_status
FROM photos WHERE id = ?;

-- name: GetPhotoByFilePath :one
SELECT id, file_path, file_name, file_size, file_mtime, format,
    width, height, captured_at, camera_make, camera_model, lens_model,
    focal_length, aperture, shutter_speed, iso, orientation,
    gps_latitude, gps_longitude, fingerprint, scanned_at, cache_status
FROM photos WHERE file_path = ?;

-- name: ListAllFingerprints :many
SELECT file_path, fingerprint FROM photos;
