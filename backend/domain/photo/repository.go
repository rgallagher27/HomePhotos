package photo

import (
	"context"
	"time"
)

type ListParams struct {
	SortBy    string // "captured_at" or "file_name"
	SortOrder string // "asc" or "desc"
	Cursor    string // opaque, base64-encoded
	Limit     int
	DateFrom  *time.Time
	DateTo    *time.Time
	Folder    string
	Format    string
}

type ListResult struct {
	Photos     []Photo
	NextCursor string
	HasMore    bool
}

type Repository interface {
	Create(ctx context.Context, p *Photo) (*Photo, error)
	Update(ctx context.Context, p *Photo) error
	GetByID(ctx context.Context, id int64) (*Photo, error)
	GetByFilePath(ctx context.Context, filePath string) (*Photo, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	DeleteOrphaned(ctx context.Context, activePaths []string) (int64, error)
	ListAllFingerprints(ctx context.Context) (map[string]string, error)
}
