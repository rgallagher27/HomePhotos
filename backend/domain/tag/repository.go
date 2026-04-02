package tag

import "context"

type Repository interface {
	// Tag groups
	CreateGroup(ctx context.Context, g *TagGroup) (*TagGroup, error)
	GetGroupByID(ctx context.Context, id int64) (*TagGroup, error)
	UpdateGroup(ctx context.Context, g *TagGroup) error
	DeleteGroup(ctx context.Context, id int64) error
	ListGroups(ctx context.Context) ([]TagGroup, error)

	// Tags
	CreateTag(ctx context.Context, t *Tag) (*Tag, error)
	GetTagByID(ctx context.Context, id int64) (*Tag, error)
	UpdateTag(ctx context.Context, t *Tag) error
	DeleteTag(ctx context.Context, id int64) error
	ListTags(ctx context.Context) ([]Tag, error)

	// Photo-tag associations
	AddPhotoTags(ctx context.Context, photoID int64, tagIDs []int64, userID int64) error
	RemovePhotoTag(ctx context.Context, photoID int64, tagID int64) error
	BulkAddPhotoTags(ctx context.Context, photoIDs []int64, tagIDs []int64, userID int64) error
	ListTagsForPhoto(ctx context.Context, photoID int64) ([]Tag, error)
	ListTagsForPhotos(ctx context.Context, photoIDs []int64) (map[int64][]Tag, error)
}
