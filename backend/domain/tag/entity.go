package tag

import "time"

type TagGroup struct {
	ID        int64
	Name      string
	SortOrder int
	CreatedAt time.Time
}

type Tag struct {
	ID        int64
	Name      string
	Color     string // empty if null
	GroupID   *int64
	GroupName string // populated from joins
	CreatedBy int64
	CreatedAt time.Time
}
