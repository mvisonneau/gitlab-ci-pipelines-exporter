package ui

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	chromaQuick "github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/xeonx/timeago"
)

var (
	color   = termenv.ColorProfile().Color
	keyword = termenv.Style{}.Foreground(color("204")).Background(color("235")).Styled
	help    = termenv.Style{}.Foreground(color("241")).Styled
)

type tab string

const (
	tabStatus tab = "status"
	tabConfig tab = "config"
)

var tabs = [...]tab{
	tabStatus,
	tabConfig,
}

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	divider   = lipgloss.NewStyle().
			SetString("•").
			Padding(0, 1).
			Foreground(subtle).
			String()

	dataStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#a9a9a9"))

	// Tabs

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	inactiveTab = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

	activeTab = inactiveTab.Copy().Border(activeTabBorder, true)

	tabGap = inactiveTab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	// List

	entityStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(subtle)

	// Status Bar

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#003d80")).
			Padding(0, 1).
			MarginRight(1)

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	versionStyle = statusNugget.Copy().
			Background(lipgloss.Color("#0062cc"))

	// Page
	docStyle = lipgloss.NewStyle()
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type model struct {
	version         string
	listenerAddress *url.URL
	rpcClient       *rpc.Client
	sub             chan monitor.Status
	lastStatus      *monitor.Status
	vp              viewport.Model
	progress        *progress.Model
	tabID           int
}

func (m *model) renderLastStatus() string {
	if m.lastStatus == nil {
		return "\nloading data.."
	}

	gitlabAPIUsage := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" GitLab API usage        ",
		m.progress.ViewAs(m.lastStatus.GitLabAPIUsage),
		"\n",
	)

	gitlabAPIRequestsCount := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" GitLab API requests    ",
		dataStyle.SetString(strconv.Itoa(int(m.lastStatus.GitLabAPIRequestsCount))).String(),
		"\n",
	)

	tasksBufferUsage := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" Tasks buffer usage      ",
		m.progress.ViewAs(m.lastStatus.TasksBufferUsage),
		"\n",
	)

	tasksExecuted := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" Tasks executed         ",
		dataStyle.SetString(strconv.Itoa(int(m.lastStatus.TasksExecutedCount))).String(),
		"\n",
	)

	return strings.Join([]string{
		"",
		gitlabAPIUsage,
		gitlabAPIRequestsCount,
		tasksBufferUsage,
		tasksExecuted,
		renderEntityStatus("Projects", m.lastStatus.Projects),
		renderEntityStatus("Environments", m.lastStatus.Envs),
		renderEntityStatus("Refs", m.lastStatus.Refs),
		renderEntityStatus("Metrics", m.lastStatus.Metrics),
	}, "\n")
}

func renderEntityStatus(name string, es monitor.EntityStatus) string {
	return entityStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		" "+name+strings.Repeat(" ", 24-len(name)),
		lipgloss.JoinVertical(
			lipgloss.Left,
			"Total      "+dataStyle.SetString(strconv.Itoa(int(es.Count))).String()+"\n",
			"Last Pull  "+dataStyle.SetString(prettyTimeago(es.LastPull)).String()+"\n",
			"Last GC    "+dataStyle.SetString(prettyTimeago(es.LastGC)).String()+"\n",
			"Next Pull  "+dataStyle.SetString(prettyTimeago(es.NextPull)).String()+"\n",
			"Next GC    "+dataStyle.SetString(prettyTimeago(es.NextGC)).String()+"\n",
		),
		"\n",
	))
}

func prettyTimeago(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return timeago.English.Format(t)
}

func newModel(version string, listenerAddress *url.URL) (m *model) {
	rpcClient := rpc.NewClient(listenerAddress)
	p := progress.NewModel(progress.WithScaledGradient("#80c904", "#ff9d5c"))

	m = &model{
		version:         version,
		listenerAddress: listenerAddress,
		sub:             make(chan monitor.Status),
		vp:              viewport.Model{},
		progress:        &p,
		rpcClient:       rpcClient,
	}
	return
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.generateActivity(),
		waitForActivity(m.sub),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height - 4
		m.progress.Width = msg.Width - 27
		m.setPaneContent()
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyLeft:
			if m.tabID > 0 {
				m.tabID--
				m.setPaneContent()
			}
			return m, nil
		case tea.KeyRight:
			if m.tabID < len(tabs)-1 {
				m.tabID++
				m.setPaneContent()
			}
			return m, nil
		case tea.KeyUp, tea.KeyDown, tea.KeyPgDown, tea.KeyPgUp:
			vp, cmd := m.vp.Update(msg)
			m.vp = vp
			return m, cmd
		}
	case monitor.Status:
		m.lastStatus = &msg
		if m.tabID == 0 {
			m.vp.SetContent(m.renderLastStatus())
		}
		return m, waitForActivity(m.sub)
	}

	return m, nil
}

func (m model) View() string {
	doc := strings.Builder{}

	// TABS
	{
		renderedTabs := []string{}
		for tabID, t := range tabs {
			if m.tabID == tabID {
				renderedTabs = append(renderedTabs, activeTab.Render(string(t)))
				continue
			}
			renderedTabs = append(renderedTabs, inactiveTab.Render(string(t)))
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
		gap := tabGap.Render(strings.Repeat(" ", max(0, m.vp.Width-lipgloss.Width(row))))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
		doc.WriteString(row + "\n")
	}

	// PANE
	{
		doc.WriteString(m.vp.View() + "\n")
	}

	// Status bar
	{
		bar := lipgloss.JoinHorizontal(lipgloss.Top,
			statusStyle.Render("github.com/mvisonneau/gitlab-ci-pipelines-exporter"),
			statusText.Copy().
				Width(max(0, m.vp.Width-(55+len(m.version)))).
				Render(""),
			versionStyle.Render(m.version),
		)

		doc.WriteString(statusBarStyle.Width(m.vp.Width).Render(bar))
	}

	return docStyle.Render(doc.String())
}

func waitForActivity(sub chan monitor.Status) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func (m model) generateActivity() tea.Cmd {
	return func() tea.Msg {
		for {
			time.Sleep(time.Second)
			m.sub <- m.rpcClient.Status()
		}
	}
}

// Start ..
func Start(version string, listenerAddress *url.URL) {
	if err := tea.NewProgram(
		newModel(version, listenerAddress),
		tea.WithAltScreen(),
	).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m *model) setPaneContent() {
	switch tabs[m.tabID] {
	case tabStatus:
		m.vp.SetContent(m.renderLastStatus())
	case tabConfig:
		var b bytes.Buffer
		foo := bufio.NewWriter(&b)
		if err := chromaQuick.Highlight(foo, m.rpcClient.Config(), "yaml", "terminal16m", "monokai"); err != nil {
			log.WithError(err).Fatal()
		}

		m.vp.SetContent(b.String())
	}
}
