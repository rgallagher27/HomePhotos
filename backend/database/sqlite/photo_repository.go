package sqlite

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rgallagher/homephotos/domain/photo"
)

type PhotoRepository struct {
	q  *Queries
	db DBTX
}

func NewPhotoRepository(db DBTX) *PhotoRepository {
	return &PhotoRepository{q: New(db), db: db}
}

func (r *PhotoRepository) Create(ctx context.Context, p *photo.Photo) (*photo.Photo, error) {
	row, err := r.q.CreatePhoto(ctx, CreatePhotoParams{
		FilePath:     p.FilePath,
		FileName:     p.FileName,
		FileSize:     p.FileSize,
		FileMtime:    p.FileMtime,
		Format:       p.Format,
		Width:        toNullInt64(p.Width),
		Height:       toNullInt64(p.Height),
		CapturedAt:   toNullTime(p.CapturedAt),
		CameraMake:   toNullString(p.CameraMake),
		CameraModel:  toNullString(p.CameraModel),
		LensModel:    toNullString(p.LensModel),
		FocalLength:  toNullFloat64(p.FocalLength),
		Aperture:     toNullFloat64(p.Aperture),
		ShutterSpeed: toNullString(p.ShutterSpeed),
		Iso:          toNullInt64(p.ISO),
		Orientation:  toNullInt64(p.Orientation),
		GpsLatitude:  toNullFloat64(p.GPSLatitude),
		GpsLongitude: toNullFloat64(p.GPSLongitude),
		Fingerprint:  p.Fingerprint,
		CacheStatus:  string(p.CacheStatus),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, photo.ErrDuplicatePath
		}
		return nil, fmt.Errorf("create photo: %w", err)
	}
	return rowToPhoto(row), nil
}

func (r *PhotoRepository) Update(ctx context.Context, p *photo.Photo) error {
	return r.q.UpdatePhoto(ctx, UpdatePhotoParams{
		ID:           p.ID,
		FileSize:     p.FileSize,
		FileMtime:    p.FileMtime,
		Format:       p.Format,
		Width:        toNullInt64(p.Width),
		Height:       toNullInt64(p.Height),
		CapturedAt:   toNullTime(p.CapturedAt),
		CameraMake:   toNullString(p.CameraMake),
		CameraModel:  toNullString(p.CameraModel),
		LensModel:    toNullString(p.LensModel),
		FocalLength:  toNullFloat64(p.FocalLength),
		Aperture:     toNullFloat64(p.Aperture),
		ShutterSpeed: toNullString(p.ShutterSpeed),
		Iso:          toNullInt64(p.ISO),
		Orientation:  toNullInt64(p.Orientation),
		GpsLatitude:  toNullFloat64(p.GPSLatitude),
		GpsLongitude: toNullFloat64(p.GPSLongitude),
		Fingerprint:  p.Fingerprint,
		CacheStatus:  string(p.CacheStatus),
	})
}

func (r *PhotoRepository) GetByID(ctx context.Context, id int64) (*photo.Photo, error) {
	row, err := r.q.GetPhotoByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, photo.ErrNotFound
		}
		return nil, fmt.Errorf("get photo by id: %w", err)
	}
	return rowToPhoto(row), nil
}

func (r *PhotoRepository) GetByFilePath(ctx context.Context, filePath string) (*photo.Photo, error) {
	row, err := r.q.GetPhotoByFilePath(ctx, filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, photo.ErrNotFound
		}
		return nil, fmt.Errorf("get photo by file path: %w", err)
	}
	return rowToPhoto(row), nil
}

func (r *PhotoRepository) ListAllFingerprints(ctx context.Context) (map[string]string, error) {
	rows, err := r.q.ListAllFingerprints(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all fingerprints: %w", err)
	}
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row.FilePath] = row.Fingerprint
	}
	return result, nil
}

