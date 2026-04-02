package rest

import (
	"errors"
	"io"
	"net/http"

	"github.com/rgallagher/homephotos/domain/photo"
	"github.com/rgallagher/homephotos/services/imaging"
)

func (s *Server) GetPhotoImage(w http.ResponseWriter, r *http.Request, id int64, params GetPhotoImageParams) {
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

	size := string(params.Size)

	if size == "full" {
		s.serveFullImage(w, r, p)
		return
	}

	imgSize := imaging.Size(size)

	// On-demand: generate if not cached
	if p.CacheStatus != photo.CacheStatusCached || !s.cache.Has(p.Fingerprint, imgSize) {
		if err := s.cache.GenerateIfNeeded(r.Context(), p); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate image")
			return
		}
	}

	// Serve from cache
	path := s.cache.Path(p.Fingerprint, imgSize)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", `"`+p.Fingerprint+"-"+size+`"`)
	http.ServeFile(w, r, path)
}

func (s *Server) serveFullImage(w http.ResponseWriter, r *http.Request, p *photo.Photo) {
	rc, contentType, err := s.cache.SourceImageReader(p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read source image")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	io.Copy(w, rc)
}
