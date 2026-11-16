package media

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	PanelMediaPolicy = models.AccessPolicy{
		Rule:   models.RuleRequirePermission,
		View:   "admin.settings.view.content",
		Modify: "admin.settings.modify.content",
	}
	PanelStoragePolicy = models.AccessPolicy{
		Rule:   models.RuleRequirePermission,
		View:   "admin.settings.view.storage",
		Modify: "admin.settings.modify.storage",
	}
	PanelStorageSyncPolicy = models.AccessPolicy{
		Rule: models.RuleSuperAdmin,
	}
)

var publicAllowedUploadTypes = map[string]bool{
	"ticket":  true,
	"tickets": true,
	"appeal":  true,
}

const tempLimitMessage = "Temporary upload limit reached. Attach your files to a ticket or appeal, or try again later."

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- Routers ---

func (h *Handler) PanelMediaRouter(perms models.PermissionChecker) chi.Router {
	r := chi.NewRouter()
	h.RegisterPanelMedia(r, perms)
	return r
}

func (h *Handler) RegisterPanelMedia(r chi.Router, perms models.PermissionChecker) {
	r.Group(func(r chi.Router) {
		r.Use(httpx.RequirePanelAuth)
		if perms != nil {
			r.Use(httpx.RequirePanelPolicy(perms, PanelMediaPolicy))
		}

		r.Get("/config", h.panelGetMediaConfig)
		r.Post("/presign", h.panelPresignUpload)
		r.Post("/confirm", h.panelConfirmUpload)
		r.Post("/upload", h.panelDirectUpload)
		r.Delete("/*", h.panelDeleteFile)
	})
}

func (h *Handler) PublicMediaRouter() chi.Router {
	r := chi.NewRouter()
	h.RegisterPublicMedia(r)
	return r
}

func (h *Handler) RegisterPublicMedia(r chi.Router) {
	r.Get("/config", h.publicGetMediaConfig)
	r.Post("/presign", h.publicPresignUpload)
	r.Post("/confirm", h.publicConfirmUpload)
	r.Put("/upload", h.handlePresignedPutUpload)
	r.Post("/upload", h.publicDirectUpload)
	r.Get("/file/{id}", h.serveMediaFile)
}

func (h *Handler) PanelStorageRouter(perms models.PermissionChecker) chi.Router {
	r := chi.NewRouter()
	h.RegisterPanelStorage(r, perms)
	return r
}

func (h *Handler) RegisterPanelStorage(r chi.Router, perms models.PermissionChecker) {
	r.Group(func(r chi.Router) {
		r.Use(httpx.RequirePanelAuth)
		if perms != nil {
			r.Use(httpx.RequirePanelPolicy(perms, PanelStoragePolicy))
		}

		r.Get("/quota", h.storageGetQuota)
		r.Get("/files", h.storageGetFiles)
		r.Post("/bulk-delete", h.storageBulkDelete)
		r.Get("/download/*", h.storageGetDownloadURL)
	})

	r.Group(func(r chi.Router) {
		r.Use(httpx.RequirePanelAuth)
		if perms != nil {
			r.Use(httpx.RequirePanelPolicy(perms, PanelStorageSyncPolicy))
		}
		r.Post("/sync", h.storageSyncFiles)
	})
}

func (h *Handler) EvidenceUploadRouter() chi.Router {
	r := chi.NewRouter()
	h.RegisterEvidenceUpload(r)
	return r
}

func (h *Handler) RegisterEvidenceUpload(r chi.Router) {
	r.Get("/file/{id}", h.serveEvidenceFile)
	r.Get("/{token}", h.evidenceValidateToken)
	r.Post("/{token}/presign", h.evidencePresignUpload)
	r.Put("/{token}/upload", h.evidencePresignedPutUpload)
	r.Post("/{token}/upload", h.evidenceDirectUpload)
	r.Post("/{token}/confirm", h.evidenceConfirmUpload)
	r.Post("/{token}/submit", h.evidenceSubmit)
}

// --- PanelMediaController ---

func (h *Handler) panelGetMediaConfig(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, buildMediaConfigResponse(h.svc.IsConfigured(), false))
}

