# Project Overview

## Problem Statement

Photos are captured on a Sony A7RV and imported into Adobe Lightroom Classic on a Mac. The RAW files (Sony ARW format) average around 50 MB each. With a growing library, syncing these files to other devices (iPads, phones, a partner's laptop) is impractical -- cloud sync services choke on the volume, and directly browsing an SMB share from a Mac or iPad delivers a poor experience with files this large. Thumbnails are slow to generate on the client, scrubbing through folders is painful, and there is no way to organize or filter without Lightroom itself.

The current workflow looks like this:

1. Shoot on the A7RV.
2. Import into Lightroom Classic, which stores originals on a TrueNAS SMB share.
3. Want to browse and share photos from other devices -- but there is no good way to do it.

## Project Goals

**HomePhotos** is a self-hosted web application that serves optimized versions of photos stored on a TrueNAS SMB share. It is designed for a household of 1-5 users who want to browse, search, and view a personal photo library without needing Lightroom or direct filesystem access.

Core goals:

- **Serve optimized images.** Extract embedded JPEGs from RAW files and generate thumbnails and previews so that browsing is fast on any device.
- **Never modify originals.** The SMB share is mounted read-only. The application treats the photo library as an immutable source of truth.
- **Multi-user with scoped access.** An admin (the photographer) manages the library. Other users (family, partner) can browse, favorite, and filter photos based on permissions.
- **Self-hosted and private.** Runs on the home network (on TrueNAS itself or a separate machine). Remote access is handled via Tailscale -- no photos are exposed to the public internet.

## User Personas

### 1. Admin / Photographer

- Imports photos via Lightroom Classic (outside of HomePhotos).
- Manages the photo library: triggers rescans when new photos are added, manages tags, creates albums.
- Manages users and their access scopes.
- Has full access to all photos and all administrative functions.

### 2. Viewer / Partner

- Browses the photo library through the web UI.
- Views photos filtered by tags, date ranges, or albums.
- Can mark favorites (stored per-user, does not affect the library).
- Does not import, edit, or delete photos.

## Scope

### In Scope for v1

- **Web UI** for browsing photos (grid view, detail view, lightbox).
- **Photo browsing** with filtering by date, tags, and albums.
- **Caching and thumbnail generation.** Embedded JPEG extraction from ARW files, thumbnail (300px) and preview (1600px) generation, stored in a local cache directory.
- **Tagging system.** Admin can tag photos; all users can filter by tags.
- **Multi-user authentication** via Clerk (managed auth service). Role-based access: admin and viewer roles.
- **File scanning.** Watches the SMB mount for new or changed files and queues them for processing.
- **EXIF metadata extraction** and display (camera, lens, exposure, date, GPS if present).
- **Responsive design.** Usable on desktop browsers, tablets, and phones over the local network or Tailscale.

### Explicitly Out of Scope for v1

- **No native mobile app.** The web UI is mobile-friendly, but there is no dedicated iOS or Android app.
- **No photo upload or import through the app.** Photos are added to the library exclusively through Lightroom Classic and the filesystem. HomePhotos is read-only.
- **No photo editing.** No cropping, rotating, filters, or adjustments. HomePhotos is a viewer, not an editor.
- **No video support.** Only still image formats (ARW, JPEG, PNG, TIFF) are supported.

### Stretch Goals

- **Lightroom Classic catalog integration.** Lightroom stores its catalog in an `.lrcat` file, which is a SQLite database. A stretch goal is to read this catalog to import Lightroom keywords, collections, star ratings, and pick/reject flags into HomePhotos, providing a richer browsing experience without manual re-tagging.

## Glossary

| Term | Definition |
|------|------------|
| **ARW** | Sony Alpha RAW. The RAW image format produced by Sony cameras, including the A7RV. Files are typically 40-60 MB and contain full sensor data plus an embedded full-resolution JPEG. |
| **lrcat** | Lightroom Catalog file. The database file used by Adobe Lightroom Classic to store metadata, edits, collections, and keywords. It is a SQLite database and can be read programmatically. |
| **SMB** | Server Message Block. A network file sharing protocol used by TrueNAS (and Windows/macOS/Linux) to share directories over a local network. |
| **Tailscale** | A mesh VPN built on WireGuard. Creates a private network between devices without exposing services to the public internet. Used here for remote access to HomePhotos from outside the home LAN. |
| **EXIF** | Exchangeable Image File Format. A metadata standard embedded in image files containing camera settings (aperture, shutter speed, ISO), date/time, GPS coordinates, lens information, and more. |
| **LibRaw** | An open-source library for reading and processing RAW image data from digital cameras. Used in HomePhotos to extract embedded JPEGs from ARW files without performing a full RAW decode. |
