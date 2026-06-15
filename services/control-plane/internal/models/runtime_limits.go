package models

import "fmt"

// RuntimeLimits defines per-tenant maximums for workflow version runtime settings.
// Tenants see these as read-only caps; platform ops adjust them in the database.
type RuntimeLimits struct {
	MaxTimeoutMs    int `json:"max_timeout_ms"`
	MaxMemoryMB     int `json:"max_memory_mb"`
	MaxCPUPercent   int `json:"max_cpu_percent"`
}

// DefaultRuntimeLimits returns platform defaults for new tenants.
func DefaultRuntimeLimits() RuntimeLimits {
	return RuntimeLimits{
		MaxTimeoutMs:  900000, // 15 minutes
		MaxMemoryMB:   2048,
		MaxCPUPercent: 100,
	}
}

// Normalize applies defaults for zero or negative stored values.
func (r RuntimeLimits) Normalize() RuntimeLimits {
	def := DefaultRuntimeLimits()
	if r.MaxTimeoutMs <= 0 {
		r.MaxTimeoutMs = def.MaxTimeoutMs
	}
	if r.MaxMemoryMB <= 0 {
		r.MaxMemoryMB = def.MaxMemoryMB
	}
	if r.MaxCPUPercent <= 0 {
		r.MaxCPUPercent = def.MaxCPUPercent
	}
	return r
}

// ValidateVersionRuntime checks per-version limits against tenant caps.
func (r RuntimeLimits) ValidateVersionRuntime(timeoutMs, maxMemoryMB, maxCPUPercent int) error {
	caps := r.Normalize()
	if timeoutMs <= 0 {
		return fmt.Errorf("timeout_ms must be positive")
	}
	if maxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be positive")
	}
	if maxCPUPercent <= 0 || maxCPUPercent > 100 {
		return fmt.Errorf("max_cpu_percent must be between 1 and 100")
	}
	if timeoutMs > caps.MaxTimeoutMs {
		return fmt.Errorf("timeout_ms exceeds tenant limit (%d)", caps.MaxTimeoutMs)
	}
	if maxMemoryMB > caps.MaxMemoryMB {
		return fmt.Errorf("max_memory_mb exceeds tenant limit (%d)", caps.MaxMemoryMB)
	}
	if maxCPUPercent > caps.MaxCPUPercent {
		return fmt.Errorf("max_cpu_percent exceeds tenant limit (%d)", caps.MaxCPUPercent)
	}
	return nil
}
