# Feature: Caching & Thumbnail Generation

## Overview

Caching and thumbnail generation is the core value proposition of HomePhotos. A photo library of RAW files (50MB+ each) cannot be browsed directly over the network with any semblance of speed. HomePhotos runs a background processing pipeline that generates optimized image variants -- thumbnails for grid views, previews for detail views -- so that browsing feels fast even with tens of thousands of large RAW files.

Source files are never modified. All generated variants are written to a separate cache directory.

## Cache Architecture

### Background Worker Pool

Thumbnail generation runs as a background worker pool with bounded concurrency:

- **Configurable concurrency** -- the number of concurrent workers is configurable (default: 4). This controls how many images are processed simultaneously.
- **Job queue** -- the queue is the SQLite `photos` table itself. Any photo with `cache_status = 'pending'` is a job waiting to be processed.
- **Worker loop** -- each worker picks up the next pending job, generates thumbnails and previews, and updates the database record to `'cached'` or `'error'`.

### Processing Priority

Workers process pending photos in order of `captured_at` descending (most recently captured first). This prioritization ensures that users see their newest photos quickly, even while the rest of the library is still being processed.

```sql
SELECT id, file_path FROM photos
WHERE cache_status = 'pending'
ORDER BY captured_at DESC
LIMIT 1;
```

## Thumbnail Pipeline

The following steps are executed for each photo, in order:

### Step 1: Read Source File

Read the source file from the SMB mount. The mount is read-only; no writes are made to the source directory.

### Step 2: Extract or Decode Image Data

The approach depends on the file format:

- **RAW formats (ARW, DNG):** Extract the embedded JPEG preview using LibRaw's `unpack_thumb()` function. Sony A7RV (and most modern cameras) embed a full-resolution JPEG preview (~10MB) inside every ARW file. Extracting this embedded JPEG is dramatically faster than decoding raw sensor data, and the embedded JPEG already has the camera's default processing applied.
- **JPEG:** Read the file directly. No extraction needed.
- **PNG, TIFF:** Read and decode the file directly.

### Step 3: Extract EXIF Metadata

Parse EXIF metadata from the source file (or extracted JPEG). Extracted fields include:

- Camera model
- Lens model
- Focal length, aperture, shutter speed, ISO
- Capture date/time (`DateTimeOriginal`)
- Image orientation
- Image dimensions

### Step 4: Apply EXIF Orientation Correction

Apply the EXIF orientation tag to produce a correctly oriented image. EXIF orientation values (1-8) encode rotations and flips that the camera recorded based on its physical orientation at capture time. The generated thumbnails and previews must have this correction baked in so they display correctly without relying on the client to interpret EXIF orientation.

### Step 5: Generate Thumbnail

Resize the image to **300px on the longest edge** and encode as JPEG at **quality 80**. This produces small files (~20-40KB) suitable for grid views.

### Step 6: Generate Preview

Resize the image to **1600px on the longest edge** and encode as JPEG at **quality 85**. This produces medium-sized files (~200-400KB) suitable for the detail view.

### Step 7: Store Full-Resolution JPEG (Optional)

For RAW files, the extracted full-resolution embedded JPEG can optionally be saved to the cache. This avoids re-reading and re-extracting from the ARW file on every full-resolution request. See the disk usage section below for the storage trade-off.

### Step 8: Write to Cache Directory

Write all generated variants to the cache directory using the layout described below.

### Step 9: Update Database

Update the photo's record in SQLite:

- Set `cache_status = 'cached'`
- Store extracted EXIF metadata (camera, lens, settings, dimensions, capture date)
- Update `scanned_at` timestamp

## Cache Directory Layout

Cache files are organized using the first two characters of the photo's fingerprint as a prefix directory. This prevents any single directory from accumulating too many entries, which can degrade filesystem performance.

```
{cache_root}/
  {first-2-chars-of-fingerprint}/
    {fingerprint}/
      thumb.jpg
      preview.jpg
      full.jpg       (optional)
```

### Example

```
cache/
  a3/
    a3f7b2c9e1d4506f8a2b.../
      thumb.jpg
      preview.jpg
      full.jpg
  7e/
    7e91c4d8f2a3b5e0c1d6.../
      thumb.jpg
      preview.jpg
```

