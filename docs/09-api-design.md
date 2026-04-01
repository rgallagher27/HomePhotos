# API Design

## Overview

HomePhotos exposes a RESTful JSON API under `/api/v1/`. All endpoints require Clerk JWT authentication unless explicitly noted otherwise (the health check endpoint is the sole exception). The API follows standard HTTP semantics: `GET` for retrieval, `POST` for creation and actions, `PUT` for updates, and `DELETE` for removal.

---

## Authentication

Every request (except `GET /api/v1/health`) must include a valid Clerk session JWT. The token can be provided in two ways:

| Method | Format |
|---|---|
| Authorization header | `Authorization: Bearer <clerk_session_jwt>` |
| Clerk session cookie | Automatically set by the Clerk frontend SDK |

### Backend verification flow

1. Backend middleware extracts the JWT from the header or cookie.
2. The token is verified using the Clerk Go SDK or by fetching the JWKS endpoint and validating the signature locally.
3. On success, the middleware extracts the `clerk_user_id` from the token claims.
4. The `clerk_user_id` is looked up in the local `users` table to resolve the user's role (`admin` or `viewer`).
5. The authenticated user context is attached to the request for downstream handlers.

### Error responses

| Scenario | Status | Description |
|---|---|---|
| Missing or invalid token | `401 Unauthorized` | The JWT is absent, expired, or fails signature verification. |
| Insufficient role | `403 Forbidden` | The authenticated user does not have the required role for the endpoint (e.g., a `viewer` calling an admin-only endpoint). |

---

## Error Format

All error responses use a consistent JSON envelope:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Photo not found"
  }
}
```

The `code` field is a machine-readable uppercase string. The `message` field is a human-readable description suitable for displaying in client UI.

### Standard HTTP status codes

| Status | Code | Usage |
|---|---|---|
| 400 | `BAD_REQUEST` | Malformed input, invalid query parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 403 | `FORBIDDEN` | Authenticated but insufficient permissions |
| 404 | `NOT_FOUND` | Resource does not exist |
| 409 | `CONFLICT` | Duplicate resource (e.g., tag name already exists) |
| 422 | `VALIDATION_ERROR` | Input fails validation rules |
| 500 | `INTERNAL_ERROR` | Unexpected server error |

---

## Pagination

The API uses **cursor-based pagination** for all list endpoints. This approach handles large datasets efficiently and avoids the offset drift problems of page-number pagination.

### Query parameters

| Parameter | Type | Default | Description |
|---|---|---|---|
| `cursor` | string | _(none)_ | Opaque cursor from a previous response's `next_cursor` field. Omit for the first page. |
| `limit` | integer | `50` | Number of items to return. Minimum 1, maximum 200. |

### Response envelope

```json
{
  "data": [ ... ],
  "next_cursor": "eyJpZCI6MTUwfQ" | null,
  "has_more": true
}
```

- `data` contains the list of items for the current page.
- `next_cursor` is an opaque string to pass as the `cursor` parameter for the next page, or `null` if there are no more results.
- `has_more` is `true` if additional results exist beyond this page, `false` otherwise.

---

## Filtering

List endpoints (primarily `GET /api/v1/photos`) support the following filter parameters:

| Parameter | Type | Example | Description |
|---|---|---|---|
| `tags` | comma-separated IDs | `?tags=1,2,3` | Filter photos by tag IDs. Behavior depends on `tag_mode`. |
| `tag_mode` | string | `?tag_mode=and` | `and` (default): photo must have **all** listed tags. `or`: photo must have **any** of the listed tags. |
| `date_from` | ISO 8601 date | `?date_from=2024-01-01` | Include photos captured on or after this date. |
| `date_to` | ISO 8601 date | `?date_to=2024-12-31` | Include photos captured on or before this date. |
| `folder` | string | `?folder=/Trips/Italy` | Filter to photos within a specific folder path. |
| `format` | comma-separated strings | `?format=arw,jpg` | Filter by file format (case-insensitive). |

Filters are combined with AND logic: specifying both `tags` and `date_from` returns only photos matching both constraints.

---

## Sorting

| Parameter | Type | Default | Description |
|---|---|---|---|
| `sort` | string | `captured_at` | Field to sort by. Accepted values: `captured_at`, `file_name`, `scanned_at`. |
| `order` | string | `desc` | Sort direction. `desc` for newest/largest first, `asc` for oldest/smallest first. |

---

## Endpoints

### Auth

#### `GET /api/v1/auth/me`

Returns the currently authenticated user's profile information.

**Response** `200 OK`

```json
{
  "id": 1,
  "clerk_user_id": "user_2xAbCdEfGhIjKl",
  "role": "admin",
  "display_name": "Jane Doe",
  "email": "jane@example.com"
}
```

---

### Users (admin only)

#### `GET /api/v1/users`

List all registered users.

**Response** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "clerk_user_id": "user_2xAbCdEfGhIjKl",
      "role": "admin",
      "display_name": "Jane Doe",
      "email": "jane@example.com"
    },
    {
      "id": 2,
      "clerk_user_id": "user_3yMnOpQrStUvWx",
      "role": "viewer",
      "display_name": "John Doe",
      "email": "john@example.com"
    }
  ]
}
```

