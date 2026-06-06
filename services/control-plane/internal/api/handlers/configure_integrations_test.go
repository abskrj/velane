package handlers

import "testing"

func TestFormatScopesForNango(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "single google scope",
			raw:  "https://www.googleapis.com/auth/spreadsheets",
			want: "https://www.googleapis.com/auth/spreadsheets",
		},
		{
			name: "comma separated",
			raw:  "scope-a,scope-b",
			want: "scope-a,scope-b",
		},
		{
			name: "space separated",
			raw:  "scope-a scope-b",
			want: "scope-a,scope-b",
		},
		{
			name: "dedupe",
			raw:  "scope-a, scope-a, scope-b",
			want: "scope-a,scope-b",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatScopesForNango(tc.raw); got != tc.want {
				t.Fatalf("formatScopesForNango(%q) = %q; want %q", tc.raw, got, tc.want)
			}
		})
	}
}