The two-character prefix creates up to 256 subdirectories (00-ff) under the cache root, distributing files evenly.

## Cache Invalidation

### Source File Changed

If a subsequent scan detects that a source file's mtime or size has changed (producing a different fingerprint), the photo's `cache_status` is set back to `'pending'`. The background worker will regenerate the cached variants on its next pass.

### Source File Deleted

If a source file is no longer found during a scan, the photo record is marked as orphaned. The cached files are **not** automatically deleted. This is a deliberate safety measure -- accidental SMB mount failures should not trigger mass cache deletion. An admin can review orphaned entries and clean up via the admin API.

### Manual Cache Purge

Admins can trigger a cache purge via the admin API endpoint (`POST /api/v1/admin/cache/purge`). This allows clearing cached files for specific photos or the entire cache, forcing regeneration.

## On-Demand Generation

If a user requests an image (thumbnail, preview, or full) that has not yet been cached, the server generates it synchronously in the request handler and caches the result before returning the response.

This serves two purposes:

1. **Immediate usability** -- the app is usable before the full background scan and cache generation completes. Users do not have to wait for the entire library to be processed.
2. **Prioritization** -- on-demand requests effectively have higher priority than the background queue because they are served immediately.

The on-demand path runs the same pipeline steps as the background worker (extract, decode, resize, cache), but does so inline in the HTTP request. The response may take 1-2 seconds for a single image, but subsequent requests for the same image will be served from cache.

## Disk Usage Estimates

The following estimates are based on a representative library of 20,000 photos, each approximately 50MB ARW files, totaling ~1TB of source data.

| Tier | Size Per Image | Total (20K photos) | Notes |
|---|---|---|---|
| **Thumbnails** (300px) | ~30KB | ~600MB | Always generated. Essential for browsing. |
| **Previews** (1600px) | ~300KB | ~6GB | Always generated. Essential for detail view. |
| **Full-res extracted JPEGs** | ~10MB | ~200GB | Optional. Can be served on-demand instead. |

### Recommendation

Skip the full-resolution cache tier initially. Instead, extract the embedded JPEG from the ARW file on-demand when a user requests the full-resolution view. This keeps cache storage at approximately **7GB** for a 20,000-photo library (thumbnails + previews only), which is very manageable.

If on-demand full-res extraction proves too slow for the use case (e.g., slow SMB link), the full-res cache can be enabled later to trade disk space for speed.

## Performance Targets

| Metric | Target | Notes |
|---|---|---|
| **Thumbnail generation rate** | 5-10 images/second | On modest hardware. The bottleneck is disk I/O on the SMB mount, not CPU. Extracting the embedded JPEG from an ARW is fast; the limiting factor is reading 50MB files over the network. |
| **Initial scan of 20,000 photos** | Under 1 hour | Assumes 4 workers and a reasonably fast SMB connection. |
| **On-demand single image generation** | Under 2 seconds | For a single image not yet in cache. Includes reading from SMB, extracting, resizing, and encoding. |
| **Cached image serving** | Under 50ms | Serving a pre-generated thumbnail or preview from local disk. |

## Error Handling

### Corrupt or Unreadable Files

If a source file cannot be read, decoded, or processed (e.g., truncated file, unsupported RAW variant, corrupt EXIF data):

- Log the error with the file path and error details
- Set `cache_status = 'error'` in the database
- Skip the file and continue processing the next job
- **Do not crash the worker.** One bad file must not halt processing of the entire library.

### SMB Mount Unavailable

If the source directory (SMB mount) becomes unavailable:

- Pause all background processing
- Retry periodically (e.g., every 30 seconds)
- Log warnings on each failed retry
- Resume processing automatically when the mount becomes available again
- Serve existing cached images normally (the cache is on local disk, not the SMB mount)

### Disk Full on Cache Volume

If the cache volume runs out of disk space:

- Stop all background processing immediately
- Log an error with the cache directory path and available space
- Report the condition via the health endpoint (`GET /api/v1/health`) so monitoring systems can alert
- Do not attempt to continue writing, which would produce corrupt partial files
- Resume automatically if space is freed
