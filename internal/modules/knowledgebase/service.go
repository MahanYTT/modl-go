package knowledgebase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"modl-go/internal/httpx"

	"golang.org/x/text/unicode/norm"
)

var (
	ErrNotFound = errors.New("not found")
)

const maxSearchResults = 20

type Category struct {
	ID          string
	ServerID    string
	Name        string
	Slug        string
	Description string
	Ordinal     int32
	IsVisible   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Article struct {
	ID         string
	ServerID   string
	Title      string
	Slug       string
	Content    string
	CategoryID string
	Ordinal    int32
	IsVisible  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetAllCategories(ctx context.Context, serverID string) ([]Category, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, name, slug, description, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_categories
		 WHERE server_id = ?
		 ORDER BY ordinal ASC, created_at ASC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCategories(rows)
}

func (s *Service) GetVisibleCategories(ctx context.Context, serverID string) ([]Category, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, name, slug, description, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_categories
		 WHERE server_id = ? AND is_visible = 1
		 ORDER BY ordinal ASC, created_at ASC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCategories(rows)
}

func (s *Service) GetCategoryByID(ctx context.Context, serverID, id string) (*Category, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, server_id, name, slug, description, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_categories
		 WHERE server_id = ? AND id = ?`,
		serverID, id,
	)
	cat, err := scanCategory(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return cat, err
}

func (s *Service) GetCategoriesByIDs(ctx context.Context, serverID string, ids []string) ([]Category, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, 0, len(ids)+1)
	args = append(args, serverID)
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx,
		fmt.Sprintf(
			`SELECT id, server_id, name, slug, description, ordinal, is_visible, created_at, updated_at
			 FROM knowledgebase_categories
			 WHERE server_id = ? AND id IN (%s)`,
			placeholders,
		),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCategories(rows)
}

func (s *Service) CreateCategory(ctx context.Context, serverID, name string, description *string) (*Category, error) {
	var maxOrd sql.NullInt64
	if err := s.db.QueryRowContext(ctx,
		`SELECT MAX(ordinal) FROM knowledgebase_categories WHERE server_id = ?`,
		serverID,
	).Scan(&maxOrd); err != nil {
		return nil, err
	}
	ordinal := int32(0)
	if maxOrd.Valid {
		ordinal = int32(maxOrd.Int64) + 1
	}

	slug, err := s.uniqueCategorySlug(ctx, serverID, Slugify(name), "")
	if err != nil {
		return nil, err
	}

	desc := ""
	if description != nil {
		desc = *description
	}
	now := time.Now().UTC()
	nowMs := now.UnixMilli()
	cat := &Category{
		ID:          httpx.NewID(),
		ServerID:    serverID,
		Name:        name,
		Slug:        slug,
		Description: desc,
		Ordinal:     ordinal,
		IsVisible:   true,
		CreatedAt:   time.UnixMilli(nowMs).UTC(),
		UpdatedAt:   time.UnixMilli(nowMs).UTC(),
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO knowledgebase_categories (id, server_id, name, slug, description, ordinal, is_visible, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		cat.ID, cat.ServerID, cat.Name, cat.Slug, cat.Description, cat.Ordinal, nowMs, nowMs,
	)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *Service) UpdateCategory(ctx context.Context, serverID, id string, name, description *string, isVisible *bool) (*Category, error) {
	existing, err := s.GetCategoryByID(ctx, serverID, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		existing.Name = *name
		slug, err := s.uniqueCategorySlug(ctx, serverID, Slugify(*name), id)
		if err != nil {
			return nil, err
		}
		existing.Slug = slug
	}
	if description != nil {
		existing.Description = *description
	}
	if isVisible != nil {
		existing.IsVisible = *isVisible
	}
	nowMs := httpx.NowMs()
	existing.UpdatedAt = time.UnixMilli(nowMs).UTC()

	visInt := 0
	if existing.IsVisible {
		visInt = 1
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE knowledgebase_categories
		 SET name = ?, slug = ?, description = ?, is_visible = ?, updated_at = ?
		 WHERE server_id = ? AND id = ?`,
		existing.Name, existing.Slug, existing.Description, visInt, nowMs, serverID, id,
	)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteCategory(ctx context.Context, serverID, id string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM knowledgebase_articles WHERE server_id = ? AND category_id = ?`,
		serverID, id,
	); err != nil {
		return false, err
	}
	res, err := tx.ExecContext(ctx,
		`DELETE FROM knowledgebase_categories WHERE server_id = ? AND id = ?`,
		serverID, id,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (s *Service) ReorderCategories(ctx context.Context, serverID string, ids []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	nowMs := httpx.NowMs()
	stmt, err := tx.PrepareContext(ctx,
		`UPDATE knowledgebase_categories SET ordinal = ?, updated_at = ? WHERE server_id = ? AND id = ?`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for idx, id := range ids {
		if _, err := stmt.ExecContext(ctx, idx, nowMs, serverID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) GetAllArticlesGroupedByCategory(ctx context.Context, serverID string) (map[string][]Article, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_articles
		 WHERE server_id = ?
		 ORDER BY ordinal ASC, created_at ASC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	articles, err := scanArticles(rows)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]Article)
	for _, a := range articles {
		grouped[a.CategoryID] = append(grouped[a.CategoryID], a)
	}
	return grouped, nil
}

func (s *Service) GetArticlesByCategory(ctx context.Context, serverID, categoryID string) ([]Article, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_articles
		 WHERE server_id = ? AND category_id = ?
		 ORDER BY ordinal ASC, created_at ASC`,
		serverID, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (s *Service) GetVisibleArticlesByCategory(ctx context.Context, serverID, categoryID string) ([]Article, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_articles
		 WHERE server_id = ? AND category_id = ? AND is_visible = 1
		 ORDER BY ordinal ASC, created_at ASC`,
		serverID, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (s *Service) GetVisibleArticlesGroupedByCategoryIDs(ctx context.Context, serverID string, categoryIDs []string) (map[string][]Article, error) {
	if len(categoryIDs) == 0 {
		return map[string][]Article{}, nil
	}
	placeholders := strings.Repeat("?,", len(categoryIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, 0, len(categoryIDs)+1)
	args = append(args, serverID)
	for _, id := range categoryIDs {
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx,
		fmt.Sprintf(
			`SELECT id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at
			 FROM knowledgebase_articles
			 WHERE server_id = ? AND category_id IN (%s) AND is_visible = 1
			 ORDER BY ordinal ASC, created_at ASC`,
			placeholders,
		),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	articles, err := scanArticles(rows)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]Article)
	for _, a := range articles {
		grouped[a.CategoryID] = append(grouped[a.CategoryID], a)
	}
	return grouped, nil
}

func (s *Service) GetArticleByID(ctx context.Context, serverID, id string) (*Article, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_articles
		 WHERE server_id = ? AND id = ?`,
		serverID, id,
	)
	a, err := scanArticle(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (s *Service) GetArticleBySlug(ctx context.Context, serverID, slug string) (*Article, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_articles
		 WHERE server_id = ? AND slug = ?`,
		serverID, slug,
	)
	a, err := scanArticle(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (s *Service) CreateArticle(ctx context.Context, serverID, categoryID, title, content string, isVisible *bool) (*Article, error) {
	var maxOrd sql.NullInt64
	if err := s.db.QueryRowContext(ctx,
		`SELECT MAX(ordinal) FROM knowledgebase_articles WHERE server_id = ? AND category_id = ?`,
		serverID, categoryID,
	).Scan(&maxOrd); err != nil {
		return nil, err
	}
	ordinal := int32(0)
	if maxOrd.Valid {
		ordinal = int32(maxOrd.Int64) + 1
	}

	slug, err := s.uniqueArticleSlug(ctx, serverID, Slugify(title), "")
	if err != nil {
		return nil, err
	}

	visible := true
	if isVisible != nil {
		visible = *isVisible
	}
	visInt := 0
	if visible {
		visInt = 1
	}
	nowMs := httpx.NowMs()
	article := &Article{
		ID:         httpx.NewID(),
		ServerID:   serverID,
		Title:      title,
		Slug:       slug,
		Content:    content,
		CategoryID: categoryID,
		Ordinal:    ordinal,
		IsVisible:  visible,
		CreatedAt:  time.UnixMilli(nowMs).UTC(),
		UpdatedAt:  time.UnixMilli(nowMs).UTC(),
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO knowledgebase_articles (id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		article.ID, article.ServerID, article.Title, article.Slug, article.Content, article.CategoryID, article.Ordinal, visInt, nowMs, nowMs,
	)
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (s *Service) UpdateArticle(ctx context.Context, serverID, id string, title, content *string, isVisible *bool) (*Article, error) {
	existing, err := s.GetArticleByID(ctx, serverID, id)
	if err != nil {
		return nil, err
	}

	if title != nil {
		existing.Title = *title
		slug, err := s.uniqueArticleSlug(ctx, serverID, Slugify(*title), id)
		if err != nil {
			return nil, err
		}
		existing.Slug = slug
	}
	if content != nil {
		existing.Content = *content
	}
	if isVisible != nil {
		existing.IsVisible = *isVisible
	}
	nowMs := httpx.NowMs()
	existing.UpdatedAt = time.UnixMilli(nowMs).UTC()

	visInt := 0
	if existing.IsVisible {
		visInt = 1
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE knowledgebase_articles
		 SET title = ?, slug = ?, content = ?, is_visible = ?, updated_at = ?
		 WHERE server_id = ? AND id = ?`,
		existing.Title, existing.Slug, existing.Content, visInt, nowMs, serverID, id,
	)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteArticle(ctx context.Context, serverID, id string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM knowledgebase_articles WHERE server_id = ? AND id = ?`,
		serverID, id,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (s *Service) SearchArticles(ctx context.Context, serverID, searchQuery string) ([]Article, error) {
	escaped := escapeLike(strings.ToLower(searchQuery))
	pattern := "%" + escaped + "%"
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, title, slug, content, category_id, ordinal, is_visible, created_at, updated_at
		 FROM knowledgebase_articles
		 WHERE server_id = ? AND is_visible = 1
		   AND (LOWER(title) LIKE ? ESCAPE '\' OR LOWER(content) LIKE ? ESCAPE '\')
		 ORDER BY ordinal ASC, created_at ASC
		 LIMIT ?`,
		serverID, pattern, pattern, maxSearchResults,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (s *Service) ReorderArticles(ctx context.Context, serverID, categoryID string, ids []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	nowMs := httpx.NowMs()
	stmt, err := tx.PrepareContext(ctx,
		`UPDATE knowledgebase_articles SET ordinal = ?, updated_at = ? WHERE server_id = ? AND category_id = ? AND id = ?`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for idx, id := range ids {
		if _, err := stmt.ExecContext(ctx, idx, nowMs, serverID, categoryID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) uniqueCategorySlug(ctx context.Context, serverID, baseSlug, excludeID string) (string, error) {
	slug := baseSlug
	for suffix := 1; ; suffix++ {
		var exists int
		var err error
		if excludeID != "" {
			err = s.db.QueryRowContext(ctx,
				`SELECT 1 FROM knowledgebase_categories WHERE server_id = ? AND slug = ? AND id != ? LIMIT 1`,
				serverID, slug, excludeID,
			).Scan(&exists)
		} else {
			err = s.db.QueryRowContext(ctx,
				`SELECT 1 FROM knowledgebase_categories WHERE server_id = ? AND slug = ? LIMIT 1`,
				serverID, slug,
			).Scan(&exists)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return slug, nil
		}
		if err != nil {
			return "", err
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
	}
}

func (s *Service) uniqueArticleSlug(ctx context.Context, serverID, baseSlug, excludeID string) (string, error) {
	slug := baseSlug
	for suffix := 1; ; suffix++ {
		var exists int
		var err error
		if excludeID != "" {
			err = s.db.QueryRowContext(ctx,
				`SELECT 1 FROM knowledgebase_articles WHERE server_id = ? AND slug = ? AND id != ? LIMIT 1`,
				serverID, slug, excludeID,
			).Scan(&exists)
		} else {
			err = s.db.QueryRowContext(ctx,
				`SELECT 1 FROM knowledgebase_articles WHERE server_id = ? AND slug = ? LIMIT 1`,
				serverID, slug,
			).Scan(&exists)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return slug, nil
		}
		if err != nil {
			return "", err
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
	}
}

func Slugify(input string) string {
	decomposed := norm.NFD.String(strings.TrimSpace(input))
	var b strings.Builder
	b.Grow(len(decomposed))
	lastDash := false
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		r = unicode.ToLower(r)
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	res := strings.Trim(b.String(), "-")
	if res == "" {
		return "item"
	}
	return res
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCategory(row rowScanner) (*Category, error) {
	var c Category
	var vis int
	var createdMs, updatedMs int64
	if err := row.Scan(&c.ID, &c.ServerID, &c.Name, &c.Slug, &c.Description, &c.Ordinal, &vis, &createdMs, &updatedMs); err != nil {
		return nil, err
	}
	c.IsVisible = vis != 0
	c.CreatedAt = time.UnixMilli(createdMs).UTC()
	c.UpdatedAt = time.UnixMilli(updatedMs).UTC()
	return &c, nil
}

func scanCategories(rows *sql.Rows) ([]Category, error) {
	var out []Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func scanArticle(row rowScanner) (*Article, error) {
	var a Article
	var vis int
	var createdMs, updatedMs int64
	if err := row.Scan(&a.ID, &a.ServerID, &a.Title, &a.Slug, &a.Content, &a.CategoryID, &a.Ordinal, &vis, &createdMs, &updatedMs); err != nil {
		return nil, err
	}
	a.IsVisible = vis != 0
	a.CreatedAt = time.UnixMilli(createdMs).UTC()
	a.UpdatedAt = time.UnixMilli(updatedMs).UTC()
	return &a, nil
}

func scanArticles(rows *sql.Rows) ([]Article, error) {
	var out []Article
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}
