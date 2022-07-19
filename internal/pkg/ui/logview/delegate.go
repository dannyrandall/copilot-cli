package logview

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const tsOffset = "  "

type log struct {
	ts  time.Time
	log string
}

func (l log) FilterValue() string {
	return l.ts.Format(time.RFC3339) + " " + l.log
}

type logDelegate struct{}

func (d logDelegate) Height() int {
	return 1
}

func (d logDelegate) Spacing() int {
	return 0
}

func (d logDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d logDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	log, ok := listItem.(log)
	if !ok {
		return
	}

	ts := log.ts.Format(time.RFC3339) + tsOffset
	maxLineWidth := m.Width() - lipgloss.Width(ts)
	if lipgloss.Width(log.log) < maxLineWidth {
		fmt.Fprint(w, ts+log.log)
		return
	}

	var lines []string
	lastChop := 0
	for i := range log.log {
		if lipgloss.Width(log.log[lastChop:i]) == maxLineWidth {
			lines = append(lines, log.log[lastChop:i])
			lastChop = i
		}
	}

	lines = append(lines, log.log[lastChop:])
	whiteSpace := strings.Repeat(" ", lipgloss.Width(ts))
	//for i := 1; i < len(lines); i++ {
	//	lines[i] = whiteSpace + lines[i]
	//}

	fmt.Fprint(w, ts+strings.Join(lines, "\n"+whiteSpace))
}
