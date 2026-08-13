package sidebar

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestModel(t *testing.T, numLines int) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)
	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config:               &cfg,
		Theme:                thm,
		Styles:               context.InitStyles(thm),
		PreviewPosition:      "right",
		MainContentHeight:    5,
		DynamicPreviewWidth:  40,
		DynamicPreviewHeight: 0,
	}

	m := NewModel()
	m.IsOpen = true
	m.UpdateProgramContext(ctx)

	lines := make([]string, numLines)
	for i := range lines {
		lines[i] = "line " + strconv.Itoa(i)
	}
	m.SetContent(strings.Join(lines, "\n"))

	return m
}

func TestScrollDown_MovesViewportDown(t *testing.T) {
	m := newTestModel(t, 50)

	require.Equal(t, 0, m.YOffset())

	m.ScrollDown(1)

	require.Equal(t, 1, m.YOffset())
}

func TestScrollUp_MovesViewportUp(t *testing.T) {
	m := newTestModel(t, 50)
	m.ScrollDown(5)
	require.Equal(t, 5, m.YOffset())

	m.ScrollUp(1)

	require.Equal(t, 4, m.YOffset())
}

func TestScrollUp_ClampsAtTop(t *testing.T) {
	m := newTestModel(t, 50)

	m.ScrollUp(1)

	require.Equal(t, 0, m.YOffset(), "scrolling up from the top should stay at the top")
}

func TestScrollDown_ClampsAtBottom(t *testing.T) {
	m := newTestModel(t, 3)

	m.ScrollDown(100)
	maxOffset := m.YOffset()

	m.ScrollDown(1)

	require.Equal(
		t, maxOffset, m.YOffset(), "scrolling down past the bottom should stay at the bottom",
	)
}
