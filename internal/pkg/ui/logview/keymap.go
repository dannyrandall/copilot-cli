package logview

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	focusToggle key.Binding
	runQuery    key.Binding
	quit        key.Binding

	choose key.Binding
	remove key.Binding
}

func newKeymap() keymap {
	return keymap{
		focusToggle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		runQuery: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "run query"),
		),
		quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}

func (m Model) helpView() string {
	switch {
	case m.focus == focusQuery:
		return m.help.ShortHelpView([]key.Binding{
			m.keymap.focusToggle,
			m.keymap.runQuery,
		})
	case m.focus == focusLogs && m.queryLoading:
		return m.help.ShortHelpView([]key.Binding{
			m.keymap.quit,
		})
	default:
		return m.help.ShortHelpView([]key.Binding{
			key.NewBinding(key.WithHelp("j", "scroll down")),
			key.NewBinding(key.WithHelp("k", "scroll up")),
			key.NewBinding(key.WithHelp("h", "prev page")),
			key.NewBinding(key.WithHelp("l", "next page")),
			m.keymap.focusToggle,
			m.keymap.quit,
		})
	}
}
