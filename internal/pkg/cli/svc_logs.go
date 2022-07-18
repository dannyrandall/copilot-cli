// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/copilot-cli/internal/pkg/aws/identity"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/deploy"
	"github.com/aws/copilot-cli/internal/pkg/logging"
	"github.com/aws/copilot-cli/internal/pkg/term/log"
	"github.com/aws/copilot-cli/internal/pkg/term/prompt"
	"github.com/aws/copilot-cli/internal/pkg/term/selector"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const (
	svcLogNamePrompt     = "Which service's logs would you like to show?"
	svcLogNameHelpPrompt = "The logs of the indicated deployed service will be shown."

	cwGetLogEventsLimitMin = 1
	cwGetLogEventsLimitMax = 10000
)

type wkldLogsVars struct {
	shouldOutputJSON bool
	follow           bool
	limit            int
	name             string
	envName          string
	appName          string
	humanStartTime   string
	humanEndTime     string
	taskIDs          []string
	since            time.Duration
	logGroup         string
}

type svcLogsOpts struct {
	wkldLogsVars
	wkldLogOpts
	// cached variables.
	targetEnv *config.Environment
}

type wkldLogOpts struct {
	// internal states
	startTime *int64
	endTime   *int64

	w           io.Writer
	configStore store
	deployStore deployedEnvironmentLister
	sel         deploySelector
	logsSvc     logEventsWriter
	initLogsSvc func() error // Overridden in tests.
}

func newSvcLogOpts(vars wkldLogsVars) (*svcLogsOpts, error) {
	sessProvider := sessions.ImmutableProvider(sessions.UserAgentExtras("svc logs"))
	defaultSess, err := sessProvider.Default()
	if err != nil {
		return nil, fmt.Errorf("default session: %v", err)
	}

	configStore := config.NewSSMStore(identity.New(defaultSess), ssm.New(defaultSess), aws.StringValue(defaultSess.Config.Region))
	deployStore, err := deploy.NewStore(sessProvider, configStore)
	if err != nil {
		return nil, fmt.Errorf("connect to deploy store: %w", err)
	}
	opts := &svcLogsOpts{
		wkldLogsVars: vars,
		wkldLogOpts: wkldLogOpts{
			w:           log.OutputWriter,
			configStore: configStore,
			deployStore: deployStore,
			sel:         selector.NewDeploySelect(prompt.New(), configStore, deployStore),
		},
	}
	opts.initLogsSvc = func() error {
		env, err := opts.getTargetEnv()
		if err != nil {
			return fmt.Errorf("get environment: %w", err)
		}
		workload, err := configStore.GetWorkload(opts.appName, opts.name)
		if err != nil {
			return fmt.Errorf("get workload: %w", err)
		}
		sess, err := sessProvider.FromRole(env.ManagerRoleARN, env.Region)
		if err != nil {
			return err
		}
		opts.logsSvc, err = logging.NewWorkloadClient(&logging.NewWorkloadLogsConfig{
			App:         opts.appName,
			Env:         opts.envName,
			Name:        opts.name,
			Sess:        sess,
			LogGroup:    opts.logGroup,
			WkldType:    workload.Type,
			TaskIDs:     opts.taskIDs,
			ConfigStore: configStore,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return opts, nil
}

// Validate returns an error for any invalid optional flags.
func (o *svcLogsOpts) Validate() error {
	return nil
}

// Ask prompts for and validates any required flags.
func (o *svcLogsOpts) Ask() error {
	return nil
}

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()

	logsViewportStyle = func() lipgloss.Style {
		b := lipgloss.DoubleBorder()
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}
)

type keymap struct {
	filter, scrollUp, scrollDown, search key.Binding
}

type model struct {
	svc          string
	filter       textarea.Model
	help         help.Model
	logsViewport viewport.Model
	logs         []string
	ready        bool
	keymap       keymap

	focus int
}

func newModel() model {
	m := model{
		logs: []string{
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
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
		help:   help.New(),
		filter: textarea.New(),
		keymap: keymap{
			filter: key.NewBinding(
				key.WithKeys("/"),
				key.WithHelp("/", "filter"),
			),
			search: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "search"),
			),
			scrollUp: key.NewBinding(
				key.WithKeys("k"),
				key.WithHelp("k", "scroll up"),
			),
			scrollDown: key.NewBinding(
				key.WithKeys("j"),
				key.WithHelp("j", "scroll down"),
			),
		},
	}

	m.filter.SetHeight(1)
	m.filter.Prompt = "filter"
	m.filter.Blur()
	m.filter.KeyMap.InsertNewline.SetEnabled(false)
	return m
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.logsViewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.logsViewport.YPosition = headerHeight
			m.logsViewport.HighPerformanceRendering = true
			m.logsViewport.SetContent(strings.Join(m.logs, "\n"))
			m.ready = true
		} else {
			m.logsViewport.Width = msg.Width
			m.logsViewport.Height = msg.Height - verticalMarginHeight
		}

		cmds = append(cmds, viewport.Sync(m.logsViewport))
	}

	m.logsViewport, cmd = m.logsViewport.Update(msg)
	cmds = append(cmds, cmd)

	m.filter, cmd = m.filter.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var view strings.Builder
	view.WriteString(m.headerView())
	view.WriteRune('\n')
	view.WriteString(m.logsViewport.View())
	view.WriteRune('\n')
	view.WriteString(m.footerView())

	return view.String()
}

