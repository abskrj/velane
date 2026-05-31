package hub

import (
	"sync"

	"github.com/abskrj/velane/services/control-plane/internal/models"
)

// Hub broadcasts SnippetVersion events to all active watchers of a snippet.
type Hub struct {
	mu   sync.RWMutex
	subs map[string][]chan *models.SnippetVersion
}

func New() *Hub {
	return &Hub{subs: make(map[string][]chan *models.SnippetVersion)}
}

// Subscribe returns a channel that receives events for snippetID.
// The returned cancel func must be called when the caller is done (prevents leaks).
func (h *Hub) Subscribe(snippetID string) (<-chan *models.SnippetVersion, func()) {
	ch := make(chan *models.SnippetVersion, 1)
	h.mu.Lock()
	h.subs[snippetID] = append(h.subs[snippetID], ch)
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.subs[snippetID]
		for i, c := range subs {
			if c == ch {
				h.subs[snippetID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}
}

// Publish sends v to all current subscribers of snippetID.
// Events are dropped for slow consumers to avoid blocking the writer.
func (h *Hub) Publish(snippetID string, v *models.SnippetVersion) {
	h.mu.RLock()
	subs := make([]chan *models.SnippetVersion, len(h.subs[snippetID]))
	copy(subs, h.subs[snippetID])
	h.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- v:
		default:
		}
	}
}
