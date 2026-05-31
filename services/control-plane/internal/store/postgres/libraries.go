package postgres

import "context"

// GetTenantLibrariesForInvocation returns an empty map.
// Tenant-level custom libraries were removed in migration 012.
// Platform libraries (including @velane/integrations) are still loaded
// from the embedded binary by the scheduler's getLibraries method.
func (s *Store) GetTenantLibrariesForInvocation(_ context.Context, _, _, _ string) (map[string]string, error) {
	return map[string]string{}, nil
}
