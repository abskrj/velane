package ids

import "github.com/google/uuid"

// New returns a new UUID v7 string for use as a primary key.
func New() string {
	return uuid.Must(uuid.NewV7()).String()
}
