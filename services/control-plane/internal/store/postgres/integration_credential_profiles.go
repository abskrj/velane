package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abskrj/velane/services/control-plane/internal/models"
)

type credentialProfileConfig struct {
	CredentialsType        string            `json:"credentials_type"`
	OAuthClientIDEncrypted string            `json:"oauth_client_id_encrypted,omitempty"`
	OAuthClientSecretEnc   string            `json:"oauth_client_secret_encrypted,omitempty"`
	OAuthScopes            string            `json:"oauth_scopes,omitempty"`
	EncryptedFields        map[string]string `json:"encrypted_fields,omitempty"`
}

var keySanitizer = regexp.MustCompile(`[^a-z0-9_-]+`)

func sanitizeKeyPart(v string) string {
	low := strings.ToLower(strings.TrimSpace(v))
	if low == "" {
		return "default"
	}
	clean := keySanitizer.ReplaceAllString(low, "_")
	clean = strings.Trim(clean, "_")
	if clean == "" {
		return "default"
	}
	return clean
}

func makeProviderConfigKey(provider, alias string) string {
	return fmt.Sprintf("velane_%s_%s_%d", sanitizeKeyPart(provider), sanitizeKeyPart(alias), time.Now().UnixNano())
}

func (s *Store) ListIntegrationCredentialProfiles(ctx context.Context, tenantID, provider string) ([]*models.IntegrationCredentialProfileView, error) {
	query := `SELECT id, tenant_id, provider, alias, name, nango_provider_config_key, config_json, is_default, created_by, created_at, updated_at
		  FROM integration_credential_profiles
		  WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []any{tenantID}
	if provider != "" {
		query += ` AND provider = $2`
		args = append(args, provider)
	}
	query += ` ORDER BY provider ASC, created_at ASC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListIntegrationCredentialProfiles query: %w", err)
	}
	defer rows.Close()

	var out []*models.IntegrationCredentialProfileView
	for rows.Next() {
		profile, err := scanCredentialProfile(rows)
		if err != nil {
			return nil, fmt.Errorf("ListIntegrationCredentialProfiles scan: %w", err)
		}
		view, err := toCredentialProfileView(profile)
		if err != nil {
			return nil, fmt.Errorf("ListIntegrationCredentialProfiles view: %w", err)
		}
		out = append(out, view)
	}
	return out, rows.Err()
}

func (s *Store) GetIntegrationCredentialProfileByID(ctx context.Context, tenantID, id string) (*models.IntegrationCredentialProfile, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, provider, alias, name, nango_provider_config_key, config_json, is_default, created_by, created_at, updated_at
		 FROM integration_credential_profiles
		 WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`,
		tenantID, id,
	)
	profile, err := scanCredentialProfile(row)
	if err != nil {
		return nil, fmt.Errorf("GetIntegrationCredentialProfileByID: %w", err)
	}
	return profile, nil
}

func (s *Store) GetIntegrationCredentialProfileByAlias(ctx context.Context, tenantID, provider, alias string) (*models.IntegrationCredentialProfile, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, provider, alias, name, nango_provider_config_key, config_json, is_default, created_by, created_at, updated_at
		 FROM integration_credential_profiles
		 WHERE tenant_id = $1 AND provider = $2 AND alias = $3 AND deleted_at IS NULL`,
		tenantID, provider, alias,
	)
	profile, err := scanCredentialProfile(row)
	if err != nil {
		return nil, fmt.Errorf("GetIntegrationCredentialProfileByAlias: %w", err)
	}
	return profile, nil
}

