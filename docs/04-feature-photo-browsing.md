# Feature: Photo Browsing

## Overview

Photo browsing is the core feature of HomePhotos. The system scans a configured source directory on the TrueNAS SMB mount, indexes photo files and their metadata into SQLite, and serves them through a responsive web UI with multiple browse views and a detail view.

The source directory is mounted read-only. HomePhotos never modifies, moves, or deletes source files.

## File Scanning

### Source Directory

The scanner recursively walks the configured source directory (e.g., `/mnt/photos`) to discover photo files. The source directory is typically a TrueNAS SMB share mounted on the HomePhotos server.

### Supported Formats

The following file extensions are recognized (case-insensitive):

| Extension | Type |
|---|---|
| `.arw` | Sony RAW |
| `.dng` | Adobe Digital Negative (RAW) |
| `.jpg`, `.jpeg` | JPEG |
| `.tif`, `.tiff` | TIFF |
| `.png` | PNG |

Files with other extensions are ignored.

### What the Scanner Records

For each discovered file, the scanner records:

- **File path** -- relative to the source root (e.g., `2025/December/DSC00123.ARW`)
- **File size** -- in bytes
- **Modification time (mtime)** -- from the filesystem
- **Content fingerprint** -- computed from `path + size + mtime` for speed. This is not a content hash (which would require reading every byte of every file on every scan). It is sufficient to detect changes because any edit to a file changes its mtime and/or size.

### Incremental Scanning

On subsequent scan runs, the scanner only processes files where the mtime or file size has changed since the last scan. Files that are unchanged are skipped entirely. This makes repeated scans fast even for large libraries.

Scan state is stored in the SQLite database. Each photo record includes the fingerprint used during the last scan, allowing the scanner to detect changes efficiently.

### Scan Triggers

Scans can be triggered in three ways:

1. **Manual** -- an admin calls the scan API endpoint (`POST /api/v1/admin/scan`)
2. **On server startup** -- the scanner runs automatically when the HomePhotos server starts
3. **Configurable interval** -- a periodic scan runs on a configurable interval (e.g., every 6 hours) to pick up new files added to the source directory

## Photo Database Model

### SQLite Schema

```sql
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
    camera_model    TEXT,
    lens_model      TEXT,
    focal_length    REAL,
    aperture        REAL,
    shutter_speed   TEXT,
    iso             INTEGER,
    orientation     INTEGER,
    fingerprint     TEXT NOT NULL,
    scanned_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cache_status    TEXT NOT NULL DEFAULT 'pending' CHECK (cache_status IN ('pending', 'cached', 'error'))
);

CREATE INDEX idx_photos_captured_at ON photos (captured_at);
CREATE INDEX idx_photos_file_path ON photos (file_path);
CREATE INDEX idx_photos_cache_status ON photos (cache_status);
CREATE INDEX idx_photos_format ON photos (format);
```

| Column | Type | Description |
|---|---|---|
| `id` | `INTEGER PRIMARY KEY` | Auto-incrementing ID |
| `file_path` | `TEXT UNIQUE` | Path relative to source root |
| `file_name` | `TEXT` | Filename component (e.g., `DSC00123.ARW`) |
| `file_size` | `INTEGER` | File size in bytes |
| `file_mtime` | `DATETIME` | Last modification time from filesystem |
| `format` | `TEXT` | Normalized format string: `arw`, `jpg`, `tif`, `png`, `dng` |
| `width` | `INTEGER` | Image width in pixels (from EXIF or decoded image) |
| `height` | `INTEGER` | Image height in pixels |
| `captured_at` | `DATETIME` | Capture date/time from EXIF (`DateTimeOriginal`). Null if EXIF is missing. |
| `camera_model` | `TEXT` | Camera model from EXIF (e.g., `ILCE-7RM5`) |
| `lens_model` | `TEXT` | Lens model from EXIF (e.g., `FE 24-70mm F2.8 GM II`) |
| `focal_length` | `REAL` | Focal length in mm |
| `aperture` | `REAL` | F-number (e.g., `2.8`) |
| `shutter_speed` | `TEXT` | Shutter speed as a string (e.g., `1/250`) |
| `iso` | `INTEGER` | ISO sensitivity |
| `orientation` | `INTEGER` | EXIF orientation tag (1-8) |
| `fingerprint` | `TEXT` | Content fingerprint for change detection |
| `scanned_at` | `DATETIME` | When this record was last scanned/updated |
| `cache_status` | `TEXT` | Thumbnail cache state: `pending`, `cached`, or `error` |

## Browse Views

### Timeline View (Default)

The default view presents photos sorted by EXIF capture date (`captured_at`), newest first. Photos are grouped by day or month depending on density:

- **Day grouping** -- used when zoomed in or when there are relatively few photos per day. Each group shows a date header (e.g., "December 15, 2025") followed by a thumbnail grid.
- **Month grouping** -- used as a higher-level overview. Each group shows a month header (e.g., "December 2025").

