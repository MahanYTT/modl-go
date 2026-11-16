package homepage

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"time"

	"modl-go/internal/httpx"
	"modl-go/internal/modules/knowledgebase"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrInvalidURL = errors.New("invalid card URL")
)

const categoryDropdownActionType = "category_dropdown"

type KBProvider interface {
	GetCategoriesByIDs(ctx context.Context, serverID string, ids []string) ([]knowledgebase.Category, error)
	GetVisibleArticlesGroupedByCategoryIDs(ctx context.Context, serverID string, categoryIDs []string) (map[string][]knowledgebase.Article, error)
}

type Card struct {
	ID               string
	ServerID         string
	Title            string
	Description      string
	Icon             string
	IconColor        string
	ActionType       string
	ActionURL        string
	ActionButtonText string
	CategoryID       string
	BackgroundColor  string
	Ordinal          int32
	IsEnabled        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type EmbeddedCategory struct {
	ID   string
	Name string
	Slug string
}

type EnrichedCard struct {
	Card     Card
	Category *EmbeddedCategory
}

type EnrichedCardWithArticles struct {
	Card     Card
	Category *knowledgebase.Category
	Articles []knowledgebase.Article
}

type CardInput struct {
	Title            *string
	Description      *string
	Icon             *string
	IconColor        *string
	ActionType       *string
	ActionURL        *string
	ActionButtonText *string
	CategoryID       *string
	BackgroundColor  *string
	IsEnabled        *bool
}

type Service struct {
	db *sql.DB
	kb KBProvider
}

func NewService(db *sql.DB, kb KBProvider) *Service {
	if kb == nil {
		kb = knowledgebase.NewService(db)
	}
	return &Service{db: db, kb: kb}
}

func IsSafeURL(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return true
	}
	if strings.ContainsAny(trimmed, "\t\n\r") {
		return false
	}
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/\\") {
		return false
	}
	if strings.HasPrefix(trimmed, "/") {
		return true
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")
}

func (s *Service) GetAllCards(ctx context.Context, serverID string) ([]Card, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, title, description, icon, icon_color, action_type, action_url,
		        action_button_text, category_id, background_color, ordinal, is_enabled, created_at, updated_at
		 FROM homepage_cards
		 WHERE server_id = ?
		 ORDER BY ordinal ASC, created_at ASC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCards(rows)
}

func (s *Service) GetVisibleCards(ctx context.Context, serverID string) ([]Card, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, title, description, icon, icon_color, action_type, action_url,
		        action_button_text, category_id, background_color, ordinal, is_enabled, created_at, updated_at
		 FROM homepage_cards
		 WHERE server_id = ? AND is_enabled = 1
		 ORDER BY ordinal ASC, created_at ASC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCards(rows)
}

func (s *Service) GetAllCardsEnriched(ctx context.Context, serverID string) ([]EnrichedCard, error) {
	cards, err := s.GetAllCards(ctx, serverID)
	if err != nil {
		return nil, err
	}
	catIDs := distinctCategoryIDs(cards, false)
	catsByID := map[string]knowledgebase.Category{}
	if len(catIDs) > 0 {
		cats, err := s.kb.GetCategoriesByIDs(ctx, serverID, catIDs)
		if err != nil {
			return nil, err
		}
		for _, c := range cats {
			catsByID[c.ID] = c
		}
	}

	out := make([]EnrichedCard, 0, len(cards))
	for _, card := range cards {
		var embedded *EmbeddedCategory
		if card.CategoryID != "" {
			if c, ok := catsByID[card.CategoryID]; ok {
				embedded = &EmbeddedCategory{ID: c.ID, Name: c.Name, Slug: c.Slug}
			}
		}
		out = append(out, EnrichedCard{Card: card, Category: embedded})
	}
	return out, nil
}

func (s *Service) GetVisibleCardsEnrichedWithArticles(ctx context.Context, serverID string) ([]EnrichedCardWithArticles, error) {
	cards, err := s.GetVisibleCards(ctx, serverID)
	if err != nil {
		return nil, err
	}
	catIDs := distinctCategoryIDs(cards, true)
	catsByID := map[string]knowledgebase.Category{}
	articlesByCatID := map[string][]knowledgebase.Article{}
	if len(catIDs) > 0 {
		cats, err := s.kb.GetCategoriesByIDs(ctx, serverID, catIDs)
		if err != nil {
			return nil, err
		}
		for _, c := range cats {
			catsByID[c.ID] = c
		}
		articlesByCatID, err = s.kb.GetVisibleArticlesGroupedByCategoryIDs(ctx, serverID, catIDs)
		if err != nil {
			return nil, err
		}
	}

	out := make([]EnrichedCardWithArticles, 0, len(cards))
	for _, card := range cards {
		var cat *knowledgebase.Category
		var arts []knowledgebase.Article
		if card.ActionType == categoryDropdownActionType && card.CategoryID != "" {
			if c, ok := catsByID[card.CategoryID]; ok {
				cCopy := c
				cat = &cCopy
				arts = articlesByCatID[c.ID]
			}
		}
		out = append(out, EnrichedCardWithArticles{
			Card:     card,
			Category: cat,
			Articles: arts,
		})
	}
	return out, nil
}

