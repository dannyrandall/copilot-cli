package logview

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
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
	return l.Log
}

type logDelegate struct {
	Styles list.DefaultItemStyles
}

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
	var (
		title        string
		matchedRunes []int
		s            = &d.Styles
	)
	log, ok := listItem.(Log)
	if !ok {
		return
	}
	title = fmt.Sprint(log.Timestamp.Format(time.RFC3339) + tsOffset + log.Log)

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Prevent text from exceeding list width
	textwidth := uint(m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
	title = truncate.StringWithTail(title, textwidth, "...")
	// Conditions
	var (
		isSelected = index == m.Index()
		// emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	)

	if isFiltered {
		// Get indices of matched characters
		matchedRunes = m.MatchesForItem(index)
	}

	// Opt 2
	if isFiltered {
		// Highlight matches
		unmatched := s.SelectedTitle.Inline(true)
		matched := unmatched.Copy().Inherit(s.FilterMatch)
		title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
	}

	if isSelected {
		title = s.SelectedTitle.Render(title)
	} else {
		title = s.NormalTitle.Render(title)
	}

	// Original
	// if emptyFilter {
	// 	title = s.DimmedTitle.Render(title)
	// } else if isSelected && m.FilterState() != list.Filtering {
	// 	if isFiltered {
	// 		// Highlight matches
	// 		unmatched := s.SelectedTitle.Inline(true)
	// 		matched := unmatched.Copy().Inherit(s.FilterMatch)
	// 		title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
	// 	}
	// 	title = s.SelectedTitle.Render(title)
	// } else {
	// 	if isFiltered {
	// 		// Highlight matches
	// 		unmatched := s.NormalTitle.Inline(true)
	// 		matched := unmatched.Copy().Inherit(s.FilterMatch)
	// 		title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
	// 	}
	// 	title = s.NormalTitle.Render(title)
	// }

	fmt.Fprintf(w, "%s", title)
}
