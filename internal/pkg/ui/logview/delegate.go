package logview

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const tsOffset = "  "

func listItems(logs []Log) []list.Item {
	items := make([]list.Item, len(logs))
	for i := range logs {
		items[i] = logs[i]
	}
	return items
}

type Log struct {
	Timestamp time.Time
	Log       string
}

// FilterValue is the value we use when filtering against this item when
// we're filtering the list.
func (l Log) FilterValue() string {
	return l.Timestamp.Format(time.RFC3339) + " " + l.Log
}

type logDelegate struct{}

// Height is the height of a list item.
func (d logDelegate) Height() int {
	return 1
}

// Spacing is the size of the horizontal gap between list items in cells.
func (d logDelegate) Spacing() int {
	return 0
}

// Update is the update loop for items. All messages in the list's update
// loop will pass through here except when the user is setting a filter.
// Use this method to perform item-level updates appropriate to this
// delegate.
func (d logDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders the item's view.
func (d logDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	log, ok := listItem.(Log)
	if !ok {
		return
	}

	fmt.Fprint(w, log.Timestamp.Format(time.RFC3339)+tsOffset+log.Log)
}