func (s *Store) GetDefaultIntegrationCredentialProfile(ctx context.Context, tenantID, provider string) (*models.IntegrationCredentialProfile, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, provider, alias, name, nango_provider_config_key, config_json, is_default, created_by, created_at, updated_at
		 FROM integration_credential_profiles
		 WHERE tenant_id = $1 AND provider = $2 AND deleted_at IS NULL
		 ORDER BY is_default DESC, created_at ASC
		 LIMIT 1`,
		tenantID, provider,
	)
	profile, err := scanCredentialProfile(row)
	if err != nil {
		return nil, fmt.Errorf("GetDefaultIntegrationCredentialProfile: %w", err)
	}
	return profile, nil
}

func (s *Store) UpsertIntegrationCredentialProfile(
	ctx context.Context,
	tenantID, provider, alias, name, createdBy string,
	credentialsType string,
	plainFields map[string]string,
	oauthScopes string,
	isDefault bool,
	providerConfigKey string,
	encKey []byte,
) (*models.IntegrationCredentialProfile, error) {
	if alias == "" {
		alias = "default"
	}
	if name == "" {
		name = alias
	}
	if providerConfigKey == "" {
		providerConfigKey = makeProviderConfigKey(provider, alias)
	}

	config, err := encryptCredentialProfileConfig(credentialsType, plainFields, oauthScopes, encKey)
	if err != nil {
		return nil, err
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal credential profile config: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var hasDefault bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM integration_credential_profiles
			WHERE tenant_id = $1 AND provider = $2 AND is_default = TRUE AND deleted_at IS NULL
		)`,
		tenantID, provider,
	).Scan(&hasDefault); err != nil {
		return nil, fmt.Errorf("check default profile: %w", err)
	}

	setDefault := isDefault || !hasDefault
	now := time.Now()
	row := tx.QueryRow(ctx,
		`INSERT INTO integration_credential_profiles
			(tenant_id, provider, alias, name, nango_provider_config_key, config_json, is_default, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		 ON CONFLICT (tenant_id, provider, alias) WHERE deleted_at IS NULL
		 DO UPDATE SET
			name = EXCLUDED.name,
			config_json = EXCLUDED.config_json,
			is_default = CASE WHEN $10 THEN TRUE ELSE integration_credential_profiles.is_default END,
			updated_at = EXCLUDED.updated_at
		 RETURNING id, tenant_id, provider, alias, name, nango_provider_config_key, config_json, is_default, created_by, created_at, updated_at`,
		tenantID, provider, alias, name, providerConfigKey, configBytes, setDefault, createdBy, now, setDefault,
	)

	profile, err := scanCredentialProfile(row)
	if err != nil {
		return nil, fmt.Errorf("upsert integration credential profile: %w", err)
	}

	if profile.IsDefault {
		if _, err := tx.Exec(ctx,
			`UPDATE integration_credential_profiles
			 SET is_default = FALSE, updated_at = NOW()
			 WHERE tenant_id = $1 AND provider = $2 AND id <> $3 AND deleted_at IS NULL`,
			tenantID, provider, profile.ID,
		); err != nil {
			return nil, fmt.Errorf("clear previous defaults: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return profile, nil
}

func (s *Store) SoftDeleteIntegrationCredentialProfile(ctx context.Context, tenantID, id string) (*models.IntegrationCredentialProfile, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	row := tx.QueryRow(ctx,
		`UPDATE integration_credential_profiles
		 SET deleted_at = NOW(), is_default = FALSE, updated_at = NOW()
		 WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
		 RETURNING id, tenant_id, provider, alias, name, nango_provider_config_key, config_json, is_default, created_by, created_at, updated_at`,
		tenantID, id,
	)
	profile, err := scanCredentialProfile(row)
	if err != nil {
		return nil, fmt.Errorf("SoftDeleteIntegrationCredentialProfile: %w", err)
	}

	var nextDefaultID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM integration_credential_profiles
		 WHERE tenant_id = $1 AND provider = $2 AND deleted_at IS NULL
		 ORDER BY created_at ASC
		 LIMIT 1`,
		tenantID, profile.Provider,
	).Scan(&nextDefaultID)
	if err == nil && nextDefaultID != "" {
		if _, err := tx.Exec(ctx,
			`UPDATE integration_credential_profiles SET is_default = TRUE, updated_at = NOW() WHERE id = $1`,
			nextDefaultID,
		); err != nil {
			return nil, fmt.Errorf("set next default profile: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return profile, nil
}

func encryptCredentialProfileConfig(credentialsType string, plainFields map[string]string, oauthScopes string, encKey []byte) (*credentialProfileConfig, error) {
	cfg := &credentialProfileConfig{
		CredentialsType: strings.ToUpper(strings.TrimSpace(credentialsType)),
		OAuthScopes:     oauthScopes,
		EncryptedFields: map[string]string{},
	}
	if cfg.CredentialsType == "" {
		cfg.CredentialsType = "OAUTH2"
	}

	for key, value := range plainFields {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		enc, err := encrypt(encKey, value)
		if err != nil {
			return nil, fmt.Errorf("encrypt field %s: %w", key, err)
		}
		cfg.EncryptedFields[key] = enc
	}

	if v, ok := cfg.EncryptedFields["oauth_client_id"]; ok {
		cfg.OAuthClientIDEncrypted = v
	}
	if v, ok := cfg.EncryptedFields["oauth_client_secret"]; ok {
		cfg.OAuthClientSecretEnc = v
	}
	return cfg, nil
}

func decryptCredentialProfileConfig(configJSON []byte, encKey []byte) (*credentialProfileConfig, map[string]string, error) {
	var cfg credentialProfileConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, nil, fmt.Errorf("decode config json: %w", err)
	}

	plain := make(map[string]string, len(cfg.EncryptedFields))
	for key, val := range cfg.EncryptedFields {
		if strings.TrimSpace(val) == "" {
			continue
		}
		dec, err := decrypt(encKey, val)
		if err != nil {
			return nil, nil, fmt.Errorf("decrypt field %s: %w", key, err)
		}
		plain[key] = dec
	}
	return &cfg, plain, nil
}

func toCredentialProfileView(profile *models.IntegrationCredentialProfile) (*models.IntegrationCredentialProfileView, error) {
	var cfg credentialProfileConfig
	if err := json.Unmarshal(profile.ConfigJSON, &cfg); err != nil {
		return nil, fmt.Errorf("decode profile config: %w", err)
	}
	return &models.IntegrationCredentialProfileView{
		ID:                     profile.ID,
		TenantID:               profile.TenantID,
		Provider:               profile.Provider,
		Alias:                  profile.Alias,
		Name:                   profile.Name,
		NangoProviderConfigKey: profile.NangoProviderConfigKey,
		CredentialsType:        cfg.CredentialsType,
		OAuthScopes:            cfg.OAuthScopes,
		IsDefault:              profile.IsDefault,
		CreatedAt:              profile.CreatedAt,
		UpdatedAt:              profile.UpdatedAt,
	}, nil
}

func (s *Store) DecryptIntegrationCredentialProfile(profile *models.IntegrationCredentialProfile, encKey []byte) (*models.IntegrationCredentialProfileView, map[string]string, error) {
	cfg, fields, err := decryptCredentialProfileConfig(profile.ConfigJSON, encKey)
	if err != nil {
		return nil, nil, err
	}
	view := &models.IntegrationCredentialProfileView{
		ID:                     profile.ID,
		TenantID:               profile.TenantID,
		Provider:               profile.Provider,
		Alias:                  profile.Alias,
		Name:                   profile.Name,
		NangoProviderConfigKey: profile.NangoProviderConfigKey,
		CredentialsType:        cfg.CredentialsType,
		OAuthScopes:            cfg.OAuthScopes,
		IsDefault:              profile.IsDefault,
		CreatedAt:              profile.CreatedAt,
		UpdatedAt:              profile.UpdatedAt,
	}
	return view, fields, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCredentialProfile(row scanner) (*models.IntegrationCredentialProfile, error) {
	var profile models.IntegrationCredentialProfile
	if err := row.Scan(
		&profile.ID,
		&profile.TenantID,
		&profile.Provider,
		&profile.Alias,
		&profile.Name,
		&profile.NangoProviderConfigKey,
		&profile.ConfigJSON,
		&profile.IsDefault,
		&profile.CreatedBy,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &profile, nil
}
