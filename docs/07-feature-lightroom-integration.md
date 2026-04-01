# Feature: Lightroom Classic Integration (Stretch Goal)

## Overview

Adobe Lightroom Classic stores its entire catalog in a `.lrcat` file, which is a SQLite database. HomePhotos can read a copy of this file to extract metadata, keywords, develop settings, and file mappings -- enriching the photo browsing experience with data that would otherwise be locked inside Lightroom.

This feature is a **stretch goal** and is not required for v1. The core photo browsing, thumbnail generation, and tagging features must work independently of any Lightroom integration.

---

## Critical Safety Rule

> **NEVER open the live `.lrcat` file directly.**

Lightroom Classic locks its catalog database while running. Opening the live file from another process risks SQLite lock contention and can **corrupt the catalog**, potentially destroying years of editing work.

**Always work from a copy of the `.lrcat` file.** All sync strategies described below produce a copy that HomePhotos reads independently. The live catalog on the user's laptop is never touched by HomePhotos.

---

## Catalog Location

The Lightroom catalog lives on the user's laptop (e.g., `~/Pictures/Lightroom/MyPhotos.lrcat`), not on the NAS. This means a mechanism is needed to transfer the catalog data to a location accessible by the HomePhotos server.

The photos themselves are already on the NAS (that is the whole point of HomePhotos), but the catalog metadata exists only on the laptop where Lightroom runs.

---

## Sync Strategies

### 1. Scheduled rsync/copy (Recommended)

A cron job or script on the laptop copies the `.lrcat` file to a known path accessible to HomePhotos -- either a dedicated folder on the NAS via SMB, or directly to the server via rsync over Tailscale.

HomePhotos watches this path for changes by checking the file's modification time (`mtime`).

**Example:**

```bash
# Cron job on the laptop (runs nightly at 2 AM)
# Copies the catalog to the NAS over Tailscale
rsync -avz ~/Pictures/Lightroom/MyPhotos.lrcat nas:/mnt/photos/.lightroom/catalog.lrcat
```

**Pros:** Simple, reliable, uses standard tools (rsync, cron), no custom code on the laptop side.

**Cons:** Requires the user to set up a cron job. The catalog copy is only as fresh as the last sync.

### 2. Manual Upload

The admin uploads the catalog file via the HomePhotos web UI. A dedicated upload endpoint accepts the `.lrcat` file and stores it on the server.

**Pros:** Zero setup on the laptop. Works from any browser.

**Cons:** Requires manual action each time the user wants to sync. The `.lrcat` file can be large (hundreds of MB for large catalogs), making uploads slow.

### 3. Lightroom Export Plugin

A custom Lua plugin for Lightroom Classic that exports catalog metadata (keywords, develop settings, file mappings) as JSON and POSTs it to a HomePhotos API endpoint.

**Pros:** Most integrated experience. Could trigger automatically on catalog changes. Sends only metadata (not the full SQLite file), so transfers are smaller.

**Cons:** Most complex to build. Requires the user to install a Lightroom plugin. Lightroom's plugin API has limitations and quirks. Ongoing maintenance burden as Lightroom updates.

### 4. Delta Sync (Optimization)

On each catalog import, diff against the previous import rather than re-importing everything:

- Compare `Adobe_images.id_local` and modification timestamps to detect new and changed entries.
- Only process entries that are new or modified since the last sync.
- Track the last-imported state (e.g., max `id_local` and timestamp) in the HomePhotos database.

This is an **optimization that should be used regardless of which delivery mechanism is chosen** (options 1-3 above). It reduces import time from minutes to seconds for incremental updates on large catalogs.

---

## Key Lightroom Catalog Tables

The `.lrcat` file is a SQLite database. The following tables contain the data HomePhotos needs:

### Core Image and File Tables

| Table                    | Purpose                                | Key Columns                                           |
| ------------------------ | -------------------------------------- | ----------------------------------------------------- |
| `Adobe_images`           | Core image records                     | `id_local`, `rootFile`, `orientation`, `captureTime`  |
| `AgLibraryFile`          | File records                           | `id_local`, `baseName`, `extension`, `idx_filename`   |
| `AgLibraryFolder`        | Folder paths (relative)                | `id_local`, `pathFromRoot`, `rootFolder`              |
| `AgLibraryRootFolder`    | Root folder absolute paths             | `id_local`, `absolutePath`, `name`                    |

### Keyword Tables

| Table                    | Purpose                                | Key Columns                                           |
| ------------------------ | -------------------------------------- | ----------------------------------------------------- |
| `AgLibraryKeyword`       | Keyword definitions                    | `id_local`, `name`, `lc_name`                         |
| `AgLibraryKeywordImage`  | Keyword-to-image associations          | `image`, `tag` (references keyword `id_local`)        |

### Develop Settings Tables

| Table                                    | Purpose                          | Key Columns                                |
| ---------------------------------------- | -------------------------------- | ------------------------------------------ |
| `Adobe_imageDevelopSettings`             | Current develop settings         | `image`, `text`, `whiteBalance`, `orientation` |
| `Adobe_libraryImageDevelopHistoryStep`   | Develop history steps            | (historical record of edits)               |

---

## File Correlation

### Path Reconstruction

Lightroom stores file paths across multiple tables. To reconstruct the full filesystem path for an image:

```
Full path = AgLibraryRootFolder.absolutePath
          + AgLibraryFolder.pathFromRoot
          + AgLibraryFile.baseName
          + '.'
          + AgLibraryFile.extension
```