func (m model) headerView() string {
	title := titleStyle.Render("Logs for Service: iss-3714")
	line := strings.Repeat("─", max(0, m.logsViewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	return m.help.ShortHelpView([]key.Binding{
		m.keymap.filter,
		m.keymap.scrollDown,
		m.keymap.scrollUp,
	})
	/*
		info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.logsViewport.ScrollPercent()*100))
		line := strings.Repeat("─", max(0, m.logsViewport.Width-lipgloss.Width(info)))
		return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
	*/
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Execute outputs logs of the service.
func (o *svcLogsOpts) Execute() error {
	model := newModel()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	return nil
}

func (o *svcLogsOpts) validateOrAskApp() error {
	if o.appName != "" {
		_, err := o.configStore.GetApplication(o.appName)
		return err
	}
	app, err := o.sel.Application(svcAppNamePrompt, svcAppNameHelpPrompt)
	if err != nil {
		return fmt.Errorf("select application: %w", err)
	}
	o.appName = app
	return nil
}

func (o *svcLogsOpts) validateAndAskSvcEnvName() error {
	if o.envName != "" {
		if _, err := o.getTargetEnv(); err != nil {
			return err
		}
	}

	if o.name != "" {
		if _, err := o.configStore.GetService(o.appName, o.name); err != nil {
			return err
		}
	}
	// Note: we let prompter handle the case when there is only option for user to choose from.
	// This is naturally the case when `o.envName != "" && o.name != ""`.
	deployedService, err := o.sel.DeployedService(svcLogNamePrompt, svcLogNameHelpPrompt, o.appName, selector.WithEnv(o.envName), selector.WithName(o.name))
	if err != nil {
		return fmt.Errorf("select deployed services for application %s: %w", o.appName, err)
	}
	o.name = deployedService.Name
	o.envName = deployedService.Env
	return nil
}

func (o *svcLogsOpts) getTargetEnv() (*config.Environment, error) {
	if o.targetEnv != nil {
		return o.targetEnv, nil
	}
	env, err := o.configStore.GetEnvironment(o.appName, o.envName)
	if err != nil {
		return nil, err
	}
	o.targetEnv = env
	return o.targetEnv, nil
}

func parseSince(since time.Duration) *int64 {
	sinceSec := int64(since.Round(time.Second).Seconds())
	timeNow := time.Now().Add(time.Duration(-sinceSec) * time.Second)
	return aws.Int64(timeNow.UnixMilli())
}

func parseRFC3339(timeStr string) (int64, error) {
	startTimeTmp, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return 0, fmt.Errorf("reading time value %s: %w", timeStr, err)
	}
	return startTimeTmp.UnixMilli(), nil
}

// buildSvcLogsCmd builds the command for displaying service logs in an application.
func buildSvcLogsCmd() *cobra.Command {
	vars := wkldLogsVars{}
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Displays logs of a deployed service.",

		Example: `
  Displays logs of the service "my-svc" in environment "test".
  /code $ copilot svc logs -n my-svc -e test
  Displays logs in the last hour.
  /code $ copilot svc logs --since 1h
  Displays logs from 2006-01-02T15:04:05 to 2006-01-02T15:05:05.
  /code $ copilot svc logs --start-time 2006-01-02T15:04:05+00:00 --end-time 2006-01-02T15:05:05+00:00
  Displays logs from specific task IDs.
  /code $ copilot svc logs --tasks 709c7eae05f947f6861b150372ddc443,1de57fd63c6a4920ac416d02add891b9
  Displays logs in real time.
  /code $ copilot svc logs --follow
  Display logs from specific log group.
  /code $ copilot svc logs --log-group system`,
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			opts, err := newSvcLogOpts(vars)
			if err != nil {
				return err
			}
			return run(opts)
		}),
	}
	cmd.Flags().StringVarP(&vars.name, nameFlag, nameFlagShort, "", svcFlagDescription)
	cmd.Flags().StringVarP(&vars.envName, envFlag, envFlagShort, "", envFlagDescription)
	cmd.Flags().StringVarP(&vars.appName, appFlag, appFlagShort, tryReadingAppName(), appFlagDescription)
	cmd.Flags().StringVar(&vars.humanStartTime, startTimeFlag, "", startTimeFlagDescription)
	cmd.Flags().StringVar(&vars.humanEndTime, endTimeFlag, "", endTimeFlagDescription)
	cmd.Flags().BoolVar(&vars.shouldOutputJSON, jsonFlag, false, jsonFlagDescription)
	cmd.Flags().BoolVar(&vars.follow, followFlag, false, followFlagDescription)
	cmd.Flags().DurationVar(&vars.since, sinceFlag, 0, sinceFlagDescription)
	cmd.Flags().IntVar(&vars.limit, limitFlag, 0, limitFlagDescription)
	cmd.Flags().StringSliceVar(&vars.taskIDs, tasksFlag, nil, tasksLogsFlagDescription)
	cmd.Flags().StringVar(&vars.logGroup, logGroupFlag, "", logGroupFlagDescription)
	return cmd
}
