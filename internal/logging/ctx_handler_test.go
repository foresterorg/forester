package logging

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"golang.org/x/exp/slog"
)

func TestContext(t *testing.T) {
	sb := &strings.Builder{}
	th := slog.NewTextHandler(sb, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	logger := slog.New(NewContextHandler(th))
	logger.Info("info")

	if got, want := sb.String(), "level=INFO msg=info\n"; !cmp.Equal(got, want) {
		t.Errorf("got = %v, want = %v", got, want)
	}
}