func (r *PhotoRepository) DeleteOrphaned(ctx context.Context, activePaths []string) (int64, error) {
	if len(activePaths) == 0 {
		// Delete all photos if no active paths
		res, err := r.db.ExecContext(ctx, "DELETE FROM photos")
		if err != nil {
			return 0, fmt.Errorf("delete all photos: %w", err)
		}
		return res.RowsAffected()
	}

	// Use a temporary table for efficiency with large sets
	_, err := r.db.ExecContext(ctx, "CREATE TEMP TABLE IF NOT EXISTS active_paths (path TEXT PRIMARY KEY)")
	if err != nil {
		return 0, fmt.Errorf("create temp table: %w", err)
	}
	defer r.db.ExecContext(ctx, "DROP TABLE IF EXISTS temp.active_paths") //nolint:errcheck

	// Batch insert active paths
	for i := 0; i < len(activePaths); i += 500 {
		end := i + 500
		if end > len(activePaths) {
			end = len(activePaths)
		}
		batch := activePaths[i:end]

		placeholders := make([]string, len(batch))
		args := make([]any, len(batch))
		for j, p := range batch {
			placeholders[j] = "(?)"
			args[j] = p
		}

		query := "INSERT OR IGNORE INTO temp.active_paths (path) VALUES " + strings.Join(placeholders, ",")
		if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
			return 0, fmt.Errorf("insert active paths: %w", err)
		}
	}

	res, err := r.db.ExecContext(ctx,
		"DELETE FROM photos WHERE file_path NOT IN (SELECT path FROM temp.active_paths)")
	if err != nil {
		return 0, fmt.Errorf("delete orphaned photos: %w", err)
	}
	return res.RowsAffected()
}

func (r *PhotoRepository) ListPending(ctx context.Context, limit int) ([]photo.Photo, error) {
	rows, err := r.q.ListPendingPhotos(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("list pending photos: %w", err)
	}
	photos := make([]photo.Photo, len(rows))
	for i, row := range rows {
		photos[i] = *rowToPhoto(row)
	}
	return photos, nil
}

func (r *PhotoRepository) UpdateCacheStatus(ctx context.Context, id int64, status photo.CacheStatus) error {
	return r.q.UpdatePhotoCacheStatus(ctx, UpdatePhotoCacheStatusParams{
		CacheStatus: string(status),
		ID:          id,
	})
}