Photos without EXIF dates fall back to `file_mtime` for sorting and are flagged in the UI.

### Folder View

The folder view mirrors the source directory structure, allowing users to navigate folders like a file browser:

- The root shows top-level directories (e.g., `2024/`, `2025/`)
- Clicking a folder navigates into it, showing subfolders and photo thumbnails
- Breadcrumb navigation shows the current path and allows jumping to parent folders
- Photos within a folder are sorted by filename

This view is useful when the source directory is organized by date or event (e.g., `2025/December/Christmas/`).

### Tag View

The tag view allows filtering photos by one or more tags. See the tagging feature documentation (doc 06) for details on how tags are managed.

- Select one or more tags from a sidebar or search
- Photos matching the selected tags are displayed in a grid
- Supports intersection (photos matching ALL selected tags) and union (photos matching ANY selected tag)

## Photo Detail View

Clicking a photo thumbnail in any browse view opens the detail view.

### Preview Image

The detail view displays the preview-tier image (1600px longest edge) by default. This provides a good-quality view that loads quickly. An option is available to load the full-resolution extracted JPEG for pixel-level inspection.

### EXIF Metadata Panel

A metadata panel displays technical information extracted from the photo's EXIF data:

| Field | Example |
|---|---|
| Camera | Sony ILCE-7RM5 |
| Lens | FE 24-70mm F2.8 GM II |
| Focal Length | 35mm |
| Aperture | f/2.8 |
| Shutter Speed | 1/250s |
| ISO | 400 |
| Date/Time | 2025-12-15 14:30:22 |
| Dimensions | 9504 x 6336 |

### Tags Panel

The tags panel shows all tags assigned to the photo and provides controls to add or remove tags (for users with appropriate permissions).

### Lightroom Develop Info

If Lightroom develop settings are available for the photo (see doc 07), a panel displays relevant processing information.

## Image Serving

Images are served at three quality tiers to balance load time and visual quality.

### Tiers

| Tier | Max Size | JPEG Quality | Use Case |
|---|---|---|---|
| **Thumbnail** | 300px longest edge | 80 | Grid views (timeline, folder, tag) |
| **Preview** | 1600px longest edge | 85 | Detail view |
| **Full** | Original resolution | Original | Full-resolution inspection |

### Endpoints

All image endpoints require authentication.

#### `GET /api/v1/photos/:id/image?size=thumb`

Returns the thumbnail (300px). Used by all grid/browse views.

#### `GET /api/v1/photos/:id/image?size=preview`

Returns the preview (1600px). Used by the detail view.

#### `GET /api/v1/photos/:id/image?size=full`

Returns the full-resolution image. For RAW files (ARW, DNG), this is the extracted embedded JPEG. For JPEG/PNG/TIFF, this is the original file.

### Cache Headers

Cached image responses include aggressive cache headers since thumbnails and previews are content-addressed (their fingerprint is embedded in the cache path):

```
Cache-Control: public, max-age=31536000, immutable
```

This tells browsers and CDNs to cache the image for one year without revalidation. If the source file changes, a new fingerprint is generated, producing a new URL.

## Pagination

Photo listing endpoints use cursor-based pagination for stable, performant results.

### Parameters

| Parameter | Type | Default | Description |
|---|---|---|---|
| `cursor` | `string` | (none) | Opaque cursor from the previous response. Omit for the first page. |
| `limit` | `integer` | `50` | Number of photos per page. Maximum `200`. |

### Cursor Format

The cursor encodes the last photo's sort key for the current view:

- **Timeline view** -- `captured_at` timestamp + `id` for uniqueness (two photos may share the same capture timestamp)
- **Folder view** -- `file_name` + `id`

Using a composite cursor ensures stable ordering even when multiple photos share the same sort value.

### Response Format

```json
{
  "photos": [ ... ],
  "next_cursor": "2025-12-15T14:30:22Z_4827",
  "has_more": true
}
```

If `has_more` is `false`, there are no more results beyond this page.

## Frontend Considerations

### Virtual Scrolling

For libraries with tens of thousands of photos, rendering all thumbnails in the DOM at once is not feasible. The frontend uses virtual scrolling (also called windowed rendering) to only render the thumbnails currently visible in the viewport, plus a small buffer above and below.

This keeps DOM node count low and maintains smooth scrolling performance regardless of library size.

### Lazy Loading of Thumbnails

Thumbnail images use the browser's native lazy loading (`loading="lazy"`) or an Intersection Observer-based approach to defer loading images until they are near the viewport. This reduces initial page load time and bandwidth consumption.

### Responsive Grid Layout

The thumbnail grid adapts to the available screen width:

- **Desktop (wide)** -- 6-8 columns of thumbnails
- **Tablet** -- 4-5 columns
- **Mobile** -- 2-3 columns

Grid columns and thumbnail sizes are computed dynamically using CSS Grid or a layout calculation that accounts for the viewport width and a target thumbnail size.
