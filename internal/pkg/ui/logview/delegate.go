package logview

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const tsOffset = "  "

func FilterLog(term string, targets []string) []list.Rank {
	type matchedTarget struct {
		timestamp time.Time
		list.Rank
	}
	var matchedTargets []matchedTarget
	for idx, t := range targets {
		var l Log
		if err := json.Unmarshal([]byte(t), &l); err != nil {
			continue
		}
		subStringIdx := strings.Index(strings.ToLower(l.Log), strings.ToLower(term))
		if subStringIdx == -1 {
			continue
		}
		var matchedIndexes []int
		for i := subStringIdx; i < int(math.Min(float64(len(l.Log)), float64(subStringIdx+len(term)))); i++ {
			matchedIndexes = append(matchedIndexes, i)
		}
		matchedTargets = append(matchedTargets, matchedTarget{
			timestamp: l.Timestamp,
			Rank: list.Rank{
				Index:          idx,
				MatchedIndexes: matchedIndexes,
			},
		})
	}
	sort.SliceStable(matchedTargets, func(i, j int) bool {
		return matchedTargets[i].timestamp.Before(matchedTargets[j].timestamp)
	})
	results := make([]list.Rank, len(matchedTargets))
	for idx, t := range matchedTargets {
		results[idx] = list.Rank{
			MatchedIndexes: t.MatchedIndexes,
			Index:          t.Index,
		}
	}
	return results
}

func listItems(logs []Log) []list.Item {
	items := make([]list.Item, len(logs))
	for i := range logs {
		items[i] = logs[i]
	}
	return items
}

type Log struct {
	Timestamp time.Time `json:"time"`
	Log       string    `json:"log"`
}

func (l Log) Title() string {
	return l.Timestamp.Format(time.RFC3339) + tsOffset + l.Log
}

func (l Log) Description() string {
	return ""
}

// FilterValue is the value we use when filtering against this item when
// we're filtering the list.
func (l Log) FilterValue() string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	}
	return string(b)
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
