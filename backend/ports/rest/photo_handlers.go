package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rgallagher/homephotos/domain/photo"
)

func (s *Server) GetPhotos(w http.ResponseWriter, r *http.Request, params GetPhotosParams) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	listParams := photo.ListParams{
		SortBy:    "captured_at",
		SortOrder: "desc",
		Limit:     50,
	}

	if params.Cursor != nil {
		listParams.Cursor = *params.Cursor
	}
	if params.Limit != nil {
		listParams.Limit = *params.Limit
	}
	if params.Sort != nil {
		listParams.SortBy = string(*params.Sort)
	}
	if params.Order != nil {
		listParams.SortOrder = string(*params.Order)
	}
	if params.DateFrom != nil {
		t := time.Time(params.DateFrom.Time)
		listParams.DateFrom = &t
	}
	if params.DateTo != nil {
		t := time.Time(params.DateTo.Time)
		listParams.DateTo = &t
	}
	if params.Folder != nil {
		listParams.Folder = *params.Folder
	}
	if params.Format != nil {
		listParams.Format = *params.Format
	}

	result, err := s.photos.List(r.Context(), listParams)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list photos")
		return
	}

	data := make([]PhotoListItem, len(result.Photos))
	for i, p := range result.Photos {
		data[i] = photoToListItem(p)
	}

	resp := PhotoListResponse{
		Data:    data,
		HasMore: result.HasMore,
	}
	if result.NextCursor != "" {
		resp.NextCursor = &result.NextCursor
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) GetPhoto(w http.ResponseWriter, r *http.Request, id int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	p, err := s.photos.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, photo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "photo not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch photo")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(photoToDetail(p))
}

func photoToListItem(p photo.Photo) PhotoListItem {
	item := PhotoListItem{
		Id:          p.ID,
		FileName:    p.FileName,
		Format:      p.Format,
		CacheStatus: PhotoListItemCacheStatus(p.CacheStatus),
		CapturedAt:  p.CapturedAt,
		FilePath:    &p.FilePath,
	}
	if p.CameraModel != "" {
		item.CameraModel = &p.CameraModel
	}
	thumbURL := fmt.Sprintf("/api/v1/photos/%d/image?size=thumb", p.ID)
	item.ThumbUrl = &thumbURL
	return item
}

func photoToDetail(p *photo.Photo) PhotoDetailResponse {
	resp := PhotoDetailResponse{
		Id:            p.ID,
		FileName:      p.FileName,
		FilePath:      p.FilePath,
		FileSizeBytes: p.FileSize,
		Format:        p.Format,
		CapturedAt:    p.CapturedAt,
		ScannedAt:     p.ScannedAt,
		CacheStatus:   PhotoDetailResponseCacheStatus(p.CacheStatus),
		Aperture:      p.Aperture,
		FocalLengthMm: p.FocalLength,
		GpsLatitude:   p.GPSLatitude,
		GpsLongitude:  p.GPSLongitude,
	}

	if p.Width != nil {
		w := int(*p.Width)
		resp.Width = &w
	}
	if p.Height != nil {
		h := int(*p.Height)
		resp.Height = &h
	}
	if p.ISO != nil {
		iso := int(*p.ISO)
		resp.Iso = &iso
	}
	if p.CameraMake != "" {
		resp.CameraMake = &p.CameraMake
	}
	if p.CameraModel != "" {
		resp.CameraModel = &p.CameraModel
	}
	if p.LensModel != "" {
		resp.LensModel = &p.LensModel
	}
	if p.ShutterSpeed != "" {
		resp.ShutterSpeed = &p.ShutterSpeed
	}

	thumbURL := fmt.Sprintf("/api/v1/photos/%d/image?size=thumb", p.ID)
	previewURL := fmt.Sprintf("/api/v1/photos/%d/image?size=preview", p.ID)
	fullURL := fmt.Sprintf("/api/v1/photos/%d/image?size=full", p.ID)
	resp.ThumbUrl = &thumbURL
	resp.PreviewUrl = &previewURL
	resp.FullUrl = &fullURL

	return resp
}