func (h *Handler) panelDeleteFile(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	key := extractWildcardKey(r, "/v1/panel/media/")
	deleted, err := h.svc.DeleteByKey(r.Context(), srv.ID, key)
	if errors.Is(err, ErrForbidden) {
		httpx.WriteError(w, r, http.StatusForbidden, "Access denied")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to delete file")
		return
	}
	if !deleted {
		httpx.WriteError(w, r, http.StatusNotFound, "File not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "File deleted"})
}

func (h *Handler) panelPresignUpload(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.PresignUploadRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	entityID := ""
	if req.EntityId != nil {
		entityID = *req.EntityId
	}
	res, err := h.svc.PresignUpload(
		r.Context(),
		srv.ID,
		req.UploadType,
		req.FileName,
		req.ContentType,
		req.FileSize,
		entityID,
		false,
		false,
		func(key string, exp int64, sig string) string {
			return buildMediaUploadURL(r, key, exp, sig)
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrStorageNotConfigured):
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "File storage is not configured"})
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrQuotaExceeded), errors.Is(err, ErrTempLimitExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate upload URL"})
		}
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toPresignProto(res))
}

func (h *Handler) panelConfirmUpload(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.ConfirmUploadRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	res, err := h.svc.ConfirmUpload(r.Context(), srv.ID, req.Key, false, func(id string) string {
		return buildMediaFileURL(r, id)
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			httpx.WriteError(w, r, http.StatusForbidden, "Access denied")
		case errors.Is(err, ErrUploadNotFound):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "Upload not found",
				"message": "The file was not uploaded or the presigned URL expired",
			})
		case errors.Is(err, ErrQuotaExceeded), errors.Is(err, ErrTempLimitExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Storage quota exceeded"})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to record upload"})
		}
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toUploadProto(res))
}

func (h *Handler) panelDirectUpload(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid multipart form"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	uploadType := strings.TrimSpace(r.FormValue("uploadType"))
	if uploadType == "" {
		uploadType = "article"
	}
	entityID := strings.TrimSpace(r.FormValue("entityId"))
	contentType := header.Header.Get("Content-Type")
	res, err := h.svc.DirectUpload(
		r.Context(),
		srv.ID,
		uploadType,
		header.Filename,
		contentType,
		header.Size,
		entityID,
		false,
		false,
		file,
		func(id string) string { return buildMediaFileURL(r, id) },
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrQuotaExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to upload file"})
		}
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toUploadProto(res))
}

// --- PublicMediaController ---

func (h *Handler) publicGetMediaConfig(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, buildMediaConfigResponse(h.svc.IsConfigured(), false))
}

func (h *Handler) publicPresignUpload(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.PresignUploadRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !publicAllowedUploadTypes[req.UploadType] {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Upload type not allowed for public uploads. Allowed: [ticket, tickets, appeal]",
		})
		return
	}
	entityID := ""
	if req.EntityId != nil {
		entityID = strings.TrimSpace(*req.EntityId)
	}
	accessToken := ""
	if req.AccessToken != nil {
		accessToken = *req.AccessToken
	}

	status, msg := h.svc.ValidatePublicUploadAccess(r.Context(), srv.ID, req.UploadType, entityID, accessToken)
	if status == http.StatusNotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Not found")
		return
	}
	if status != http.StatusOK {
		httpx.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}

	res, err := h.svc.PresignUpload(
		r.Context(),
		srv.ID,
		NormalizeUploadType(req.UploadType),
		req.FileName,
		req.ContentType,
		req.FileSize,
		entityID,
		false,
		IsAnonymousTempEntity(entityID),
		func(key string, exp int64, sig string) string {
			return buildMediaUploadURL(r, key, exp, sig)
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrQuotaExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrTempLimitExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": tempLimitMessage})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate upload URL"})
		}
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toPresignProto(res))
}

func (h *Handler) publicConfirmUpload(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.ConfirmUploadRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !IsKeyOwnedByServer(req.Key, srv.ID) {
		httpx.WriteError(w, r, http.StatusForbidden, "Access denied")
		return
	}
	uploadType := ExtractUploadType(req.Key)
	entityID := ExtractEntityID(req.Key)
	if !publicAllowedUploadTypes[uploadType] {
		httpx.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Upload type not allowed for public confirmation"})
		return
	}
	if strings.TrimSpace(entityID) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid upload key")
		return
	}
	accessToken := ""
	if req.AccessToken != nil {
		accessToken = *req.AccessToken
	}

	status, msg := h.svc.ValidatePublicUploadAccess(r.Context(), srv.ID, uploadType, entityID, accessToken)
	if status == http.StatusNotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Not found")
		return
	}
	if status != http.StatusOK {
		httpx.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}

	res, err := h.svc.ConfirmUpload(r.Context(), srv.ID, req.Key, IsAnonymousTempEntity(entityID), func(id string) string {
		return buildMediaFileURL(r, id)
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUploadNotFound):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "Upload not found",
				"message": "The file was not uploaded or the presigned URL expired",
			})
		case errors.Is(err, ErrQuotaExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Storage quota exceeded"})
		case errors.Is(err, ErrTempLimitExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": tempLimitMessage})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to record upload"})
		}
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toUploadProto(res))
}

