package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// SettingsRepo is a pgxpool-backed implementation of port.SettingsRepository
// covering the key/value app settings, branding configuration and fee tiers.
type SettingsRepo struct {
	pool *pgxpool.Pool
}

var _ port.SettingsRepository = (*SettingsRepo)(nil)

// NewSettingsRepo creates a new SettingsRepo.
func NewSettingsRepo(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{pool: pool}
}

// --- Key/value settings ---

const sqlGetSetting = `SELECT value FROM app_settings WHERE key = $1`

func (r *SettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	var v string
	err := r.pool.QueryRow(ctx, sqlGetSetting, key).Scan(&v)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("setting %q not found", key)
		}
		return "", err
	}
	return v, nil
}

const sqlSetSetting = `
INSERT INTO app_settings (key, value) VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`

func (r *SettingsRepo) SetSetting(ctx context.Context, key, value string) error {
	_, err := r.pool.Exec(ctx, sqlSetSetting, key, value)
	if err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

const sqlGetAllSettings = `SELECT key, value FROM app_settings`

func (r *SettingsRepo) GetAllSettings(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, sqlGetAllSettings)
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// --- Branding ---

const sqlGetBranding = `
SELECT club_name, primary_color, accent_color, logo_url
FROM branding_config WHERE id = 1`

func (r *SettingsRepo) GetBranding(ctx context.Context) (*domain.BrandingConfig, error) {
	var b domain.BrandingConfig
	err := r.pool.QueryRow(ctx, sqlGetBranding).
		Scan(&b.ClubName, &b.PrimaryColor, &b.AccentColor, &b.LogoURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			def := domain.DefaultBranding()
			return &def, nil
		}
		return nil, err
	}
	return &b, nil
}

const sqlUpsertBranding = `
INSERT INTO branding_config (id, club_name, primary_color, accent_color, logo_url)
VALUES (1, $1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET club_name = EXCLUDED.club_name, primary_color = EXCLUDED.primary_color,
    accent_color = EXCLUDED.accent_color, logo_url = EXCLUDED.logo_url`

func (r *SettingsRepo) UpdateBranding(ctx context.Context, b *domain.BrandingConfig) error {
	_, err := r.pool.Exec(ctx, sqlUpsertBranding, b.ClubName, b.PrimaryColor, b.AccentColor, b.LogoURL)
	if err != nil {
		return fmt.Errorf("update branding: %w", err)
	}
	return nil
}

// --- Fee tiers ---

const sqlGetFeeTiers = `
SELECT id, club_year_id, position, amount_cents
FROM fee_tiers WHERE club_year_id = $1 ORDER BY position`

func (r *SettingsRepo) GetFeeTiers(ctx context.Context, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	rows, err := r.pool.Query(ctx, sqlGetFeeTiers, clubYearID)
	if err != nil {
		return nil, fmt.Errorf("get fee tiers: %w", err)
	}
	defer rows.Close()

	var tiers []*domain.FeeTier
	for rows.Next() {
		var t domain.FeeTier
		if err := rows.Scan(&t.ID, &t.ClubYearID, &t.Position, &t.AmountCents); err != nil {
			return nil, err
		}
		tiers = append(tiers, &t)
	}
	return tiers, rows.Err()
}

// ReplaceFeeTiers atomically deletes the existing fee tiers for a club year and
// inserts the provided set.
func (r *SettingsRepo) ReplaceFeeTiers(ctx context.Context, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM fee_tiers WHERE club_year_id = $1`, clubYearID); err != nil {
		return fmt.Errorf("delete fee tiers: %w", err)
	}

	for _, t := range tiers {
		if t.ID == uuid.Nil {
			t.ID = uuid.New()
		}
		t.ClubYearID = clubYearID
		if _, err := tx.Exec(ctx,
			`INSERT INTO fee_tiers (id, club_year_id, position, amount_cents) VALUES ($1, $2, $3, $4)`,
			t.ID, t.ClubYearID, t.Position, t.AmountCents,
		); err != nil {
			return fmt.Errorf("insert fee tier: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// --- Per-member fee tier overrides (G-004) ---

const sqlGetMemberFeeTiers = `
SELECT id, member_id, club_year_id, position, amount_cents
FROM member_fee_tiers WHERE member_id = $1 AND club_year_id = $2 ORDER BY position`

func (r *SettingsRepo) GetMemberFeeTiers(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	rows, err := r.pool.Query(ctx, sqlGetMemberFeeTiers, memberID, clubYearID)
	if err != nil {
		return nil, fmt.Errorf("get member fee tiers: %w", err)
	}
	defer rows.Close()

	var tiers []*domain.FeeTier
	for rows.Next() {
		var t domain.FeeTier
		var mid uuid.UUID
		if err := rows.Scan(&t.ID, &mid, &t.ClubYearID, &t.Position, &t.AmountCents); err != nil {
			return nil, err
		}
		tiers = append(tiers, &t)
	}
	return tiers, rows.Err()
}

func (r *SettingsRepo) ReplaceMemberFeeTiers(ctx context.Context, memberID, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`DELETE FROM member_fee_tiers WHERE member_id = $1 AND club_year_id = $2`,
		memberID, clubYearID,
	); err != nil {
		return fmt.Errorf("delete member fee tiers: %w", err)
	}

	for _, t := range tiers {
		if t.ID == uuid.Nil {
			t.ID = uuid.New()
		}
		t.ClubYearID = clubYearID
		if _, err := tx.Exec(ctx,
			`INSERT INTO member_fee_tiers (id, member_id, club_year_id, position, amount_cents) VALUES ($1, $2, $3, $4, $5)`,
			t.ID, memberID, t.ClubYearID, t.Position, t.AmountCents,
		); err != nil {
			return fmt.Errorf("insert member fee tier: %w", err)
		}
	}

	return tx.Commit(ctx)
}
