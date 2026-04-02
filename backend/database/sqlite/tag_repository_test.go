package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/domain/photo"
	"github.com/rgallagher/homephotos/domain/tag"
	"github.com/rgallagher/homephotos/domain/user"
)

func createTestUser(t *testing.T, repo *sqlite.UserRepository) *user.User {
	t.Helper()
	u, err := repo.Create(context.Background(), &user.User{
		Username:     "testadmin",
		PasswordHash: "hash",
		Role:         user.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func createTestPhoto2(t *testing.T, repo *sqlite.PhotoRepository, path string) *photo.Photo {
	t.Helper()
	p, err := repo.Create(context.Background(), &photo.Photo{
		FilePath:    path,
		FileName:    path,
		FileSize:    1024,
		FileMtime:   time.Now().Truncate(time.Second),
		Format:      "jpg",
		Fingerprint: path + "|fp",
		CacheStatus: photo.CacheStatusPending,
	})
	if err != nil {
		t.Fatalf("create photo: %v", err)
	}
	return p
}

func TestTagGroupRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	// Create
	g, err := repo.CreateGroup(ctx, &tag.TagGroup{Name: "People", SortOrder: 1})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if g.ID == 0 || g.Name != "People" || g.SortOrder != 1 {
		t.Errorf("unexpected group: %+v", g)
	}

	// Get
	got, err := repo.GetGroupByID(ctx, g.ID)
	if err != nil {
		t.Fatalf("get group: %v", err)
	}
	if got.Name != "People" {
		t.Errorf("name = %q, want People", got.Name)
	}

	// Update
	got.Name = "Persons"
	got.SortOrder = 5
	if err := repo.UpdateGroup(ctx, got); err != nil {
		t.Fatalf("update group: %v", err)
	}
	got2, _ := repo.GetGroupByID(ctx, g.ID)
	if got2.Name != "Persons" || got2.SortOrder != 5 {
		t.Errorf("after update: %+v", got2)
	}

	// List ordering
	repo.CreateGroup(ctx, &tag.TagGroup{Name: "AAA", SortOrder: 0})
	groups, err := repo.ListGroups(ctx)
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	if len(groups) != 2 || groups[0].Name != "AAA" {
		t.Errorf("list order: got %v", groups)
	}

	// Delete
	if err := repo.DeleteGroup(ctx, g.ID); err != nil {
		t.Fatalf("delete group: %v", err)
	}
	_, err = repo.GetGroupByID(ctx, g.ID)
	if err != tag.ErrGroupNotFound {
		t.Errorf("after delete: err = %v, want ErrGroupNotFound", err)
	}
}

func TestTagGroupRepository_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewTagRepository(db)
	ctx := context.Background()

	repo.CreateGroup(ctx, &tag.TagGroup{Name: "People"})
	_, err := repo.CreateGroup(ctx, &tag.TagGroup{Name: "People"})
	if err != tag.ErrDuplicateName {
		t.Errorf("err = %v, want ErrDuplicateName", err)
	}
}

func TestTagRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	g, _ := tagRepo.CreateGroup(ctx, &tag.TagGroup{Name: "Places"})

	// Create with group and color
	created, err := tagRepo.CreateTag(ctx, &tag.Tag{
		Name:      "Beach",
		Color:     "#3498db",
		GroupID:   &g.ID,
		CreatedBy: u.ID,
	})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if created.Name != "Beach" || created.Color != "#3498db" || *created.GroupID != g.ID {
		t.Errorf("unexpected tag: %+v", created)
	}

	// Create ungrouped
	ungrouped, err := tagRepo.CreateTag(ctx, &tag.Tag{
		Name:      "Misc",
		CreatedBy: u.ID,
	})
	if err != nil {
		t.Fatalf("create ungrouped: %v", err)
	}
	if ungrouped.GroupID != nil {
		t.Errorf("expected nil GroupID, got %v", ungrouped.GroupID)
	}

	// Get with group name
	got, err := tagRepo.GetTagByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get tag: %v", err)
	}
	if got.GroupName != "Places" {
		t.Errorf("group_name = %q, want Places", got.GroupName)
	}

	// Update
	got.Name = "Mountain"
	got.Color = "#e74c3c"
	if err := tagRepo.UpdateTag(ctx, got); err != nil {
		t.Fatalf("update tag: %v", err)
	}
	got2, _ := tagRepo.GetTagByID(ctx, created.ID)
	if got2.Name != "Mountain" || got2.Color != "#e74c3c" {
		t.Errorf("after update: %+v", got2)
	}

	// List
	tags, err := tagRepo.ListTags(ctx)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("list count = %d, want 2", len(tags))
	}

	// Delete
	if err := tagRepo.DeleteTag(ctx, created.ID); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	_, err = tagRepo.GetTagByID(ctx, created.ID)
	if err != tag.ErrNotFound {
		t.Errorf("after delete: err = %v, want ErrNotFound", err)
	}
}

func TestTagRepository_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	g1, _ := tagRepo.CreateGroup(ctx, &tag.TagGroup{Name: "Group1"})
	g2, _ := tagRepo.CreateGroup(ctx, &tag.TagGroup{Name: "Group2"})

	// Same name in same group = error
	tagRepo.CreateTag(ctx, &tag.Tag{Name: "Beach", GroupID: &g1.ID, CreatedBy: u.ID})
	_, err := tagRepo.CreateTag(ctx, &tag.Tag{Name: "Beach", GroupID: &g1.ID, CreatedBy: u.ID})
	if err != tag.ErrDuplicateName {
		t.Errorf("same group: err = %v, want ErrDuplicateName", err)
	}

	// Same name in different group = ok
	_, err = tagRepo.CreateTag(ctx, &tag.Tag{Name: "Beach", GroupID: &g2.ID, CreatedBy: u.ID})
	if err != nil {
		t.Errorf("different group: unexpected err = %v", err)
	}
}

