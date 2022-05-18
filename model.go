package prettylog

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sirupsen/logrus"
)

var (
	// ansi codes https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html

	timerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	// logStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("225")).Render
	// progressMsgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("158")).Render
	// failedMsgStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render
	// successMsgStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render
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
	Details   string
	Completed CompletedMessage
}

type CompletedMessage struct {
	Details string
	Success bool
}

type model struct {
	startTime    time.Time
	duration     time.Duration
	cfg          Config
	spinner      spinner.Model
	progress     string
	messages     *buffer
	logLines     *buffer
	messageRows  int
	logRows      int
	screenWidth  int
	screenHeight int
	logCh        chan LogMsg
	progressCh   chan ProgressMsg
}

func newModel(cfg Config) model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("206"))
	sp.Spinner = spinner.Dot

	m := model{
		cfg:          cfg,
		spinner:      sp,
		messages:     newBuffer(cfg.MaxMessageRows),
		logLines:     newBuffer(cfg.MaxLogRows),
		startTime:    time.Now().UTC(),
		screenWidth:  80,
		screenHeight: 25,
		logCh:        make(chan LogMsg),
		progressCh:   make(chan ProgressMsg),
	}

	m.messageRows, m.logRows = m.adjustMessageAndLogSizes()

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.timerTick,
		spinner.Tick,
		m.fetchLogMsg,
		m.fetchProgressMsg,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// always update duration
	m.duration = time.Since(m.startTime)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		m.messageRows, m.logRows = m.adjustMessageAndLogSizes()
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case TimerTickMsg:
		//m.duration = time.Since(m.startTime)
		return m, m.timerTick
	case LogMsg:
		if msg.Details != "" {
			m.logLines.Add(msg)
		}
		return m, m.fetchLogMsg
	case ProgressMsg:
		m.progress = msg.Details
		if msg.Completed.Details != "" {
			m.messages.Add(msg.Completed)
		}
		return m, m.fetchProgressMsg
	}

	return m, nil
}

func (m model) View() string {
	sb := strings.Builder{}
	sb.WriteString("\n")

	skipCount := m.messages.Len() - m.messageRows
	m.messages.Iterate(func(res tea.Msg) {
		if skipCount > 0 {
			skipCount--
			return
		}

		msg, ok := res.(CompletedMessage)
		if !ok {
			return // shouldnt happen
		}
		p := wordwrap.String(msg.Details, m.screenWidth-5)
		if msg.Success {
			p = m.successMsgStyle()(p)
		} else {
			p = m.failedMsgStyle()(p)
		}
		sb.WriteString(p)
		sb.WriteString("\n")
	})

	if m.progress != "" {
		p := fmt.Sprintf("%s %s", m.spinner.View(), m.progress)
		p = wordwrap.String(p, m.screenWidth-5)
		sb.WriteString(m.progressMsgStyle()(p))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	logLineCount := 0
	skipCount = m.logLines.Len() - m.logRows
	//adjest the message rows back
	if m.messages.Len() < m.messageRows {
		skipCount -= (m.messageRows - m.messages.Len())
	}
	m.logLines.Iterate(func(res tea.Msg) {
		if skipCount > 0 {
			skipCount--
			return
		}

		if logLineCount >= m.cfg.MaxLogRows {
			return
		}

		msg, ok := res.(LogMsg)
		if !ok {
			return //shouldnt happen
		}

		lvl := strings.ToUpper(fmt.Sprintf("[%s]", msg.Level.String()))
		lvl = m.logLvlStyle(msg.Level)(lvl)
		l := fmt.Sprintf("%s %s", lvl, m.logMsgStyle()(msg.Details))
		l = wordwrap.String(l, m.screenWidth-5)
		sb.WriteString(l)
		sb.WriteString("\n")

		// width of -6 since we will do an indent of 1 later
		logLineCount += m.countLinesInString(l, m.screenWidth-6)
	})

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

func (m model) timerTick() tea.Msg {
	<-time.Tick(5 * time.Second)
	return TimerTickMsg(true)
}

func (m model) logLvlStyle(lvl logrus.Level) func(string) string {
	c, ok := m.cfg.LogLevelColors[lvl]
	if !ok {
		c = m.cfg.LogTextColor
	}

	return m.getStyleRender(c)
}

func (m model) logMsgStyle() func(string) string {
	return m.getStyleRender(m.cfg.LogTextColor)
}

func (m model) progressMsgStyle() func(string) string {
	return m.getStyleRender(m.cfg.ProgressColor)
}

func (m model) successMsgStyle() func(string) string {
	return m.getStyleRender(m.cfg.SuccessMessageColor)
}

func (m model) failedMsgStyle() func(string) string {
	return m.getStyleRender(m.cfg.FailedMessageColor)
}

func (m model) getStyleRender(c lipgloss.Color) func(string) string {
	return lipgloss.NewStyle().Foreground(c).Render
}

func (m model) countLinesInString(str string, screenWidth int) int {
	count := 0

	for _, s := range strings.Split(str, "\n") {
		strWidth := ansi.PrintableRuneWidth(s)
		// int division so this only captures the quotient
		count += strWidth / screenWidth
		// if there is a remainder add one more line
		if strWidth%screenWidth > 0 {
			count++
		}
	}

	return count
}

func (m model) adjustMessageAndLogSizes() (int, int) {
	// adjust the message and log size based on screen height
	// the max is set in config
	// min is set to 1
	messageSize := 1
	logSize := 1

	// ===== view ================
	// blank line
	// messages
	// progress
	// blank line
	// logs
	// blank line
	// duration
	// blank line
	// =======================

	height := m.screenHeight - 6 // everything but messages and logs
	height -= 2                  // min size for messages and logs
	// whatever is left split between log and messages in the same ratio
	// as max message rows and max log rows
	if height > 0 {
		messageSize = int(math.Ceil(float64(height) * float64(m.cfg.MaxMessageRows) / float64(m.cfg.MaxMessageRows+m.cfg.MaxLogRows)))
		if messageSize > m.cfg.MaxMessageRows {
			messageSize = m.cfg.MaxMessageRows
		}

		logSize = height - messageSize
		if logSize < 0 {
			logSize = 1
		}
		if logSize > m.cfg.MaxLogRows {
			logSize = m.cfg.MaxLogRows
		}
	}

	return messageSize, logSize
}