func (h *Handler) publicDirectUpload(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid multipart form"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	uploadType := strings.TrimSpace(r.FormValue("uploadType"))
	if !publicAllowedUploadTypes[uploadType] {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Upload type not allowed for public uploads. Allowed: [ticket, tickets, appeal]",
		})
		return
	}
	entityID := strings.TrimSpace(r.FormValue("entityId"))
	accessToken := strings.TrimSpace(r.FormValue("accessToken"))

	status, msg := h.svc.ValidatePublicUploadAccess(r.Context(), srv.ID, uploadType, entityID, accessToken)
	if status == http.StatusNotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Not found")
		return
	}
	if status != http.StatusOK {
		httpx.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}

	res, err := h.svc.DirectUpload(
		r.Context(),
		srv.ID,
		NormalizeUploadType(uploadType),
		header.Filename,
		header.Header.Get("Content-Type"),
		header.Size,
		entityID,
		false,
		IsAnonymousTempEntity(entityID),
		file,
		func(id string) string { return buildMediaFileURL(r, id) },
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrQuotaExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrTempLimitExceeded):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": tempLimitMessage})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to upload file"})
		}
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toUploadProto(res))
}

func (h *Handler) handlePresignedPutUpload(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key := q.Get("key")
	exp := q.Get("exp")
	sig := q.Get("sig")
	if key == "" || exp == "" || sig == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing upload parameters"})
		return
	}
	err := h.svc.HandlePresignedPut(key, exp, sig, r.Header.Get("Content-Type"), r.ContentLength, r.Body)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			httpx.WriteJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrInvalidInput):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to store upload"})
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "key": key})
}

func (h *Handler) serveMediaFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	serverID := ""
	if srv := httpx.GetServer(r); srv != nil {
		serverID = srv.ID
	}
	rec, diskPath, err := h.svc.GetFileByID(r.Context(), serverID, id)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "File not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to serve file")
		return
	}
	serveDiskFile(w, r, diskPath, rec.FileName, rec.ContentType)
}

// --- PanelStorageController ---

func (h *Handler) storageGetQuota(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	q, err := h.svc.GetQuota(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load storage quota")
		return
	}
	resp := &modlv1.StorageQuotaResponse{
		UsedBytes:          q.UsedBytes,
		MaxBytes:           q.MaxBytes,
		UsedPercentage:     q.UsedPercentage,
		UsedFormatted:      q.UsedFormatted,
		MaxFormatted:       q.MaxFormatted,
		ByType:             q.ByType,
		IsPremium:          q.IsPremium,
		StorageOverageRate: q.StorageOverageRate,
		AiQuota: &modlv1.AiQuotaInfo{
			ByService: map[string]int64{"other": 0},
		},
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) storageGetFiles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	prefix := r.URL.Query().Get("prefix")
	files, err := h.svc.ListFiles(r.Context(), srv.ID, prefix)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to list files")
		return
	}
	resp := &modlv1.StorageFilesResponse{
		Files: make([]*modlv1.StorageFileResponse, 0, len(files)),
	}
	for _, f := range files {
		ct := f.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		resp.Files = append(resp.Files, &modlv1.StorageFileResponse{
			Key:          f.Key,
			Name:         f.FileName,
			Size:         f.Size,
			ContentType:  ct,
			LastModified: timestamppb.New(time.UnixMilli(f.CreatedAt).UTC()),
			Url:          buildMediaFileURL(r, f.ID),
		})
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) storageBulkDelete(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.BulkDeleteRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	deleted, err := h.svc.BulkDelete(r.Context(), srv.ID, req.Keys)
	if errors.Is(err, ErrForbidden) {
		httpx.WriteError(w, r, http.StatusForbidden, "Access denied")
		return
	}
	if errors.Is(err, ErrInvalidInput) {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to delete files")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StorageBulkDeleteResponse{Deleted: int32(deleted)})
}

func (h *Handler) storageSyncFiles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	synced, err := h.svc.SyncFiles(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to sync storage files")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StorageSyncResponse{Synced: int32(synced)})
}

func (h *Handler) storageGetDownloadURL(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	key := extractWildcardKey(r, "/v1/panel/storage/download/")
	rec, err := h.svc.GetDownloadRecord(r.Context(), srv.ID, key)
	if errors.Is(err, ErrForbidden) {
		httpx.WriteError(w, r, http.StatusForbidden, "Access denied")
		return
	}
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "File not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to get download URL")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StorageDownloadUrlResponse{
		Url: buildMediaFileURL(r, rec.ID),
	})
}