func TestTagRepository_DuplicateUngrouped(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)

	tagRepo.CreateTag(ctx, &tag.Tag{Name: "Misc", CreatedBy: u.ID})
	_, err := tagRepo.CreateTag(ctx, &tag.Tag{Name: "Misc", CreatedBy: u.ID})
	if err != tag.ErrDuplicateName {
		t.Errorf("duplicate ungrouped: err = %v, want ErrDuplicateName", err)
	}
}

func TestTagRepository_DeleteGroupSetsNull(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	g, _ := tagRepo.CreateGroup(ctx, &tag.TagGroup{Name: "ToDelete"})
	created, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "Orphan", GroupID: &g.ID, CreatedBy: u.ID})

	tagRepo.DeleteGroup(ctx, g.ID)

	got, err := tagRepo.GetTagByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get after group delete: %v", err)
	}
	if got.GroupID != nil {
		t.Errorf("group_id = %v, want nil", got.GroupID)
	}
}

func TestTagRepository_DeleteCascadesToPhotoTags(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	photoRepo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	p := createTestPhoto2(t, photoRepo, "cascade.jpg")
	created, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "ToDelete", CreatedBy: u.ID})
	tagRepo.AddPhotoTags(ctx, p.ID, []int64{created.ID}, u.ID)

	tagRepo.DeleteTag(ctx, created.ID)

	tags, _ := tagRepo.ListTagsForPhoto(ctx, p.ID)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags after cascade, got %d", len(tags))
	}
}

func TestPhotoTagRepository_AddRemoveList(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	photoRepo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	p := createTestPhoto2(t, photoRepo, "tagged.jpg")
	t1, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "Tag1", CreatedBy: u.ID})
	t2, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "Tag2", CreatedBy: u.ID})

	// Add
	if err := tagRepo.AddPhotoTags(ctx, p.ID, []int64{t1.ID, t2.ID}, u.ID); err != nil {
		t.Fatalf("add: %v", err)
	}

	// List
	tags, err := tagRepo.ListTagsForPhoto(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("count = %d, want 2", len(tags))
	}

	// Remove
	tagRepo.RemovePhotoTag(ctx, p.ID, t1.ID)
	tags, _ = tagRepo.ListTagsForPhoto(ctx, p.ID)
	if len(tags) != 1 || tags[0].Name != "Tag2" {
		t.Errorf("after remove: %v", tags)
	}
}

func TestPhotoTagRepository_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	photoRepo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	p := createTestPhoto2(t, photoRepo, "idempotent.jpg")
	created, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "Tag1", CreatedBy: u.ID})

	// Add twice — no error
	tagRepo.AddPhotoTags(ctx, p.ID, []int64{created.ID}, u.ID)
	if err := tagRepo.AddPhotoTags(ctx, p.ID, []int64{created.ID}, u.ID); err != nil {
		t.Fatalf("second add: %v", err)
	}

	tags, _ := tagRepo.ListTagsForPhoto(ctx, p.ID)
	if len(tags) != 1 {
		t.Errorf("count = %d, want 1", len(tags))
	}
}

func TestPhotoTagRepository_BulkAdd(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	photoRepo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	p1 := createTestPhoto2(t, photoRepo, "bulk1.jpg")
	p2 := createTestPhoto2(t, photoRepo, "bulk2.jpg")
	p3 := createTestPhoto2(t, photoRepo, "bulk3.jpg")
	t1, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "BulkTag1", CreatedBy: u.ID})
	t2, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "BulkTag2", CreatedBy: u.ID})

	err := tagRepo.BulkAddPhotoTags(ctx, []int64{p1.ID, p2.ID, p3.ID}, []int64{t1.ID, t2.ID}, u.ID)
	if err != nil {
		t.Fatalf("bulk add: %v", err)
	}

	for _, p := range []*photo.Photo{p1, p2, p3} {
		tags, _ := tagRepo.ListTagsForPhoto(ctx, p.ID)
		if len(tags) != 2 {
			t.Errorf("photo %d: count = %d, want 2", p.ID, len(tags))
		}
	}
}

func TestPhotoTagRepository_ListTagsForPhotos(t *testing.T) {
	db := setupTestDB(t)
	tagRepo := sqlite.NewTagRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	photoRepo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	u := createTestUser(t, userRepo)
	p1 := createTestPhoto2(t, photoRepo, "batch1.jpg")
	p2 := createTestPhoto2(t, photoRepo, "batch2.jpg")
	t1, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "A", CreatedBy: u.ID})
	t2, _ := tagRepo.CreateTag(ctx, &tag.Tag{Name: "B", CreatedBy: u.ID})

	tagRepo.AddPhotoTags(ctx, p1.ID, []int64{t1.ID, t2.ID}, u.ID)
	tagRepo.AddPhotoTags(ctx, p2.ID, []int64{t1.ID}, u.ID)

	result, err := tagRepo.ListTagsForPhotos(ctx, []int64{p1.ID, p2.ID})
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(result[p1.ID]) != 2 {
		t.Errorf("p1 tags = %d, want 2", len(result[p1.ID]))
	}
	if len(result[p2.ID]) != 1 {
		t.Errorf("p2 tags = %d, want 1", len(result[p2.ID]))
	}
}
