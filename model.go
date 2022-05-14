package prettylog

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sirupsen/logrus"
)

var (
	// ansi codes https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html

	timerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	logStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("225")).Render

	progressMsgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("158")).Render
	failedMsgStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render
	successMsgStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render
)

const (
	warningIcon = "⚠️"
	failedIcon  = "✗"
	successIcon = "✓"
)

type TimerTickMsg bool

type LogMsg struct {
	Level   logrus.Level
	Details string
}

type ProgressMsg struct {
	Details string
}

type AppendMessage struct {
	Details string
	Success bool
}

type model struct {
	startTime    time.Time
	duration     time.Duration
	cfg          Config
	spinner      spinner.Model
	progress     string
	messages     []string
	logLines     *buffer
	screenWidth  int
	screenHeight int
	logCh        chan LogMsg
	progressCh   chan ProgressMsg
	messagesCh   chan AppendMessage
}

func newModel(cfg Config) model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("206"))

	return model{
		cfg:          cfg,
		spinner:      sp,
		messages:     make([]string, 0),
		logLines:     newBuffer(cfg.MaxLogRows),
		startTime:    time.Now().UTC(),
		screenWidth:  80,
		screenHeight: 25,
		logCh:        make(chan LogMsg),
		progressCh:   make(chan ProgressMsg),
		messagesCh:   make(chan AppendMessage),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.timerTick,
		spinner.Tick,
		m.fetchLogMsg,
		m.fetchProgressMsg,
		m.fetchAppendMsg,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case TimerTickMsg:
		m.duration = time.Since(m.startTime)
		return m, m.timerTick
	case LogMsg:
		if msg.Details != "" {
			m.logLines.Add(msg.Details)
		}
		return m, m.fetchLogMsg
	case ProgressMsg:
		if msg.Details != "" {
			m.progress = msg.Details
		}
		return m, m.fetchProgressMsg
	case AppendMessage:
		if msg.Details != "" {
			m.messages = append(m.messages, msg.Details)
		}
		m.progress = ""
		return m, m.fetchAppendMsg
	}

	return m, nil
}

func (m model) View() string {
	sb := strings.Builder{}
	sb.WriteString("\n")

	for _, res := range m.messages {
		p := wordwrap.String(res, m.screenWidth-5)
		sb.WriteString(successMsgStyle(p))
		sb.WriteString("\n")
	}

	if m.progress != "" {
		p := fmt.Sprintf("%s %s", m.spinner.View(), m.progress)
		p = wordwrap.String(p, m.screenWidth-5)
		sb.WriteString(progressMsgStyle(p))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	for _, l := range m.logLines.Lines() {
		l = wordwrap.String(l, m.screenWidth-5)
		sb.WriteString(logStyle(l))
		sb.WriteString("\n")
	}

	sb.WriteString(timerStyle(fmt.Sprintf("\nDuration: %s\n", m.duration.Truncate(time.Second))))

	sb.WriteString("\n")
	return indent.String(sb.String(), 1)
}

func (m model) fetchLogMsg() tea.Msg {
	msg := <-m.logCh
	return msg
}

func (m model) fetchProgressMsg() tea.Msg {
	msg := <-m.progressCh
	return msg
}

func (m model) fetchAppendMsg() tea.Msg {
	msg := <-m.messagesCh
	return msg
}

func (m model) timerTick() tea.Msg {
	<-time.Tick(2 * time.Second)
	return TimerTickMsg(true)
}