**Example:**

| Table                 | Value                          |
| --------------------- | ------------------------------ |
| `absolutePath`        | `/Users/username/Pictures/`    |
| `pathFromRoot`        | `2024/Summer/`                 |
| `baseName`            | `IMG_4523`                     |
| `extension`           | `CR3`                          |
| **Reconstructed path** | `/Users/username/Pictures/2024/Summer/IMG_4523.CR3` |

### Path Remapping

The reconstructed paths reference the laptop filesystem, which will not match the NAS mount path inside the HomePhotos container. A configurable path remapping bridges this gap.

**Configuration example:**

```
Laptop path prefix:     /Users/username/Pictures/Photos
HomePhotos source path: /source
```

With this mapping, the Lightroom path `/Users/username/Pictures/Photos/2024/Summer/IMG_4523.CR3` becomes `/source/2024/Summer/IMG_4523.CR3`.

### Fallback Matching

When absolute paths differ in unexpected ways (e.g., different root folder naming), fall back to matching by:

1. **Filename + relative path**: Match on `baseName.extension` within the same relative directory structure.
2. **Filename only**: As a last resort, match on filename alone (with collision detection -- skip if multiple files share the same name).

---

## Date Handling

Lightroom uses the **Apple/Core Data epoch** for timestamps: **January 1, 2001 00:00:00 UTC**.

This differs from the Unix epoch (January 1, 1970) by exactly **978,307,200 seconds**.

### Conversion Formula

```
unix_timestamp = lrcat_timestamp + 978307200
```

### Example

```
Lightroom captureTime:  733276800.0
Unix timestamp:         733276800 + 978307200 = 1711584000
ISO 8601:               2024-03-28T00:00:00Z
```

Apply this conversion to **all date fields** read from the catalog, including `captureTime`, modification timestamps, and develop history dates.

---

## Keyword Import

### Process

1. Read all keywords from `AgLibraryKeyword`.
2. Read all keyword-to-image associations from `AgLibraryKeywordImage`.
3. Create (or ensure existence of) a **"Lightroom Keywords"** tag group in HomePhotos.
4. Map each Lightroom keyword to a tag within that group.
5. Associate tags with photos based on the keyword-image relationships and file correlation (see above).

### Hierarchical Keywords

Lightroom supports keyword hierarchies via a `parent_id` field on `AgLibraryKeyword`. For example:

```
Animals
  ├── Dogs
  │   ├── Golden Retriever
  │   └── Labrador
  └── Cats
```

HomePhotos flattens the hierarchy for v1 -- each keyword becomes an individual tag. The hierarchical relationship is preserved in the tag name using a separator (e.g., "Animals > Dogs > Golden Retriever") so the context is not lost.

### Source Tracking

All Lightroom-imported tags are marked with `source: lightroom` metadata. This allows the UI to:

- Display a visual indicator (e.g., a Lightroom icon) on these tags.
- Prevent users from manually adding or removing Lightroom tags from photos (they are **read-only** in the UI).
- Update them only when a new catalog import occurs.

See also: [06-feature-tagging.md](./06-feature-tagging.md) for how Lightroom tags interact with the broader tagging system.

---

## Develop Settings Display

### Parsing

The `Adobe_imageDevelopSettings.text` column contains key-value pairs representing the current develop settings for an image. Parse this text to extract the settings below.

### Displayed Settings

Show the following as **read-only metadata** in the photo detail view:

| Setting Group    | Settings                                    |
| ---------------- | ------------------------------------------- |
| Tone             | Exposure, Contrast, Highlights, Shadows, Whites, Blacks |
| White Balance    | Temperature, Tint                           |
| Presence         | Clarity, Vibrance, Saturation               |
| Profile          | Camera profile name                         |

These values are purely informational -- HomePhotos does not re-render images based on develop settings.

### Rendering Edited Images (Advanced Stretch Goal)

Lightroom's non-destructive edits are stored as processing instructions, not pixel data. In theory, some basic adjustments (exposure, white balance, tone curve) could be applied during thumbnail generation using a compatible processing pipeline (e.g., `dcraw` with adjusted parameters, or `libraw` + custom tone mapping).

**This is extremely complex.** Lightroom's processing pipeline is proprietary and involves:

- Camera-specific color profiles
- Proprietary tone curve algorithms
- Complex masking and local adjustments
- Lens correction profiles

**Recommendation:** Defer this significantly. The informational display of develop settings provides value without the engineering cost of attempting to replicate Lightroom's rendering engine.

---

## Configuration

| Environment Variable            | Description                                              | Default        |
| ------------------------------- | -------------------------------------------------------- | -------------- |
| `HOMEPHOTOS_LRCAT_PATH`        | Path to the Lightroom catalog copy on the server         | *(none)*       |
| `HOMEPHOTOS_LR_PATH_MAP`       | Path remapping from laptop paths to HomePhotos source path. JSON object (e.g., `{"/Users/me/Pictures": "/source"}`) or colon-separated (e.g., `/Users/me/Pictures:/source`) | *(none)*       |
| `HOMEPHOTOS_LR_SYNC_INTERVAL`  | How often to check for catalog updates (e.g., `1h`, `6h`, `24h`) | Manual only    |

When `HOMEPHOTOS_LRCAT_PATH` is not set, the Lightroom integration is completely disabled and no related UI elements are shown.