// --- EvidenceUploadController ---

func (h *Handler) evidenceValidateToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	info, err := h.svc.GetEvidenceTokenInfo(r.Context(), token)
	if errors.Is(err, ErrInvalidToken) {
		evidenceError(w, http.StatusNotFound, "Invalid or expired upload token")
		return
	}
	if err != nil {
		evidenceError(w, http.StatusInternalServerError, "Failed to validate upload token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":       200,
		"punishmentId": info.PunishmentID,
		"playerName":   info.PlayerName,
		"issuerName":   info.IssuerName,
	})
}

func (h *Handler) evidencePresignUpload(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	var req struct {
		FileName    string `json:"fileName"`
		ContentType string `json:"contentType"`
		FileSize    int64  `json:"fileSize"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		evidenceError(w, http.StatusBadRequest, "Malformed request body")
		return
	}
	if strings.TrimSpace(req.FileName) == "" || len(req.FileName) > 255 {
		evidenceError(w, http.StatusBadRequest, "fileName: File name is required")
		return
	}
	if strings.TrimSpace(req.ContentType) == "" || len(req.ContentType) > 128 {
		evidenceError(w, http.StatusBadRequest, "contentType: Content type is required")
		return
	}
	if req.FileSize <= 0 {
		evidenceError(w, http.StatusBadRequest, "fileSize: File size must be positive")
		return
	}

	res, err := h.svc.PresignEvidenceUpload(r.Context(), token, req.FileName, req.ContentType, req.FileSize,
		func(key string, exp int64, sig string) string {
			return buildEvidenceUploadURL(r, token, key, exp, sig)
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			evidenceError(w, http.StatusNotFound, "Invalid or expired upload token")
		case errors.Is(err, ErrStorageNotConfigured):
			evidenceError(w, http.StatusServiceUnavailable, "File storage is not configured")
		case errors.Is(err, ErrQuotaExceeded):
			evidenceError(w, http.StatusBadRequest, "Storage quota exceeded. Please contact the server administrator.")
		case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrTempLimitExceeded):
			evidenceError(w, http.StatusBadRequest, err.Error())
		default:
			evidenceError(w, http.StatusInternalServerError, "Failed to generate upload URL")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":          200,
		"presignedUrl":    res.PresignedURL,
		"key":             res.Key,
		"expiresAt":       res.ExpiresAt.Format(time.RFC3339),
		"method":          res.Method,
		"requiredHeaders": res.RequiredHeaders,
	})
}

func (h *Handler) evidencePresignedPutUpload(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	tok, err := h.svc.ValidateEvidenceToken(r.Context(), token)
	if err != nil {
		evidenceError(w, http.StatusNotFound, "Invalid or expired upload token")
		return
	}
	q := r.URL.Query()
	key := q.Get("key")
	exp := q.Get("exp")
	sig := q.Get("sig")
	expectedPrefix := tok.ServerID + "/evidence/" + tok.PunishmentID + "/"
	if !strings.HasPrefix(key, expectedPrefix) {
		evidenceError(w, http.StatusForbidden, "Upload key does not belong to this evidence token")
		return
	}
	if key == "" || exp == "" || sig == "" {
		evidenceError(w, http.StatusBadRequest, "Missing upload parameters")
		return
	}
	if err := h.svc.HandlePresignedPut(key, exp, sig, r.Header.Get("Content-Type"), r.ContentLength, r.Body); err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			evidenceError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, ErrInvalidInput):
			evidenceError(w, http.StatusBadRequest, err.Error())
		default:
			evidenceError(w, http.StatusInternalServerError, "Failed to store upload")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "key": key})
}

func (h *Handler) evidenceDirectUpload(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		evidenceError(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		evidenceError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	res, err := h.svc.DirectEvidenceUpload(
		r.Context(),
		token,
		header.Filename,
		header.Header.Get("Content-Type"),
		header.Size,
		file,
		func(id string) string { return buildEvidenceFileURL(r, id) },
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			evidenceError(w, http.StatusNotFound, "Invalid or expired upload token")
		case errors.Is(err, ErrQuotaExceeded):
			evidenceError(w, http.StatusBadRequest, "Storage quota exceeded. Please contact the server administrator.")
		case errors.Is(err, ErrInvalidInput):
			evidenceError(w, http.StatusBadRequest, err.Error())
		default:
			evidenceError(w, http.StatusInternalServerError, "Failed to upload evidence")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":      200,
		"key":         res.Key,
		"url":         res.URL,
		"fileName":    res.FileName,
		"size":        res.Size,
		"contentType": res.ContentType,
	})
}

func (h *Handler) evidenceConfirmUpload(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	var req struct {
		Key string `json:"key"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		evidenceError(w, http.StatusBadRequest, "Malformed request body")
		return
	}
	if strings.TrimSpace(req.Key) == "" || len(req.Key) > 1024 {
		evidenceError(w, http.StatusBadRequest, "key: Key is required")
		return
	}
	res, err := h.svc.ConfirmEvidenceUpload(r.Context(), token, req.Key, func(id string) string {
		return buildEvidenceFileURL(r, id)
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			evidenceError(w, http.StatusNotFound, "Invalid or expired upload token")
		case errors.Is(err, ErrForbidden):
			evidenceError(w, http.StatusForbidden, "Upload key does not belong to this evidence token")
		case errors.Is(err, ErrUploadNotFound):
			evidenceError(w, http.StatusNotFound, "Upload not found. File may not have been uploaded yet.")
		case errors.Is(err, ErrQuotaExceeded):
			evidenceError(w, http.StatusBadRequest, "Storage quota exceeded")
		default:
			evidenceError(w, http.StatusInternalServerError, "Failed to record upload")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":      200,
		"key":         res.Key,
		"url":         res.URL,
		"fileName":    res.FileName,
		"size":        res.Size,
		"contentType": res.ContentType,
	})
}

