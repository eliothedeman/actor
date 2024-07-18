package actor

import (
	"fmt"
	"log/slog"
)

func (m *msg) slog(parent *slog.Logger) *slog.Logger {
	if parent == nil {
		parent = slog.With()
	}
	return parent.With("to", m.to, "from", m.from, "type", fmt.Sprintf("%T", m.data), "data", m.data)
}
