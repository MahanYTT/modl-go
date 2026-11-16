package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"modl-go/internal/httpx"
)

type FileRecord struct {
	ID          string `json:"id"`
	ServerID    string `json:"serverId"`
	Key         string `json:"key"`
	FileName    string `json:"fileName"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	Category    string `json:"category"`
	CreatedAt   int64  `json:"createdAt"`
}

type LocalStore struct {
	db      *sql.DB
	baseDir string
}

func NewLocalStore(db *sql.DB, baseDir string) (*LocalStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalStore{db: db, baseDir: baseDir}, nil
}

func (s *LocalStore) Save(ctx context.Context, serverID, category, fileName, contentType string, r io.Reader) (*FileRecord, error) {
	id := httpx.NewID()
	cleanName := filepath.Base(strings.ReplaceAll(fileName, "\\", "/"))
	if cleanName == "." || cleanName == "" {
		cleanName = "file.bin"
	}
	key := fmt.Sprintf("%s/%s/%s-%s", serverID, category, id, cleanName)
	diskPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(diskPath)
	if err != nil {
		return nil, err
	}
	n, copyErr := io.Copy(f, r)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(diskPath)
		return nil, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(diskPath)
		return nil, closeErr
	}

	now := time.Now().UnixMilli()
	rec := &FileRecord{
		ID:          id,
		ServerID:    serverID,
		Key:         key,
		FileName:    cleanName,
		Size:        n,
		ContentType: contentType,
		Category:    category,
		CreatedAt:   now,
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO storage_files (id, server_id, key, file_name, size, content_type, category, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.ServerID, rec.Key, rec.FileName, rec.Size, rec.ContentType, rec.Category, rec.CreatedAt,
	)
	if err != nil {
		_ = os.Remove(diskPath)
		return nil, err
	}
	return rec, nil
}

func (s *LocalStore) GetByID(ctx context.Context, serverID, id string) (*FileRecord, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, server_id, key, file_name, size, content_type, category, created_at FROM storage_files WHERE server_id = ? AND id = ?`,
		serverID, id,
	)
	var rec FileRecord
	if err := row.Scan(&rec.ID, &rec.ServerID, &rec.Key, &rec.FileName, &rec.Size, &rec.ContentType, &rec.Category, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *LocalStore) GetByKey(ctx context.Context, serverID, key string) (*FileRecord, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, server_id, key, file_name, size, content_type, category, created_at FROM storage_files WHERE server_id = ? AND key = ?`,
		serverID, key,
	)
	var rec FileRecord
	if err := row.Scan(&rec.ID, &rec.ServerID, &rec.Key, &rec.FileName, &rec.Size, &rec.ContentType, &rec.Category, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *LocalStore) List(ctx context.Context, serverID, category string) ([]FileRecord, error) {
	q := `SELECT id, server_id, key, file_name, size, content_type, category, created_at FROM storage_files WHERE server_id = ?`
	args := []any{serverID}
	if category != "" {
		q += ` AND category = ?`
		args = append(args, category)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FileRecord
	for rows.Next() {
		var rec FileRecord
		if err := rows.Scan(&rec.ID, &rec.ServerID, &rec.Key, &rec.FileName, &rec.Size, &rec.ContentType, &rec.Category, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (s *LocalStore) Delete(ctx context.Context, serverID, id string) error {
	rec, err := s.GetByID(ctx, serverID, id)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM storage_files WHERE server_id = ? AND id = ?`, serverID, id); err != nil {
		return err
	}
	_ = os.Remove(s.DiskPath(rec.Key))
	return nil
}

func (s *LocalStore) TotalSize(ctx context.Context, serverID string) (int64, int64, error) {
	var totalBytes, count int64
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(size), 0), COUNT(*) FROM storage_files WHERE server_id = ?`, serverID).Scan(&totalBytes, &count)
	return totalBytes, count, err
}

func (s *LocalStore) DiskPath(key string) string {
	clean := filepath.Clean(filepath.FromSlash(key))
	if strings.HasPrefix(clean, "..") {
		return filepath.Join(s.baseDir, "invalid")
	}
	return filepath.Join(s.baseDir, clean)
}
