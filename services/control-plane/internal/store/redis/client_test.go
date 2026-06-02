package redisstore

import "testing"

func TestOptionsForAddr(t *testing.T) {
	opts, err := optionsFor("localhost:6379")
	if err != nil {
		t.Fatalf("optionsFor returned error: %v", err)
	}

	if opts.Addr != "localhost:6379" {
		t.Fatalf("Addr = %q; want %q", opts.Addr, "localhost:6379")
	}
	if opts.Password != "" {
		t.Fatalf("Password = %q; want empty", opts.Password)
	}
	if opts.TLSConfig != nil {
		t.Fatal("TLSConfig should be nil for plain host:port input")
	}
}

func TestOptionsForRedisURL(t *testing.T) {
	opts, err := optionsFor("rediss://default:secret@example.com:6380/0")
	if err != nil {
		t.Fatalf("optionsFor returned error: %v", err)
	}

	if opts.Addr != "example.com:6380" {
		t.Fatalf("Addr = %q; want %q", opts.Addr, "example.com:6380")
	}
	if opts.Username != "default" {
		t.Fatalf("Username = %q; want %q", opts.Username, "default")
	}
	if opts.Password != "secret" {
		t.Fatalf("Password = %q; want %q", opts.Password, "secret")
	}
	if opts.DB != 0 {
		t.Fatalf("DB = %d; want 0", opts.DB)
	}
	if opts.TLSConfig == nil {
		t.Fatal("TLSConfig should be set for rediss URLs")
	}
}
