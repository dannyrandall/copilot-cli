package logview

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keymap struct {
	focusToggle, runQuery, quit key.Binding
}

type focus int

const (
	focusViewport = iota
	focusQuery
)

type Model struct {
	keymap   keymap
	logs     []string
	help     help.Model
	viewport viewport.Model
	ready    bool
	query    textinput.Model
	focus    focus
	count    int
}

var (
	focusedBorder        = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	blurredBorder        = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
	queryFocusedStyle    = focusedBorder.Copy()
	queryBlurredStyle    = blurredBorder.Copy().Foreground(lipgloss.Color("240"))
	viewportFocusedStyle = focusedBorder.Copy().PaddingLeft(2).PaddingRight(2)
	viewportBlurredStyle = blurredBorder.Copy().PaddingLeft(2).PaddingRight(2)
)

func New() Model {
	m := Model{
		logs: []string{
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna alialiquaaliquaaliquaaliquaaliquaaliquaaliquaaliquaaliquaqua aliquaalialiquaaliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco",
			"laboris nisi ut aliquip ex ea commodo consequat.",
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum",
			"dolore eu fugiat nulla pariatur.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Excepteur sint occaecat cupidatat non proident,",
			"sunt in culpa qui officia deserunt mollit anim id est laborum.",
		},
		help: help.New(),
		keymap: keymap{
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
		},
		query: textinput.New(),
	}

	m.query.Prompt = "Cloudformation Query: "
	m.query.Blur()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.focus {
		case focusViewport:
			switch {
			case key.Matches(msg, m.keymap.focusToggle):
				cmds = append(cmds, m.query.Focus())
				m.viewport.Style = viewportBlurredStyle
				m.focus = focusQuery
			case key.Matches(msg, m.keymap.quit):
				return m, tea.Quit
			default:
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		case focusQuery:
			switch {
			case key.Matches(msg, m.keymap.focusToggle):
				m.viewport.Style = viewportFocusedStyle
				m.query.Blur()
				m.focus = focusViewport
			case key.Matches(msg, m.keymap.runQuery):
				m.query.Blur()
				m.focus = focusViewport
			default:
				m.query, cmd = m.query.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		heightUsed := lipgloss.Height(m.queryView()+"\n") + lipgloss.Height(m.helpView()+"\n")

		if !m.ready {
			m.query.Width = msg.Width - 2
			m.viewport = viewport.New(msg.Width-2, msg.Height-heightUsed-2) // -2 from height otherwise the border doesn't show...
			m.viewport.Style = viewportFocusedStyle
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
			m.ready = true
		} else {
			m.query.Width = msg.Width - 2
			m.viewport.Width = msg.Width - 2
			m.viewport.Height = msg.Height - heightUsed - 2
		}

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		m.count++
		return "\n Initializing... " + strconv.Itoa(m.count)
	}

	var view strings.Builder
	view.WriteString(m.queryView())
	view.WriteRune('\n')
	view.WriteString(m.viewport.View())
	view.WriteRune('\n')
	view.WriteString(m.helpView())
	return view.String()
}

func (m Model) queryView() string {
	if m.focus == focusQuery {
		return queryFocusedStyle.Render(m.query.View())
	}
	return queryBlurredStyle.Render(m.query.View())
}

func (m Model) helpView() string {
	if m.focus == focusQuery {
		return m.help.ShortHelpView([]key.Binding{
			m.keymap.focusToggle,
			m.keymap.runQuery,
		})
	}
	return m.help.ShortHelpView([]key.Binding{
		m.keymap.focusToggle,
		key.NewBinding(key.WithHelp("j", "scroll down")),
		key.NewBinding(key.WithHelp("k", "scroll up")),
		m.keymap.quit,
	})
}
