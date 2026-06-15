package models

import "testing"

func TestRuntimeLimitsValidateVersionRuntime(t *testing.T) {
	caps := DefaultRuntimeLimits()

	if err := caps.ValidateVersionRuntime(30000, 128, 50); err != nil {
		t.Fatalf("expected valid limits, got %v", err)
	}

	if err := caps.ValidateVersionRuntime(caps.MaxTimeoutMs+1, 128, 50); err == nil {
		t.Fatal("expected timeout cap error")
	}

	if err := caps.ValidateVersionRuntime(30000, caps.MaxMemoryMB+1, 50); err == nil {
		t.Fatal("expected memory cap error")
	}

	if err := caps.ValidateVersionRuntime(30000, 128, caps.MaxCPUPercent+1); err == nil {
		t.Fatal("expected cpu cap error")
	}
}

func TestRuntimeLimitsNormalize(t *testing.T) {
	zero := RuntimeLimits{}
	norm := zero.Normalize()
	if norm.MaxTimeoutMs != DefaultRuntimeLimits().MaxTimeoutMs {
		t.Fatalf("expected default timeout, got %d", norm.MaxTimeoutMs)
	}
}
