# Feature: Tagging

## Overview

HomePhotos provides a tag-based organization system that lets users categorize photos without modifying the original files on disk. All tag data is stored exclusively in the SQLite database, completely independent of the source filesystem. This preserves the read-only constraint on the SMB share while giving users a flexible way to organize, search, and filter their photo library.

---

## Data Model (SQLite Schema)

### `tag_groups`

Organizational containers for related tags.

| Column       | Type                  | Constraints              |
| ------------ | --------------------- | ------------------------ |
| `id`         | INTEGER               | PRIMARY KEY              |
| `name`       | TEXT                  | UNIQUE NOT NULL          |
| `sort_order` | INTEGER               | DEFAULT 0               |
| `created_at` | DATETIME              |                          |

### `tags`

Individual tag definitions. Each tag optionally belongs to a group.

| Column       | Type    | Constraints                              |
| ------------ | ------- | ---------------------------------------- |
| `id`         | INTEGER | PRIMARY KEY                              |
| `name`       | TEXT    | NOT NULL                                 |
| `color`      | TEXT    | Hex color code (e.g., `#e74c3c`)        |
| `group_id`   | INTEGER | REFERENCES `tag_groups(id)`, nullable for ungrouped tags |
| `created_by` | TEXT    | Clerk user ID                            |
| `created_at` | DATETIME |                                         |

**Unique constraint:** `UNIQUE(name, group_id)` -- the same tag name can exist in different groups, but not duplicated within a single group.

### `photo_tags`

Junction table associating photos with tags.

| Column       | Type    | Constraints                    |
| ------------ | ------- | ------------------------------ |
| `photo_id`   | INTEGER | REFERENCES `photos(id)`       |
| `tag_id`     | INTEGER | REFERENCES `tags(id)`         |
| `created_by` | TEXT    | Clerk user ID                  |
| `created_at` | DATETIME |                               |

**Primary key:** `PRIMARY KEY(photo_id, tag_id)` -- a photo can only have each tag applied once.

---

## Tag Groups

Tag groups are organizational containers that cluster related tags together. They have no semantic effect on filtering or search -- they exist purely to keep the tag list manageable as it grows.

### Examples

| Group      | Tags                              |
| ---------- | --------------------------------- |
| People     | Alice, Bob, Charlie               |
| Places     | Beach, Mountains, Home, Park      |
| Events     | Birthday 2024, Vacation, Wedding  |
| Seasons    | Spring, Summer, Fall, Winter      |

### Administration

Tag groups are managed by admins only:

- **Create** a new group with a name and optional sort order.
- **Rename** an existing group.
- **Reorder** groups by updating `sort_order` values.
- **Delete** a group. Tags within the deleted group become ungrouped (they are not deleted).

---

## UI Interactions

### Photo Detail View

The photo detail view includes a tag sidebar showing:

- All tags currently applied to the photo, displayed as colored chips grouped by tag group.
- An autocomplete input field to add new tags. Typing filters existing tags; pressing Enter on a match applies it. If the typed name does not match an existing tag and the user has admin privileges, they are offered the option to create a new tag inline.
- A remove button (x) on each tag chip to remove the association.

### Grid View (Bulk Tagging)

- Select multiple photos using checkboxes or shift-click range selection.
- A bulk action toolbar appears with "Add Tags" and "Remove Tags" buttons.
- **Bulk-apply**: opens a tag picker; selected tags are added to all selected photos.
- **Bulk-remove**: opens a tag picker showing only tags present on at least one selected photo; selected tags are removed from all selected photos.

### Tag Browser

A sidebar panel or dedicated page that displays:

- All tag groups in sort order, each expandable to show its tags.
- Ungrouped tags in a separate section at the bottom.
- Photo count next to each tag (e.g., "Beach (42)").
- Clicking a tag applies it as a filter to the main photo grid.

### Tag Management Page (Admin Only)

An admin-only settings page for full tag lifecycle management:

- Create new tags with a name, optional color, and optional group assignment.
- Rename existing tags.
- Change a tag's color.
- Move tags between groups (or make them ungrouped).
- Delete tags. Deletion cascades to `photo_tags` -- the tag is removed from all photos.
- Create, rename, reorder, and delete tag groups.