func (s *Service) GetCardByID(ctx context.Context, serverID, id string) (*Card, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, server_id, title, description, icon, icon_color, action_type, action_url,
		        action_button_text, category_id, background_color, ordinal, is_enabled, created_at, updated_at
		 FROM homepage_cards
		 WHERE server_id = ? AND id = ?`,
		serverID, id,
	)
	card, err := scanCard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return card, err
}

func (s *Service) CreateCard(ctx context.Context, serverID string, title string, in CardInput) (*Card, error) {
	if in.ActionURL != nil && !IsSafeURL(*in.ActionURL) {
		return nil, ErrInvalidURL
	}

	var maxOrd sql.NullInt64
	if err := s.db.QueryRowContext(ctx,
		`SELECT MAX(ordinal) FROM homepage_cards WHERE server_id = ?`,
		serverID,
	).Scan(&maxOrd); err != nil {
		return nil, err
	}
	ordinal := int32(0)
	if maxOrd.Valid {
		ordinal = int32(maxOrd.Int64) + 1
	}

	enabled := true
	if in.IsEnabled != nil {
		enabled = *in.IsEnabled
	}
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	nowMs := httpx.NowMs()

	card := &Card{
		ID:               httpx.NewID(),
		ServerID:         serverID,
		Title:            title,
		Description:      deref(in.Description),
		Icon:             deref(in.Icon),
		IconColor:        deref(in.IconColor),
		ActionType:       deref(in.ActionType),
		ActionURL:        deref(in.ActionURL),
		ActionButtonText: deref(in.ActionButtonText),
		CategoryID:       deref(in.CategoryID),
		BackgroundColor:  deref(in.BackgroundColor),
		Ordinal:          ordinal,
		IsEnabled:        enabled,
		CreatedAt:        time.UnixMilli(nowMs).UTC(),
		UpdatedAt:        time.UnixMilli(nowMs).UTC(),
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO homepage_cards (
			id, server_id, title, description, icon, icon_color, action_type, action_url,
			action_button_text, category_id, background_color, ordinal, is_enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		card.ID, card.ServerID, card.Title, card.Description, card.Icon, card.IconColor,
		card.ActionType, card.ActionURL, card.ActionButtonText, card.CategoryID,
		card.BackgroundColor, card.Ordinal, enabledInt, nowMs, nowMs,
	)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (s *Service) UpdateCard(ctx context.Context, serverID, id string, in CardInput) (*Card, error) {
	if in.ActionURL != nil && !IsSafeURL(*in.ActionURL) {
		return nil, ErrInvalidURL
	}

	existing, err := s.GetCardByID(ctx, serverID, id)
	if err != nil {
		return nil, err
	}

	if in.Title != nil {
		existing.Title = *in.Title
	}
	if in.Description != nil {
		existing.Description = *in.Description
	}
	if in.Icon != nil {
		existing.Icon = *in.Icon
	}
	if in.IconColor != nil {
		existing.IconColor = *in.IconColor
	}
	if in.ActionType != nil {
		existing.ActionType = *in.ActionType
	}
	if in.ActionURL != nil {
		existing.ActionURL = *in.ActionURL
	}
	if in.ActionButtonText != nil {
		existing.ActionButtonText = *in.ActionButtonText
	}
	if in.CategoryID != nil {
		existing.CategoryID = *in.CategoryID
	}
	if in.BackgroundColor != nil {
		existing.BackgroundColor = *in.BackgroundColor
	}
	if in.IsEnabled != nil {
		existing.IsEnabled = *in.IsEnabled
	}

	nowMs := httpx.NowMs()
	existing.UpdatedAt = time.UnixMilli(nowMs).UTC()
	enabledInt := 0
	if existing.IsEnabled {
		enabledInt = 1
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE homepage_cards
		 SET title = ?, description = ?, icon = ?, icon_color = ?, action_type = ?, action_url = ?,
		     action_button_text = ?, category_id = ?, background_color = ?, is_enabled = ?, updated_at = ?
		 WHERE server_id = ? AND id = ?`,
		existing.Title, existing.Description, existing.Icon, existing.IconColor,
		existing.ActionType, existing.ActionURL, existing.ActionButtonText,
		existing.CategoryID, existing.BackgroundColor, enabledInt, nowMs, serverID, id,
	)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteCard(ctx context.Context, serverID, id string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM homepage_cards WHERE server_id = ? AND id = ?`,
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

func (s *Service) ReorderCards(ctx context.Context, serverID string, ids []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	nowMs := httpx.NowMs()
	stmt, err := tx.PrepareContext(ctx,
		`UPDATE homepage_cards SET ordinal = ?, updated_at = ? WHERE server_id = ? AND id = ?`,
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

func distinctCategoryIDs(cards []Card, dropdownOnly bool) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, c := range cards {
		if dropdownOnly && c.ActionType != categoryDropdownActionType {
			continue
		}
		if c.CategoryID == "" {
			continue
		}
		if _, ok := seen[c.CategoryID]; ok {
			continue
		}
		seen[c.CategoryID] = struct{}{}
		ids = append(ids, c.CategoryID)
	}
	return ids
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCard(row rowScanner) (*Card, error) {
	var c Card
	var enabled int
	var createdMs, updatedMs int64
	if err := row.Scan(
		&c.ID, &c.ServerID, &c.Title, &c.Description, &c.Icon, &c.IconColor,
		&c.ActionType, &c.ActionURL, &c.ActionButtonText, &c.CategoryID,
		&c.BackgroundColor, &c.Ordinal, &enabled, &createdMs, &updatedMs,
	); err != nil {
		return nil, err
	}
	c.IsEnabled = enabled != 0
	c.CreatedAt = time.UnixMilli(createdMs).UTC()
	c.UpdatedAt = time.UnixMilli(updatedMs).UTC()
	return &c, nil
}

func scanCards(rows *sql.Rows) ([]Card, error) {
	var out []Card
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}
