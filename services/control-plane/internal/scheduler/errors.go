package scheduler

import "errors"

// ErrAccessDenied is returned when the authenticated tenant cannot access a workflow.
var ErrAccessDenied = errors.New("access denied")
