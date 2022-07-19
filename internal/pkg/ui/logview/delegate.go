package logview

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const tsOffset = "  "

type QueryResult []Log

func (q QueryResult) listItems() []list.Item {
	items := make([]list.Item, len(q))
	for i := range q {
		items[i] = q[i]
	}
	return items
}

type Log struct {
	Timestamp time.Time
	Log       string
}

func (l Log) FilterValue() string {
	return l.Timestamp.Format(time.RFC3339) + " " + l.Log
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
	log, ok := listItem.(Log)
	if !ok {
		return
	}

	fmt.Fprint(w, log.Timestamp.Format(time.RFC3339)+tsOffset+log.Log)
}
