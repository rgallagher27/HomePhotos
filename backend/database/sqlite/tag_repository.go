package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rgallagher/homephotos/domain/tag"
)

type TagRepository struct {
	q  *Queries
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{q: New(db), db: db}
}

// Tag groups

func (r *TagRepository) CreateGroup(ctx context.Context, g *tag.TagGroup) (*tag.TagGroup, error) {
	row, err := r.q.CreateTagGroup(ctx, CreateTagGroupParams{
		Name:      g.Name,
		SortOrder: int64(g.SortOrder),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, tag.ErrDuplicateName
		}
		return nil, fmt.Errorf("create tag group: %w", err)
	}
	return rowToTagGroup(row), nil
}

func (r *TagRepository) GetGroupByID(ctx context.Context, id int64) (*tag.TagGroup, error) {
	row, err := r.q.GetTagGroupByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tag.ErrGroupNotFound
		}
		return nil, fmt.Errorf("get tag group: %w", err)
	}
	return rowToTagGroup(row), nil
}

func (r *TagRepository) UpdateGroup(ctx context.Context, g *tag.TagGroup) error {
	err := r.q.UpdateTagGroup(ctx, UpdateTagGroupParams{
		Name:      g.Name,
		SortOrder: int64(g.SortOrder),
		ID:        g.ID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return tag.ErrDuplicateName
		}
		return fmt.Errorf("update tag group: %w", err)
	}
	return nil
}

func (r *TagRepository) DeleteGroup(ctx context.Context, id int64) error {
	return r.q.DeleteTagGroup(ctx, id)
}

func (r *TagRepository) ListGroups(ctx context.Context) ([]tag.TagGroup, error) {
	rows, err := r.q.ListTagGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tag groups: %w", err)
	}
	groups := make([]tag.TagGroup, len(rows))
	for i, row := range rows {
		groups[i] = *rowToTagGroup(row)
	}
	return groups, nil
}

// Tags

func (r *TagRepository) CreateTag(ctx context.Context, t *tag.Tag) (*tag.Tag, error) {
	row, err := r.q.CreateTag(ctx, CreateTagParams{
		Name:      t.Name,
		Color:     toNullString(t.Color),
		GroupID:   toNullInt64(t.GroupID),
		CreatedBy: t.CreatedBy,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, tag.ErrDuplicateName
		}
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return &tag.Tag{
		ID:        row.ID,
		Name:      row.Name,
		Color:     row.Color.String,
		GroupID:   fromNullInt64(row.GroupID),
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *TagRepository) GetTagByID(ctx context.Context, id int64) (*tag.Tag, error) {
	row, err := r.q.GetTagByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tag.ErrNotFound
		}
		return nil, fmt.Errorf("get tag: %w", err)
	}
	return tagRowToTag(row), nil
}

func (r *TagRepository) UpdateTag(ctx context.Context, t *tag.Tag) error {
	err := r.q.UpdateTag(ctx, UpdateTagParams{
		Name:    t.Name,
		Color:   toNullString(t.Color),
		GroupID: toNullInt64(t.GroupID),
		ID:      t.ID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return tag.ErrDuplicateName
		}
		return fmt.Errorf("update tag: %w", err)
	}
	return nil
}

func (r *TagRepository) DeleteTag(ctx context.Context, id int64) error {
	return r.q.DeleteTag(ctx, id)
}

func (r *TagRepository) ListTags(ctx context.Context) ([]tag.Tag, error) {
	rows, err := r.q.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	tags := make([]tag.Tag, len(rows))
	for i, row := range rows {
		tags[i] = tag.Tag{
			ID:        row.ID,
			Name:      row.Name,
			Color:     row.Color.String,
			GroupID:   fromNullInt64(row.GroupID),
			GroupName: row.GroupName.String,
			CreatedBy: row.CreatedBy,
			CreatedAt: row.CreatedAt,
		}
	}
	return tags, nil
}

// Photo-tag associations