func (h *Handler) evidenceSubmit(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	var req struct {
		Evidence []struct {
			URL      string `json:"url"`
			FileName string `json:"fileName"`
			FileType string `json:"fileType"`
			FileSize int64  `json:"fileSize"`
		} `json:"evidence"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		evidenceError(w, http.StatusBadRequest, "Malformed request body")
		return
	}
	if len(req.Evidence) == 0 || len(req.Evidence) > EvidenceMaxItems {
		evidenceError(w, http.StatusBadRequest, "evidence: must contain between 1 and 10 items")
		return
	}
	items := make([]EvidenceSubmitItem, 0, len(req.Evidence))
	for _, it := range req.Evidence {
		if strings.TrimSpace(it.URL) == "" || len(it.URL) > 2048 ||
			strings.TrimSpace(it.FileName) == "" || len(it.FileName) > 255 ||
			strings.TrimSpace(it.FileType) == "" || len(it.FileType) > 128 ||
			it.FileSize < 0 {
			evidenceError(w, http.StatusBadRequest, "Invalid evidence item")
			return
		}
		items = append(items, EvidenceSubmitItem{
			URL:      it.URL,
			FileName: it.FileName,
			FileType: it.FileType,
			FileSize: it.FileSize,
		})
	}
	err := h.svc.SubmitEvidence(r.Context(), token, items)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			evidenceError(w, http.StatusNotFound, "Invalid or expired upload token")
		case errors.Is(err, ErrInvalidEvidenceURL):
			evidenceError(w, http.StatusBadRequest, "Invalid evidence URL")
		case errors.Is(err, ErrPunishmentNotFound):
			evidenceError(w, http.StatusNotFound, "Punishment not found")
		default:
			evidenceError(w, http.StatusInternalServerError, "Failed to submit evidence")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Evidence uploaded successfully",
	})
}

func (h *Handler) serveEvidenceFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rec, diskPath, err := h.svc.GetEvidenceFileByID(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		evidenceError(w, http.StatusNotFound, "File not found")
		return
	}
	if err != nil {
		evidenceError(w, http.StatusInternalServerError, "Failed to serve file")
		return
	}
	serveDiskFile(w, r, diskPath, rec.FileName, rec.ContentType)
}

// --- Helpers ---

func buildMediaConfigResponse(configured, isPremium bool) *modlv1.MediaConfigResponse {
	var supported map[string][]string
	var limits map[string]int64
	if configured {
		supported = AllSupportedTypes()
		limits = AllSizeLimits(isPremium)
	} else {
		supported = map[string][]string{
			"evidence": {}, "tickets": {}, "appeals": {}, "articles": {}, "server-icons": {}, "replays": {},
		}
		limits = map[string]int64{
			"evidence": 0, "tickets": 0, "appeals": 0, "articles": 0, "server-icons": 0, "replays": 0,
		}
	}
	pbSupported := make(map[string]*modlv1.MediaSupportedTypes, len(supported))
	for k, v := range supported {
		pbSupported[k] = &modlv1.MediaSupportedTypes{Types: v}
	}
	return &modlv1.MediaConfigResponse{
		BackblazeConfigured: configured,
		SupportedTypes:      pbSupported,
		FileSizeLimits:      limits,
	}
}

func toPresignProto(res *PresignResult) *modlv1.PresignUploadResponse {
	return &modlv1.PresignUploadResponse{
		PresignedUrl:    res.PresignedURL,
		Key:             res.Key,
		ExpiresAt:       timestamppb.New(res.ExpiresAt),
		Method:          res.Method,
		RequiredHeaders: res.RequiredHeaders,
	}
}

func toUploadProto(res *UploadResult) *modlv1.UploadResponse {
	return &modlv1.UploadResponse{
		Key:         res.Key,
		Url:         res.URL,
		FileName:    res.FileName,
		Size:        res.Size,
		ContentType: res.ContentType,
	}
}

func evidenceError(w http.ResponseWriter, status int, message string) {
	httpx.WriteJSON(w, status, map[string]any{
		"status":  status,
		"message": message,
	})
}

func extractWildcardKey(r *http.Request, fallbackPrefix string) string {
	key := chi.URLParam(r, "*")
	if key == "" && strings.HasPrefix(r.URL.Path, fallbackPrefix) {
		key = strings.TrimPrefix(r.URL.Path, fallbackPrefix)
	}
	if unescaped, err := url.PathUnescape(key); err == nil {
		key = unescaped
	}
	return StripLeadingSlash(key)
}

func requestBaseURL(r *http.Request) string {
	if r == nil || r.Host == "" {
		return ""
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func buildMediaUploadURL(r *http.Request, key string, exp int64, sig string) string {
	q := url.Values{}
	q.Set("key", key)
	q.Set("exp", fmt.Sprintf("%d", exp))
	q.Set("sig", sig)
	return requestBaseURL(r) + "/v1/public/media/upload?" + q.Encode()
}

func buildEvidenceUploadURL(r *http.Request, token, key string, exp int64, sig string) string {
	q := url.Values{}
	q.Set("key", key)
	q.Set("exp", fmt.Sprintf("%d", exp))
	q.Set("sig", sig)
	return requestBaseURL(r) + "/v1/public/evidence-upload/" + url.PathEscape(token) + "/upload?" + q.Encode()
}

func buildMediaFileURL(r *http.Request, id string) string {
	return requestBaseURL(r) + "/v1/public/media/file/" + url.PathEscape(id)
}

func buildEvidenceFileURL(r *http.Request, id string) string {
	return requestBaseURL(r) + "/v1/public/evidence-upload/file/" + url.PathEscape(id)
}

func serveDiskFile(w http.ResponseWriter, r *http.Request, diskPath, fileName, contentType string) {
	f, err := os.Open(diskPath)
	if err != nil {
		httpx.WriteError(w, r, http.StatusNotFound, "File not found")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		httpx.WriteError(w, r, http.StatusNotFound, "File not found")
		return
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	safeName := strings.NewReplacer(`"`, "", "\r", "", "\n", "").Replace(filepath.Base(fileName))
	if safeName == "" {
		safeName = "file"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, safeName))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeContent(w, r, safeName, info.ModTime(), f)
}