#### `PUT /api/v1/users/:id/role`

Update a user's role.

**Request body**

```json
{
  "role": "admin"
}
```

Valid values: `"admin"`, `"viewer"`.

**Response** `200 OK`

```json
{
  "id": 2,
  "clerk_user_id": "user_3yMnOpQrStUvWx",
  "role": "admin",
  "display_name": "John Doe",
  "email": "john@example.com"
}
```

---

### Photos

#### `GET /api/v1/photos`

List photos with pagination, filtering, and sorting.

**Query parameters**: See [Pagination](#pagination), [Filtering](#filtering), and [Sorting](#sorting) sections above.

**Example request**

```
GET /api/v1/photos?tags=1,3&tag_mode=and&date_from=2024-06-01&sort=captured_at&order=desc&limit=50
```

**Response** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "file_name": "DSC00123.ARW",
      "captured_at": "2024-06-15T14:30:00Z",
      "camera_model": "ILCE-7RM5",
      "tags": [
        { "id": 1, "name": "Vacation" }
      ],
      "cache_status": "cached",
      "thumb_url": "/api/v1/photos/1/image?size=thumb"
    },
    {
      "id": 2,
      "file_name": "DSC00124.ARW",
      "captured_at": "2024-06-15T14:31:12Z",
      "camera_model": "ILCE-7RM5",
      "tags": [
        { "id": 1, "name": "Vacation" },
        { "id": 3, "name": "Landscape" }
      ],
      "cache_status": "cached",
      "thumb_url": "/api/v1/photos/2/image?size=thumb"
    }
  ],
  "next_cursor": "eyJpZCI6Mn0",
  "has_more": true
}
```

#### `GET /api/v1/photos/:id`

Retrieve full detail for a single photo, including all EXIF metadata, tags, and Lightroom data if available.

**Response** `200 OK`

```json
{
  "id": 1,
  "file_name": "DSC00123.ARW",
  "file_path": "/Trips/Italy/DSC00123.ARW",
  "file_size_bytes": 62914560,
  "format": "ARW",
  "width": 9504,
  "height": 6336,
  "captured_at": "2024-06-15T14:30:00Z",
  "scanned_at": "2024-07-01T10:00:00Z",
  "camera_make": "Sony",
  "camera_model": "ILCE-7RM5",
  "lens_model": "FE 24-70mm F2.8 GM II",
  "focal_length_mm": 35,
  "aperture": 5.6,
  "shutter_speed": "1/250",
  "iso": 200,
  "gps_latitude": 43.7696,
  "gps_longitude": 11.2558,
  "tags": [
    { "id": 1, "name": "Vacation" },
    { "id": 3, "name": "Landscape" }
  ],
  "cache_status": "cached",
  "thumb_url": "/api/v1/photos/1/image?size=thumb",
  "preview_url": "/api/v1/photos/1/image?size=preview",
  "full_url": "/api/v1/photos/1/image?size=full",
  "lightroom": {
    "rating": 4,
    "pick_status": "flagged",
    "color_label": "red",
    "develop_settings": true,
    "keywords": ["Italy", "Florence", "Architecture"]
  }
}
```

The `lightroom` field is `null` if no Lightroom catalog has been synced or the photo has no Lightroom metadata.

#### `GET /api/v1/photos/:id/image?size=thumb|preview|full`

Serve the image file at the requested size.

**Query parameters**

| Parameter | Required | Values | Description |
|---|---|---|---|
| `size` | Yes | `thumb`, `preview`, `full` | `thumb`: 400px wide JPEG. `preview`: 1600px wide JPEG. `full`: maximum resolution JPEG rendered from the source file. |

**Response headers**

```
Content-Type: image/jpeg
Cache-Control: public, max-age=31536000, immutable
```

The aggressive caching is safe because image content is immutable for a given photo ID and size.

**Error cases**

- `404 Not Found` if the photo does not exist or the requested size has not been cached and on-demand generation is disabled.
- If on-demand generation is enabled, the server renders the requested size from the source file before responding.

---

### Tags

#### `GET /api/v1/tags`

List all tags, grouped by tag group.

**Response** `200 OK`

```json
{
  "data": [
    {
      "group": { "id": 1, "name": "People" },
      "tags": [
        { "id": 1, "name": "Alice", "color": "#FF5733", "photo_count": 42 },
        { "id": 2, "name": "Bob", "color": "#3357FF", "photo_count": 28 }
      ]
    },
    {
      "group": { "id": 2, "name": "Places" },
      "tags": [
        { "id": 3, "name": "Beach", "color": "#33FF57", "photo_count": 15 }
      ]
    },
    {
      "group": null,
      "tags": [
        { "id": 4, "name": "Favorite", "color": "#FFD700", "photo_count": 100 }
      ]
    }
  ]
}
```

Tags without an assigned group appear under a `null` group entry.

#### `POST /api/v1/tags` (admin only)

Create a new tag.

**Request body**

```json
{
  "name": "Alice",
  "color": "#FF5733",
  "group_id": 1
}
```

The `group_id` field is optional. If omitted, the tag is ungrouped.

**Response** `201 Created`

```json
{
  "id": 5,
  "name": "Alice",
  "color": "#FF5733",
  "group_id": 1,
  "photo_count": 0
}
```

#### `PUT /api/v1/tags/:id` (admin only)

Update an existing tag.

**Request body**

```json
{
  "name": "Alice Smith",
  "color": "#FF5733",
  "group_id": 1
}
```

All fields are optional; only provided fields are updated.

**Response** `200 OK`

Returns the updated tag object.

#### `DELETE /api/v1/tags/:id` (admin only)

Delete a tag. This cascades removal from all photos that have the tag assigned.

**Response** `204 No Content`

---

### Tag Groups

#### `GET /api/v1/tag-groups`

List all tag groups, ordered by `sort_order`.

**Response** `200 OK`

```json
{
  "data": [
    { "id": 1, "name": "People", "sort_order": 1 },
    { "id": 2, "name": "Places", "sort_order": 2 },
    { "id": 3, "name": "Events", "sort_order": 3 }
  ]
}
```

#### `POST /api/v1/tag-groups` (admin only)

Create a new tag group.

**Request body**

```json
{
  "name": "People",
  "sort_order": 1
}
```

**Response** `201 Created`

```json
{
  "id": 4,
  "name": "People",
  "sort_order": 1
}
```

#### `PUT /api/v1/tag-groups/:id` (admin only)

Update an existing tag group.

**Request body**

```json
{
  "name": "Family",
  "sort_order": 1
}
```

**Response** `200 OK`

Returns the updated tag group object.

#### `DELETE /api/v1/tag-groups/:id` (admin only)

Delete a tag group. Tags that belonged to the deleted group become ungrouped (their `group_id` is set to `NULL`). The tags themselves are not deleted.

**Response** `204 No Content`

---

### Photo Tagging

#### `POST /api/v1/photos/:id/tags`

Assign one or more tags to a photo. If a tag is already assigned, it is silently ignored (idempotent).

**Request body**

```json
{
  "tag_ids": [1, 2, 3]
}
```

**Response** `200 OK`

```json
{
  "photo_id": 1,
  "tags": [
    { "id": 1, "name": "Vacation" },
    { "id": 2, "name": "Alice" },
    { "id": 3, "name": "Landscape" }
  ]
}
```

#### `DELETE /api/v1/photos/:id/tags/:tagId`

Remove a single tag from a photo.

**Response** `204 No Content`

#### `POST /api/v1/photos/bulk-tag`

Assign tags to multiple photos in a single request. Useful for batch tagging in the UI.

**Request body**

```json
{
  "photo_ids": [1, 2, 3],
  "tag_ids": [4, 5]
}
```

**Response** `200 OK`

```json
{
  "updated_count": 3
}
```

---

### Scanner (admin only)

#### `POST /api/v1/scanner/run`

Trigger a filesystem scan. If a scan is already in progress, returns `409 Conflict`.

**Response** `202 Accepted`

```json
{
  "status": "started",
  "job_id": "scan_20240701_100000"
}
```

#### `GET /api/v1/scanner/status`

Get the current scanner status.

**Response** `200 OK`

```json
{
  "status": "scanning",
  "total_files": 20000,
  "processed": 15000,
  "errors": 3,
  "started_at": "2024-07-01T10:00:00Z"
}
```

When idle:

```json
{
  "status": "idle",
  "total_files": 0,
  "processed": 0,
  "errors": 0,
  "started_at": null
}
```

---

### System

#### `GET /api/v1/health`

Health check endpoint. **No authentication required.**

**Response** `200 OK`

```json
{
  "status": "ok",
  "smb_mount": "connected",
  "database": "ok",
  "cache_dir": "writable",
  "scan_status": "idle"
}
```

If any subsystem is unhealthy, the top-level `status` changes to `"degraded"` and the affected field indicates the problem.

#### `GET /api/v1/system/stats` (admin only)

System statistics for the admin dashboard.

**Response** `200 OK`

```json
{
  "total_photos": 20000,
  "cached_photos": 19500,
  "pending_cache": 500,
  "cache_size_bytes": 7000000000,
  "total_tags": 45
}
```

---

### Lightroom (admin only, stretch goal)

These endpoints are part of the optional Lightroom integration for importing metadata from an Adobe Lightroom Classic catalog.

#### `POST /api/v1/lightroom/sync`

Trigger a Lightroom catalog sync. Reads the `.lrcat` SQLite database and matches photos by file path.

**Response** `202 Accepted`

```json
{
  "status": "started"
}
```

#### `GET /api/v1/lightroom/status`

Get the current Lightroom sync status.

**Response** `200 OK`

```json
{
  "last_sync": "2024-07-01T12:00:00Z",
  "photos_matched": 18000,
  "keywords_imported": 250
}
```

If no sync has ever been performed:

```json
{
  "last_sync": null,
  "photos_matched": 0,
  "keywords_imported": 0
}
```