func (r *TagRepository) AddPhotoTags(ctx context.Context, photoID int64, tagIDs []int64, userID int64) error {
	for _, tagID := range tagIDs {
		if err := r.q.AddPhotoTag(ctx, AddPhotoTagParams{
			PhotoID:   photoID,
			TagID:     tagID,
			CreatedBy: userID,
		}); err != nil {
			return fmt.Errorf("add photo tag: %w", err)
		}
	}
	return nil
}

func (r *TagRepository) RemovePhotoTag(ctx context.Context, photoID int64, tagID int64) error {
	return r.q.RemovePhotoTag(ctx, RemovePhotoTagParams{
		PhotoID: photoID,
		TagID:   tagID,
	})
}

func (r *TagRepository) BulkAddPhotoTags(ctx context.Context, photoIDs []int64, tagIDs []int64, userID int64) error {
	const batchSize = 500
	var values []string
	var args []any

	for _, photoID := range photoIDs {
		for _, tagID := range tagIDs {
			values = append(values, "(?, ?, ?)")
			args = append(args, photoID, tagID, userID)

			if len(values) >= batchSize {
				if err := r.executeBulkInsert(ctx, values, args); err != nil {
					return err
				}
				values = values[:0]
				args = args[:0]
			}
		}
	}

	if len(values) > 0 {
		return r.executeBulkInsert(ctx, values, args)
	}
	return nil
}

func (r *TagRepository) executeBulkInsert(ctx context.Context, values []string, args []any) error {
	query := "INSERT OR IGNORE INTO photo_tags (photo_id, tag_id, created_by) VALUES " +
		strings.Join(values, ", ")
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("bulk add photo tags: %w", err)
	}
	return nil
}

func (r *TagRepository) ListTagsForPhoto(ctx context.Context, photoID int64) ([]tag.Tag, error) {
	rows, err := r.q.ListTagsForPhoto(ctx, photoID)
	if err != nil {
		return nil, fmt.Errorf("list tags for photo: %w", err)
	}
	tags := make([]tag.Tag, len(rows))
	for i, row := range rows {
		tags[i] = tag.Tag{
			ID:        row.ID,
			Name:      row.Name,
			Color:     row.Color.String,
			GroupID:   fromNullInt64(row.GroupID),
			GroupName: row.GroupName.String,
			CreatedBy: row.CreatedBy,
			CreatedAt: row.CreatedAt,
		}
	}
	return tags, nil
}

func (r *TagRepository) ListTagsForPhotos(ctx context.Context, photoIDs []int64) (map[int64][]tag.Tag, error) {
	if len(photoIDs) == 0 {
		return make(map[int64][]tag.Tag), nil
	}

	placeholders := make([]string, len(photoIDs))
	args := make([]any, len(photoIDs))
	for i, id := range photoIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT pt.photo_id, t.id, t.name, t.color, t.group_id, t.created_by, t.created_at,
	                 tg.name AS group_name
	          FROM photo_tags pt
	          JOIN tags t ON t.id = pt.tag_id
	          LEFT JOIN tag_groups tg ON tg.id = t.group_id
	          WHERE pt.photo_id IN (` + strings.Join(placeholders, ", ") + `)
	          ORDER BY t.name`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tags for photos: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]tag.Tag)
	for rows.Next() {
		var photoID int64
		var t tag.Tag
		var color, groupName sql.NullString
		var groupID sql.NullInt64
		if err := rows.Scan(&photoID, &t.ID, &t.Name, &color, &groupID, &t.CreatedBy, &t.CreatedAt, &groupName); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		t.Color = color.String
		t.GroupID = fromNullInt64(groupID)
		t.GroupName = groupName.String
		result[photoID] = append(result[photoID], t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return result, nil
}

// Helpers

func tagRowToTag(row GetTagByIDRow) *tag.Tag {
	return &tag.Tag{
		ID:        row.ID,
		Name:      row.Name,
		Color:     row.Color.String,
		GroupID:   fromNullInt64(row.GroupID),
		GroupName: row.GroupName.String,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}
}

func rowToTagGroup(row TagGroup) *tag.TagGroup {
	return &tag.TagGroup{
		ID:        row.ID,
		Name:      row.Name,
		SortOrder: int(row.SortOrder),
		CreatedAt: row.CreatedAt,
	}
}

