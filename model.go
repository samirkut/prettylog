package prettylog

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
)

var (
	// ansi codes https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	logStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("225")).Render

	progressMsgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("158")).Render
	failedMsgStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render
	successMsgStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render
)

const (
	warningIcon  = "⚠️"
	failedIcon   = "✗"
	successIcon  = "✓"
	progressIcon = "🏃"
)

type LogLine string
type Message string

type model struct {
	cfg      Config
	spinner  spinner.Model
	messages []string
	logLines *buffer
}

func newModel(cfg Config) model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("206"))

	return model{
		cfg:      cfg,
		spinner:  sp,
		messages: make([]string, 0),
		logLines: newBuffer(cfg.MaxLogRows),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		fetchMessages,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case processFinishedMsg:
		d := time.Duration(msg)
		res := result{emoji: randomEmoji(), duration: d}
		log.Printf("%s Job finished in %s", res.emoji, res.duration)
		m.results = append(m.results[1:], res)
		return m, runPretendProcess
	default:
		return m, nil
	}
}

func (m model) View() string {
	s := "\n" +
		m.spinner.View() + " Doing some work...\n\n"

	for _, res := range m.results {
		if res.duration == 0 {
			s += "........................\n"
		} else {
			s += fmt.Sprintf("%s Job finished in %s\n", res.emoji, res.duration)
		}
	}

	s += helpStyle("\nPress any key to exit\n")

	if m.quitting {
		s += "\n"
	}

	return indent.String(s, 1)
}

func fetchMessages() tea.Msg {

}
