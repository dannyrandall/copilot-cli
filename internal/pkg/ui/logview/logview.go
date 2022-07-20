package logview

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focus int

const (
	focusLogs = iota
	focusQuery
)

var spinners = []spinner.Spinner{
	spinner.Line,
	spinner.Dot,
	spinner.MiniDot,
	spinner.Jump,
	spinner.Pulse,
	spinner.Points,
	spinner.Globe,
	spinner.Moon,
	spinner.Monkey,
}

type Model struct {
	keymap   keymap
	help     help.Model
	viewport viewport.Model
	list     list.Model
	focus    focus
	count    int

	query        textinput.Model
	queryLoading bool
	spinner      spinner.Model
	queryFunc    func(string) func() tea.Msg

	Logs chan Log
	Errs chan error
	done chan struct{}

	Follow bool
}

var (
	focusedBorder        = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	blurredBorder        = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
	queryFocusedStyle    = focusedBorder.Copy()
	queryBlurredStyle    = blurredBorder.Copy().Foreground(lipgloss.Color("240"))
	viewportFocusedStyle = focusedBorder.Copy().PaddingLeft(2).PaddingRight(2)
	viewportBlurredStyle = blurredBorder.Copy().PaddingLeft(2).PaddingRight(2)
)

func NewStreamer() (Model, chan struct{}) {
	delegate := logDelegate{}

	m := Model{
		help: help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("q"),
				key.WithHelp("q", "quit"),
			),
		},
		query:   textinput.New(),
		list:    list.New(nil, delegate, 0, 0),
		spinner: spinner.New(),
		done:    make(chan struct{}),
	}

	m.list.SetShowTitle(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowPagination(false)
	m.list.SetShowHelp(false)

	m.spinner.Spinner = randSpinner()

	m.query.Prompt = "CloudFormation Query: "
	m.query.Blur()
	return m, m.done
}

func (m Model) getNextLog() tea.Cmd {
	return func() tea.Msg {
		select {
		case l := <-m.Logs:
			return l
		case err := <-m.Errs:
			return err
		case <-m.done:
			return nil
		}
	}
}

func New(logs []Log, query func(string) []Log) Model {
	delegate := logDelegate{}

	m := Model{
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
		queryFunc: func(q string) func() tea.Msg {
			return func() tea.Msg {
				return query(q)
			}
		},
		query:   textinput.New(),
		list:    list.New(listItems(logs), delegate, 0, 0),
		spinner: spinner.New(),
	}

	m.list.SetShowTitle(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowPagination(false)
	m.list.SetShowHelp(false)

	m.spinner.Spinner = randSpinner()

	m.query.Prompt = "CloudFormation Query: "
	m.query.Blur()
	return m
}

func (m Model) Init() tea.Cmd {
	return m.getNextLog()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.focus {
		case focusLogs:
			switch {
			case key.Matches(msg, m.keymap.focusToggle) && !m.queryLoading:
				cmds = append(cmds, m.query.Focus())
				m.focus = focusQuery
			case key.Matches(msg, m.keymap.quit):
				if m.Follow {
					close(m.done)
				}
				return m, tea.Quit
			default:
				if !m.Follow {
					m.list, cmd = m.list.Update(msg)
					cmds = append(cmds, cmd)
				}

			}
		case focusQuery:
			switch {
			case key.Matches(msg, m.keymap.focusToggle):
				m.query.Blur()
				m.focus = focusLogs
			case key.Matches(msg, m.keymap.runQuery):
				m.queryLoading = true
				cmds = append(cmds, m.queryFunc(m.query.Value()))
				cmds = append(cmds, m.spinner.Tick)
				m.query.Blur()
				m.focus = focusLogs
			default:
				m.query, cmd = m.query.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	case Log:
		cmds = append(cmds, m.list.InsertItem(len(m.list.Items()), msg))
		cmds = append(cmds, m.getNextLog())
	case error:
		if m.Follow {
			close(m.done)
		}
		return m, tea.Quit
	case []Log:
		m.queryLoading = false
		m.spinner.Spinner = randSpinner()
		cmds = append(cmds, m.list.SetItems(listItems(msg)))
	case spinner.TickMsg:
		if m.queryLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		m.list.SetHeight(msg.Height - lipgloss.Height(m.queryView()+"\n") - lipgloss.Height(m.helpView()+"\n") - lipgloss.Height(m.list.FilterInput.View()))
		m.list.SetWidth(msg.Width)
		m.query.Width = msg.Width
	}

	if m.Follow {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var view strings.Builder
	view.WriteString(m.queryView())
	view.WriteRune('\n')
	view.WriteString(m.logView())
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

func (m Model) logView() string {
	var logView string
	if m.focus == focusQuery {
		logView = viewportBlurredStyle.Render(m.list.View())
	} else {
		logView = viewportFocusedStyle.Render(m.list.View())
	}

	if m.queryLoading {
		return viewportFocusedStyle.Render(fmt.Sprintf("%s Loading query...", m.spinner.View()))
	}
	return logView
}

func randSpinner() spinner.Spinner {
	idx := rand.Intn(len(spinners))
	return spinners[idx]
}