---

## Filtering

### Single Tag Filter

Click any tag (in the tag browser, on a photo, or in search results) to filter the timeline/grid view to show only photos that have that tag applied.

### Multi-Tag Filter (AND / OR)

Users can build compound tag filters by selecting multiple tags:

- **AND mode**: Photos must have ALL selected tags. Example: "People: Alice" AND "Places: Beach" returns only photos tagged with both Alice and Beach.
- **OR mode**: Photos must have AT LEAST ONE of the selected tags. Example: "People: Alice" OR "People: Bob" returns photos tagged with either person.

A toggle in the filter UI switches between AND and OR mode. The current mode is clearly indicated.

### Combining with Date Range

Tag filters can be combined with date range filtering. The filters are composed with AND logic:

- "People: Alice" AND date range 2024-06-01 to 2024-06-30 returns photos of Alice taken in June 2024.

---

## API Endpoints

### Tags

| Method | Endpoint                          | Auth    | Description                          |
| ------ | --------------------------------- | ------- | ------------------------------------ |
| GET    | `/api/v1/tags`                    | User    | List all tags, grouped by tag group  |
| POST   | `/api/v1/tags`                    | Admin   | Create a new tag                     |
| PUT    | `/api/v1/tags/:id`                | Admin   | Update a tag (name, color, group)    |
| DELETE | `/api/v1/tags/:id`                | Admin   | Delete a tag (cascades to photo_tags)|

**POST /api/v1/tags** request body:

```json
{
  "name": "Beach",
  "color": "#3498db",
  "group_id": 2
}
```

Both `color` and `group_id` are optional. Omitting `group_id` creates an ungrouped tag.

### Tag Groups

| Method | Endpoint                          | Auth    | Description                                  |
| ------ | --------------------------------- | ------- | -------------------------------------------- |
| GET    | `/api/v1/tag-groups`              | User    | List all tag groups                          |
| POST   | `/api/v1/tag-groups`              | Admin   | Create a new tag group                       |
| PUT    | `/api/v1/tag-groups/:id`          | Admin   | Update a tag group (name, sort_order)        |
| DELETE | `/api/v1/tag-groups/:id`          | Admin   | Delete a group (tags become ungrouped)       |

### Photo-Tag Associations

| Method | Endpoint                              | Auth | Description                        |
| ------ | ------------------------------------- | ---- | ---------------------------------- |
| POST   | `/api/v1/photos/:id/tags`             | User | Assign tags to a photo             |
| DELETE | `/api/v1/photos/:id/tags/:tagId`      | User | Remove a tag from a photo          |
| POST   | `/api/v1/photos/bulk-tag`             | User | Bulk assign tags to multiple photos|

**POST /api/v1/photos/:id/tags** request body:

```json
{
  "tag_ids": [1, 2, 3]
}
```

**POST /api/v1/photos/bulk-tag** request body:

```json
{
  "photo_ids": [101, 102, 103, 104],
  "tag_ids": [1, 5]
}
```

---

## Lightroom Keyword Import

When Lightroom Classic integration is enabled (see [07-feature-lightroom-integration.md](./07-feature-lightroom-integration.md)), imported Lightroom keywords are mapped to HomePhotos tags under a dedicated **"Lightroom Keywords"** tag group.

Key behaviors:

- Each Lightroom keyword becomes a tag in the "Lightroom Keywords" group.
- Keyword-to-photo associations from the Lightroom catalog are replicated as `photo_tags` entries.
- These tags are marked with `source: lightroom` metadata so the UI can display them with a distinct visual indicator (e.g., a Lightroom icon or "Imported from Lightroom" tooltip).
- **Lightroom-sourced tags are read-only in the UI.** Users cannot manually add or remove Lightroom keyword tags from photos, since these tags reflect the state of the Lightroom catalog. They are updated only when a new catalog copy is imported.
- Users can still apply their own (non-Lightroom) tags to the same photos. The two tag sources coexist without conflict.