func (r *PhotoRepository) List(ctx context.Context, params photo.ListParams) (*photo.ListResult, error) {
	sortCol := "captured_at"
	if params.SortBy == "file_name" {
		sortCol = "file_name"
	}

	direction := "DESC"
	if params.SortOrder == "asc" {
		direction = "ASC"
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var where []string
	var args []any

	// Cursor pagination
	if params.Cursor != "" {
		cursorVal, cursorID, err := decodeCursor(params.Cursor)
		if err == nil {
			// For datetime columns, parse the cursor value back to time.Time
			// so the driver passes it with the same type as the stored value.
			var cursorArg any = cursorVal
			if sortCol == "captured_at" {
				if t, err := time.Parse(time.RFC3339Nano, cursorVal); err == nil {
					cursorArg = t
				}
			}
			if direction == "DESC" {
				where = append(where, fmt.Sprintf("(%s < ? OR (%s = ? AND id < ?))", sortCol, sortCol))
			} else {
				where = append(where, fmt.Sprintf("(%s > ? OR (%s = ? AND id > ?))", sortCol, sortCol))
			}
			args = append(args, cursorArg, cursorArg, cursorID)
		}
	}

	// Filters
	if params.DateFrom != nil {
		where = append(where, "captured_at >= ?")
		args = append(args, params.DateFrom.UTC().Format(time.RFC3339))
	}
	if params.DateTo != nil {
		where = append(where, "captured_at <= ?")
		args = append(args, params.DateTo.UTC().Format(time.RFC3339))
	}
	if params.Folder != "" {
		where = append(where, "file_path LIKE ?")
		args = append(args, params.Folder+"%")
	}
	if params.Format != "" {
		where = append(where, "format = ?")
		args = append(args, strings.ToLower(params.Format))
	}

	query := "SELECT id, file_path, file_name, file_size, file_mtime, format, " +
		"width, height, captured_at, camera_make, camera_model, lens_model, " +
		"focal_length, aperture, shutter_speed, iso, orientation, " +
		"gps_latitude, gps_longitude, fingerprint, scanned_at, cache_status " +
		"FROM photos"

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY %s %s, id %s LIMIT ?", sortCol, direction, direction)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list photos: %w", err)
	}
	defer rows.Close()

	var photos []photo.Photo
	for rows.Next() {
		var row Photo
		if err := rows.Scan(
			&row.ID, &row.FilePath, &row.FileName, &row.FileSize, &row.FileMtime, &row.Format,
			&row.Width, &row.Height, &row.CapturedAt, &row.CameraMake, &row.CameraModel, &row.LensModel,
			&row.FocalLength, &row.Aperture, &row.ShutterSpeed, &row.Iso, &row.Orientation,
			&row.GpsLatitude, &row.GpsLongitude, &row.Fingerprint, &row.ScannedAt, &row.CacheStatus,
		); err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		photos = append(photos, *rowToPhoto(row))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	hasMore := len(photos) > limit
	if hasMore {
		photos = photos[:limit]
	}

	var nextCursor string
	if hasMore && len(photos) > 0 {
		last := photos[len(photos)-1]
		var sortVal string
		if sortCol == "file_name" {
			sortVal = last.FileName
		} else {
			if last.CapturedAt != nil {
				sortVal = last.CapturedAt.UTC().Format(time.RFC3339Nano)
			}
		}
		nextCursor = encodeCursor(sortVal, last.ID)
	}

	return &photo.ListResult{
		Photos:     photos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func rowToPhoto(row Photo) *photo.Photo {
	return &photo.Photo{
		ID:           row.ID,
		FilePath:     row.FilePath,
		FileName:     row.FileName,
		FileSize:     row.FileSize,
		FileMtime:    row.FileMtime,
		Format:       row.Format,
		Width:        fromNullInt64(row.Width),
		Height:       fromNullInt64(row.Height),
		CapturedAt:   fromNullTime(row.CapturedAt),
		CameraMake:   fromNullString(row.CameraMake),
		CameraModel:  fromNullString(row.CameraModel),
		LensModel:    fromNullString(row.LensModel),
		FocalLength:  fromNullFloat64(row.FocalLength),
		Aperture:     fromNullFloat64(row.Aperture),
		ShutterSpeed: fromNullString(row.ShutterSpeed),
		ISO:          fromNullInt64(row.Iso),
		Orientation:  fromNullInt64(row.Orientation),
		GPSLatitude:  fromNullFloat64(row.GpsLatitude),
		GPSLongitude: fromNullFloat64(row.GpsLongitude),
		Fingerprint:  row.Fingerprint,
		ScannedAt:    row.ScannedAt,
		CacheStatus:  photo.CacheStatus(row.CacheStatus),
	}
}

func toNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func fromNullInt64(v sql.NullInt64) *int64 {
	if v.Valid {
		return &v.Int64
	}
	return nil
}

func toNullFloat64(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}

func fromNullFloat64(v sql.NullFloat64) *float64 {
	if v.Valid {
		return &v.Float64
	}
	return nil
}

func toNullTime(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}

func encodeCursor(sortVal string, id int64) string {
	raw := sortVal + "_" + strconv.FormatInt(id, 10)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (string, int64, error) {
	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", 0, err
	}
	s := string(data)
	idx := strings.LastIndex(s, "_")
	if idx < 0 {
		return "", 0, fmt.Errorf("invalid cursor format")
	}
	id, err := strconv.ParseInt(s[idx+1:], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid cursor id: %w", err)
	}
	return s[:idx], id, nil
}
